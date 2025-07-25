package collection_test

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	hotstuffmodel "github.com/onflow/flow-go/consensus/hotstuff/model"
	model "github.com/onflow/flow-go/model/cluster"
	"github.com/onflow/flow-go/model/flow"
	builder "github.com/onflow/flow-go/module/builder/collection"
	"github.com/onflow/flow-go/module/mempool"
	"github.com/onflow/flow-go/module/mempool/herocache"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/module/trace"
	"github.com/onflow/flow-go/state/cluster"
	clusterkv "github.com/onflow/flow-go/state/cluster/badger"
	"github.com/onflow/flow-go/state/protocol"
	pbadger "github.com/onflow/flow-go/state/protocol/badger"
	"github.com/onflow/flow-go/state/protocol/events"
	"github.com/onflow/flow-go/state/protocol/inmem"
	"github.com/onflow/flow-go/state/protocol/protocol_state/kvstore"
	"github.com/onflow/flow-go/state/protocol/util"
	"github.com/onflow/flow-go/storage"
	bstorage "github.com/onflow/flow-go/storage/badger"
	"github.com/onflow/flow-go/storage/badger/operation"
	"github.com/onflow/flow-go/storage/badger/procedure"
	"github.com/onflow/flow-go/utils/unittest"
)

var noopSetter = func(*flow.Header) error { return nil }
var noopSigner = func(*flow.Header) error { return nil }

type BuilderSuite struct {
	suite.Suite
	db    *badger.DB
	dbdir string

	genesis      *model.Block
	chainID      flow.ChainID
	epochCounter uint64

	headers  storage.Headers
	payloads storage.ClusterPayloads
	blocks   storage.Blocks

	state cluster.MutableState

	// protocol state for reference blocks for transactions
	protoState protocol.FollowerState

	pool    mempool.Transactions
	builder *builder.Builder
}

// runs before each test runs
func (suite *BuilderSuite) SetupTest() {
	var err error

	suite.genesis = model.Genesis()
	suite.chainID = suite.genesis.Header.ChainID

	suite.pool = herocache.NewTransactions(1000, unittest.Logger(), metrics.NewNoopCollector())

	suite.dbdir = unittest.TempDir(suite.T())
	suite.db = unittest.BadgerDB(suite.T(), suite.dbdir)

	metrics := metrics.NewNoopCollector()
	tracer := trace.NewNoopTracer()
	log := zerolog.Nop()

	all := bstorage.InitAll(metrics, suite.db)
	consumer := events.NewNoop()

	suite.headers = all.Headers
	suite.blocks = all.Blocks
	suite.payloads = bstorage.NewClusterPayloads(metrics, suite.db)

	// just bootstrap with a genesis block, we'll use this as reference
	root, result, seal := unittest.BootstrapFixture(unittest.IdentityListFixture(5, unittest.WithAllRoles()))
	// ensure we don't enter a new epoch for tests that build many blocks
	result.ServiceEvents[0].Event.(*flow.EpochSetup).FinalView = root.Header.View + 100000
	seal.ResultID = result.ID()
	safetyParams, err := protocol.DefaultEpochSafetyParams(root.Header.ChainID)
	require.NoError(suite.T(), err)
	minEpochStateEntry, err := inmem.EpochProtocolStateFromServiceEvents(
		result.ServiceEvents[0].Event.(*flow.EpochSetup),
		result.ServiceEvents[1].Event.(*flow.EpochCommit),
	)
	require.NoError(suite.T(), err)
	rootProtocolState, err := kvstore.NewDefaultKVStore(
		safetyParams.FinalizationSafetyThreshold,
		safetyParams.EpochExtensionViewCount,
		minEpochStateEntry.ID(),
	)
	require.NoError(suite.T(), err)
	root.Payload.ProtocolStateID = rootProtocolState.ID()
	rootSnapshot, err := inmem.SnapshotFromBootstrapState(root, result, seal, unittest.QuorumCertificateFixture(unittest.QCWithRootBlockID(root.ID())))
	require.NoError(suite.T(), err)
	suite.epochCounter = rootSnapshot.Encodable().SealingSegment.LatestProtocolStateEntry().EpochEntry.EpochCounter()

	require.NoError(suite.T(), err)
	clusterQC := unittest.QuorumCertificateFixture(unittest.QCWithRootBlockID(suite.genesis.ID()))
	minEpochStateEntry, err = inmem.EpochProtocolStateFromServiceEvents(
		result.ServiceEvents[0].Event.(*flow.EpochSetup),
		result.ServiceEvents[1].Event.(*flow.EpochCommit),
	)
	require.NoError(suite.T(), err)
	rootProtocolState, err = kvstore.NewDefaultKVStore(
		safetyParams.FinalizationSafetyThreshold, safetyParams.EpochExtensionViewCount,
		minEpochStateEntry.ID(),
	)
	require.NoError(suite.T(), err)
	root.Payload.ProtocolStateID = rootProtocolState.ID()
	clusterStateRoot, err := clusterkv.NewStateRoot(suite.genesis, clusterQC, suite.epochCounter)
	suite.Require().NoError(err)
	clusterState, err := clusterkv.Bootstrap(suite.db, clusterStateRoot)
	suite.Require().NoError(err)

	suite.state, err = clusterkv.NewMutableState(clusterState, tracer, suite.headers, suite.payloads)
	suite.Require().NoError(err)

	state, err := pbadger.Bootstrap(
		metrics,
		suite.db,
		all.Headers,
		all.Seals,
		all.Results,
		all.Blocks,
		all.QuorumCertificates,
		all.Setups,
		all.EpochCommits,
		all.EpochProtocolStateEntries,
		all.ProtocolKVStore,
		all.VersionBeacons,
		rootSnapshot,
	)
	require.NoError(suite.T(), err)

	suite.protoState, err = pbadger.NewFollowerState(
		log,
		tracer,
		consumer,
		state,
		all.Index,
		all.Payloads,
		util.MockBlockTimer(),
	)
	require.NoError(suite.T(), err)

	// add some transactions to transaction pool
	for i := 0; i < 3; i++ {
		transaction := unittest.TransactionBodyFixture(func(tx *flow.TransactionBody) {
			tx.ReferenceBlockID = root.ID()
			tx.ProposalKey.SequenceNumber = uint64(i)
			tx.GasLimit = uint64(9999)
		})
		added := suite.pool.Add(&transaction)
		suite.Assert().True(added)
	}

	suite.builder, _ = builder.NewBuilder(suite.db, tracer, suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter)
}

// runs after each test finishes
func (suite *BuilderSuite) TearDownTest() {
	err := suite.db.Close()
	suite.Assert().NoError(err)
	err = os.RemoveAll(suite.dbdir)
	suite.Assert().NoError(err)
}

func (suite *BuilderSuite) InsertBlock(block model.Block) {
	err := suite.db.Update(procedure.InsertClusterBlock(&block))
	suite.Assert().NoError(err)
}

func (suite *BuilderSuite) FinalizeBlock(block model.Block) {
	err := suite.db.Update(func(tx *badger.Txn) error {
		var refBlock flow.Header
		err := operation.RetrieveHeader(block.Payload.ReferenceBlockID, &refBlock)(tx)
		if err != nil {
			return err
		}
		err = procedure.FinalizeClusterBlock(block.ID())(tx)
		if err != nil {
			return err
		}
		err = operation.IndexClusterBlockByReferenceHeight(refBlock.Height, block.ID())(tx)
		return err
	})
	suite.Assert().NoError(err)
}

// Payload returns a payload containing the given transactions, with a valid
// reference block ID.
func (suite *BuilderSuite) Payload(transactions ...*flow.TransactionBody) model.Payload {
	final, err := suite.protoState.Final().Head()
	suite.Require().NoError(err)
	return model.PayloadFromTransactions(final.ID(), transactions...)
}

// ProtoStateRoot returns the root block of the protocol state.
func (suite *BuilderSuite) ProtoStateRoot() *flow.Header {
	return suite.protoState.Params().FinalizedRoot()
}

// ClearPool removes all items from the pool
func (suite *BuilderSuite) ClearPool() {
	// TODO use Clear()
	for _, tx := range suite.pool.All() {
		suite.pool.Remove(tx.ID())
	}
}

// FillPool adds n transactions to the pool, using the given generator function.
func (suite *BuilderSuite) FillPool(n int, create func() *flow.TransactionBody) {
	for i := 0; i < n; i++ {
		tx := create()
		suite.pool.Add(tx)
	}
}

func TestBuilder(t *testing.T) {
	suite.Run(t, new(BuilderSuite))
}

func (suite *BuilderSuite) TestBuildOn_NonExistentParent() {
	// use a non-existent parent ID
	parentID := unittest.IdentifierFixture()

	_, err := suite.builder.BuildOn(parentID, noopSetter, noopSigner)
	suite.Assert().Error(err)
}

func (suite *BuilderSuite) TestBuildOn_Success() {

	var expectedHeight uint64 = 42
	setter := func(h *flow.Header) error {
		h.Height = expectedHeight
		return nil
	}

	header, err := suite.builder.BuildOn(suite.genesis.ID(), setter, noopSigner)
	suite.Require().NoError(err)

	// setter should have been run
	suite.Assert().Equal(expectedHeight, header.Height)

	// should be able to retrieve built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Assert().NoError(err)
	builtCollection := built.Payload.Collection

	// should reference a valid reference block
	// (since genesis is the only block, it's the only valid reference)
	mainGenesis, err := suite.protoState.AtHeight(0).Head()
	suite.Assert().NoError(err)
	suite.Assert().Equal(mainGenesis.ID(), built.Payload.ReferenceBlockID)

	// payload should include only items from mempool
	mempoolTransactions := suite.pool.All()
	suite.Assert().Len(builtCollection.Transactions, 3)
	suite.Assert().True(collectionContains(builtCollection, flow.GetIDs(mempoolTransactions)...))
}

// TestBuildOn_SetterErrorPassthrough validates that errors from the setter function are passed through to the caller.
func (suite *BuilderSuite) TestBuildOn_SetterErrorPassthrough() {
	sentinel := errors.New("sentinel")
	setter := func(h *flow.Header) error {
		return sentinel
	}
	_, err := suite.builder.BuildOn(suite.genesis.ID(), setter, noopSigner)
	suite.Assert().ErrorIs(err, sentinel)
}

// TestBuildOn_SignerErrorPassthrough validates that errors from the sign function are passed through to the caller.
func (suite *BuilderSuite) TestBuildOn_SignerErrorPassthrough() {
	suite.T().Run("unexpected Exception", func(t *testing.T) {
		exception := errors.New("exception")
		sign := func(h *flow.Header) error {
			return exception
		}
		_, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, sign)
		suite.Assert().ErrorIs(err, exception)
	})
	suite.T().Run("NoVoteError", func(t *testing.T) {
		// the EventHandler relies on this sentinel in particular to be passed through
		sentinel := hotstuffmodel.NewNoVoteErrorf("not voting")
		sign := func(h *flow.Header) error {
			return sentinel
		}
		_, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, sign)
		suite.Assert().ErrorIs(err, sentinel)
	})
}

// when there are transactions with an unknown reference block in the pool, we should not include them in collections
func (suite *BuilderSuite) TestBuildOn_WithUnknownReferenceBlock() {

	// before modifying the mempool, note the valid transactions already in the pool
	validMempoolTransactions := suite.pool.All()

	// add a transaction unknown reference block to the pool
	unknownReferenceTx := unittest.TransactionBodyFixture()
	unknownReferenceTx.ReferenceBlockID = unittest.IdentifierFixture()
	suite.pool.Add(&unknownReferenceTx)

	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	// should be able to retrieve built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Assert().NoError(err)
	builtCollection := built.Payload.Collection

	suite.Assert().Len(builtCollection.Transactions, 3)
	// payload should include only the transactions with a valid reference block
	suite.Assert().True(collectionContains(builtCollection, flow.GetIDs(validMempoolTransactions)...))
	// should not contain the unknown-reference transaction
	suite.Assert().False(collectionContains(builtCollection, unknownReferenceTx.ID()))
}

// when there are transactions with a known but unfinalized reference block in the pool, we should not include them in collections
func (suite *BuilderSuite) TestBuildOn_WithUnfinalizedReferenceBlock() {

	// before modifying the mempool, note the valid transactions already in the pool
	validMempoolTransactions := suite.pool.All()

	// add an unfinalized block to the protocol state
	genesis, err := suite.protoState.Final().Head()
	suite.Require().NoError(err)
	protocolState, err := suite.protoState.Final().ProtocolState()
	suite.Require().NoError(err)
	protocolStateID := protocolState.ID()

	unfinalizedReferenceBlock := unittest.BlockWithParentFixture(genesis)
	unfinalizedReferenceBlock.SetPayload(unittest.PayloadFixture(unittest.WithProtocolStateID(protocolStateID)))
	err = suite.protoState.ExtendCertified(context.Background(), unfinalizedReferenceBlock,
		unittest.CertifyBlock(unfinalizedReferenceBlock.Header))
	suite.Require().NoError(err)

	// add a transaction with unfinalized reference block to the pool
	unfinalizedReferenceTx := unittest.TransactionBodyFixture()
	unfinalizedReferenceTx.ReferenceBlockID = unfinalizedReferenceBlock.ID()
	suite.pool.Add(&unfinalizedReferenceTx)

	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	// should be able to retrieve built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Assert().NoError(err)
	builtCollection := built.Payload.Collection

	suite.Assert().Len(builtCollection.Transactions, 3)
	// payload should include only the transactions with a valid reference block
	suite.Assert().True(collectionContains(builtCollection, flow.GetIDs(validMempoolTransactions)...))
	// should not contain the unfinalized-reference transaction
	suite.Assert().False(collectionContains(builtCollection, unfinalizedReferenceTx.ID()))
}

// when there are transactions with an orphaned reference block in the pool, we should not include them in collections
func (suite *BuilderSuite) TestBuildOn_WithOrphanedReferenceBlock() {

	// before modifying the mempool, note the valid transactions already in the pool
	validMempoolTransactions := suite.pool.All()

	// add an orphaned block to the protocol state
	genesis, err := suite.protoState.Final().Head()
	suite.Require().NoError(err)
	protocolState, err := suite.protoState.Final().ProtocolState()
	suite.Require().NoError(err)
	protocolStateID := protocolState.ID()

	// create a block extending genesis which will be orphaned
	orphan := unittest.BlockWithParentFixture(genesis)
	orphan.SetPayload(unittest.PayloadFixture(unittest.WithProtocolStateID(protocolStateID)))
	err = suite.protoState.ExtendCertified(context.Background(), orphan, unittest.CertifyBlock(orphan.Header))
	suite.Require().NoError(err)
	// create and finalize a block on top of genesis, orphaning `orphan`
	block1 := unittest.BlockWithParentFixture(genesis)
	block1.SetPayload(unittest.PayloadFixture(unittest.WithProtocolStateID(protocolStateID)))
	err = suite.protoState.ExtendCertified(context.Background(), block1, unittest.CertifyBlock(block1.Header))
	suite.Require().NoError(err)
	err = suite.protoState.Finalize(context.Background(), block1.ID())
	suite.Require().NoError(err)

	// add a transaction with orphaned reference block to the pool
	orphanedReferenceTx := unittest.TransactionBodyFixture()
	orphanedReferenceTx.ReferenceBlockID = orphan.ID()
	suite.pool.Add(&orphanedReferenceTx)

	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	// should be able to retrieve built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Assert().NoError(err)
	builtCollection := built.Payload.Collection

	suite.Assert().Len(builtCollection.Transactions, 3)
	// payload should include only the transactions with a valid reference block
	suite.Assert().True(collectionContains(builtCollection, flow.GetIDs(validMempoolTransactions)...))
	// should not contain the unknown-reference transaction
	suite.Assert().False(collectionContains(builtCollection, orphanedReferenceTx.ID()))
	// the transaction with orphaned reference should be removed from the mempool
	suite.Assert().False(suite.pool.Has(orphanedReferenceTx.ID()))
}

func (suite *BuilderSuite) TestBuildOn_WithForks() {
	t := suite.T()

	mempoolTransactions := suite.pool.All()
	tx1 := mempoolTransactions[0] // in fork 1
	tx2 := mempoolTransactions[1] // in fork 2
	tx3 := mempoolTransactions[2] // in no block

	// build first fork on top of genesis
	payload1 := suite.Payload(tx1)
	block1 := unittest.ClusterBlockWithParent(suite.genesis)
	block1.SetPayload(payload1)

	// insert block on fork 1
	suite.InsertBlock(block1)

	// build second fork on top of genesis
	payload2 := suite.Payload(tx2)
	block2 := unittest.ClusterBlockWithParent(suite.genesis)
	block2.SetPayload(payload2)

	// insert block on fork 2
	suite.InsertBlock(block2)

	// build on top of fork 1
	header, err := suite.builder.BuildOn(block1.ID(), noopSetter, noopSigner)
	require.NoError(t, err)

	// should be able to retrieve built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	assert.NoError(t, err)
	builtCollection := built.Payload.Collection

	// payload should include ONLY tx2 and tx3
	assert.Len(t, builtCollection.Transactions, 2)
	assert.True(t, collectionContains(builtCollection, tx2.ID(), tx3.ID()))
	assert.False(t, collectionContains(builtCollection, tx1.ID()))
}

func (suite *BuilderSuite) TestBuildOn_ConflictingFinalizedBlock() {
	t := suite.T()

	mempoolTransactions := suite.pool.All()
	tx1 := mempoolTransactions[0] // in a finalized block
	tx2 := mempoolTransactions[1] // in an un-finalized block
	tx3 := mempoolTransactions[2] // in no blocks

	t.Logf("tx1: %s\ntx2: %s\ntx3: %s", tx1.ID(), tx2.ID(), tx3.ID())

	// build a block containing tx1 on genesis
	finalizedPayload := suite.Payload(tx1)
	finalizedBlock := unittest.ClusterBlockWithParent(suite.genesis)
	finalizedBlock.SetPayload(finalizedPayload)
	suite.InsertBlock(finalizedBlock)
	t.Logf("finalized: height=%d id=%s txs=%s parent_id=%s\t\n", finalizedBlock.Header.Height, finalizedBlock.ID(), finalizedPayload.Collection.Light(), finalizedBlock.Header.ParentID)

	// build a block containing tx2 on the first block
	unFinalizedPayload := suite.Payload(tx2)
	unFinalizedBlock := unittest.ClusterBlockWithParent(&finalizedBlock)
	unFinalizedBlock.SetPayload(unFinalizedPayload)
	suite.InsertBlock(unFinalizedBlock)
	t.Logf("finalized: height=%d id=%s txs=%s parent_id=%s\t\n", unFinalizedBlock.Header.Height, unFinalizedBlock.ID(), unFinalizedPayload.Collection.Light(), unFinalizedBlock.Header.ParentID)

	// finalize first block
	suite.FinalizeBlock(finalizedBlock)

	// build on the un-finalized block
	header, err := suite.builder.BuildOn(unFinalizedBlock.ID(), noopSetter, noopSigner)
	require.NoError(t, err)

	// retrieve the built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	assert.NoError(t, err)
	builtCollection := built.Payload.Collection

	// payload should only contain tx3
	assert.Len(t, builtCollection.Light().Transactions, 1)
	assert.True(t, collectionContains(builtCollection, tx3.ID()))
	assert.False(t, collectionContains(builtCollection, tx1.ID(), tx2.ID()))

	// tx1 should be removed from mempool, as it is in a finalized block
	assert.False(t, suite.pool.Has(tx1.ID()))
	// tx2 should NOT be removed from mempool, as it is in an un-finalized block
	assert.True(t, suite.pool.Has(tx2.ID()))
}

func (suite *BuilderSuite) TestBuildOn_ConflictingInvalidatedForks() {
	t := suite.T()

	mempoolTransactions := suite.pool.All()
	tx1 := mempoolTransactions[0] // in a finalized block
	tx2 := mempoolTransactions[1] // in an invalidated block
	tx3 := mempoolTransactions[2] // in no blocks

	t.Logf("tx1: %s\ntx2: %s\ntx3: %s", tx1.ID(), tx2.ID(), tx3.ID())

	// build a block containing tx1 on genesis - will be finalized
	finalizedPayload := suite.Payload(tx1)
	finalizedBlock := unittest.ClusterBlockWithParent(suite.genesis)
	finalizedBlock.SetPayload(finalizedPayload)

	suite.InsertBlock(finalizedBlock)
	t.Logf("finalized: id=%s\tparent_id=%s\theight=%d\n", finalizedBlock.ID(), finalizedBlock.Header.ParentID, finalizedBlock.Header.Height)

	// build a block containing tx2 ALSO on genesis - will be invalidated
	invalidatedPayload := suite.Payload(tx2)
	invalidatedBlock := unittest.ClusterBlockWithParent(suite.genesis)
	invalidatedBlock.SetPayload(invalidatedPayload)
	suite.InsertBlock(invalidatedBlock)
	t.Logf("invalidated: id=%s\tparent_id=%s\theight=%d\n", invalidatedBlock.ID(), invalidatedBlock.Header.ParentID, invalidatedBlock.Header.Height)

	// finalize first block - this indirectly invalidates the second block
	suite.FinalizeBlock(finalizedBlock)

	// build on the finalized block
	header, err := suite.builder.BuildOn(finalizedBlock.ID(), noopSetter, noopSigner)
	require.NoError(t, err)

	// retrieve the built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	assert.NoError(t, err)
	builtCollection := built.Payload.Collection

	// tx2 and tx3 should be in the built collection
	assert.Len(t, builtCollection.Light().Transactions, 2)
	assert.True(t, collectionContains(builtCollection, tx2.ID(), tx3.ID()))
	assert.False(t, collectionContains(builtCollection, tx1.ID()))
}

func (suite *BuilderSuite) TestBuildOn_LargeHistory() {
	t := suite.T()

	// use a mempool with 2000 transactions, one per block
	suite.pool = herocache.NewTransactions(2000, unittest.Logger(), metrics.NewNoopCollector())
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter, builder.WithMaxCollectionSize(10000))

	// get a valid reference block ID
	final, err := suite.protoState.Final().Head()
	require.NoError(t, err)
	refID := final.ID()

	// keep track of the head of the chain
	head := *suite.genesis

	// keep track of invalidated transaction IDs
	var invalidatedTxIds []flow.Identifier

	// create a large history of blocks with invalidated forks every 3 blocks on
	// average - build until the height exceeds transaction expiry
	for i := 0; ; i++ {

		// create a transaction
		tx := unittest.TransactionBodyFixture(func(tx *flow.TransactionBody) {
			tx.ReferenceBlockID = refID
			tx.ProposalKey.SequenceNumber = uint64(i)
		})
		added := suite.pool.Add(&tx)
		assert.True(t, added)

		// 1/3 of the time create a conflicting fork that will be invalidated
		// don't do this the first and last few times to ensure we don't
		// try to fork genesis and the last block is the valid fork.
		conflicting := rand.Intn(3) == 0 && i > 5 && i < 995

		// by default, build on the head - if we are building a
		// conflicting fork, build on the parent of the head
		parent := head
		if conflicting {
			err = suite.db.View(procedure.RetrieveClusterBlock(parent.Header.ParentID, &parent))
			assert.NoError(t, err)
			// add the transaction to the invalidated list
			invalidatedTxIds = append(invalidatedTxIds, tx.ID())
		}

		// create a block containing the transaction
		block := unittest.ClusterBlockWithParent(&head)
		payload := suite.Payload(&tx)
		block.SetPayload(payload)
		suite.InsertBlock(block)

		// reset the valid head if we aren't building a conflicting fork
		if !conflicting {
			head = block
			suite.FinalizeBlock(block)
			assert.NoError(t, err)
		}

		// stop building blocks once we've built a history which exceeds the transaction
		// expiry length - this tests that deduplication works properly against old blocks
		// which nevertheless have a potentially conflicting reference block
		if head.Header.Height > flow.DefaultTransactionExpiry+100 {
			break
		}
	}

	t.Log("conflicting: ", len(invalidatedTxIds))

	// build on the head block
	header, err := suite.builder.BuildOn(head.ID(), noopSetter, noopSigner)
	require.NoError(t, err)

	// retrieve the built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	require.NoError(t, err)
	builtCollection := built.Payload.Collection

	// payload should only contain transactions from invalidated blocks
	assert.Len(t, builtCollection.Transactions, len(invalidatedTxIds), "expected len=%d, got len=%d", len(invalidatedTxIds), len(builtCollection.Transactions))
	assert.True(t, collectionContains(builtCollection, invalidatedTxIds...))
}

func (suite *BuilderSuite) TestBuildOn_MaxCollectionSize() {
	// set the max collection size to 1
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter, builder.WithMaxCollectionSize(1))

	// build a block
	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	// retrieve the built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Require().NoError(err)
	builtCollection := built.Payload.Collection

	// should be only 1 transaction in the collection
	suite.Assert().Equal(builtCollection.Len(), 1)
}

func (suite *BuilderSuite) TestBuildOn_MaxCollectionByteSize() {
	// set the max collection byte size to 400 (each tx is about 150 bytes)
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter, builder.WithMaxCollectionByteSize(400))

	// build a block
	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	// retrieve the built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Require().NoError(err)
	builtCollection := built.Payload.Collection

	// should be only 2 transactions in the collection, since each tx is ~273 bytes and the limit is 600 bytes
	suite.Assert().Equal(builtCollection.Len(), 2)
}

func (suite *BuilderSuite) TestBuildOn_MaxCollectionTotalGas() {
	// set the max gas to 20,000
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter, builder.WithMaxCollectionTotalGas(20000))

	// build a block
	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	// retrieve the built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Require().NoError(err)
	builtCollection := built.Payload.Collection

	// should be only 2 transactions in collection, since each transaction has gas limit of 9,999 and collection limit is set to 20,000
	suite.Assert().Equal(builtCollection.Len(), 2)
}

func (suite *BuilderSuite) TestBuildOn_ExpiredTransaction() {

	// create enough main-chain blocks that an expired transaction is possible
	genesis, err := suite.protoState.Final().Head()
	suite.Require().NoError(err)
	protocolState, err := suite.protoState.Final().ProtocolState()
	suite.Require().NoError(err)
	protocolStateID := protocolState.ID()

	head := genesis
	for i := 0; i < flow.DefaultTransactionExpiry+1; i++ {
		block := unittest.BlockWithParentFixture(head)
		block.SetPayload(unittest.PayloadFixture(unittest.WithProtocolStateID(protocolStateID)))
		err = suite.protoState.ExtendCertified(context.Background(), block, unittest.CertifyBlock(block.Header))
		suite.Require().NoError(err)
		err = suite.protoState.Finalize(context.Background(), block.ID())
		suite.Require().NoError(err)
		head = block.Header
	}

	// reset the pool and builder
	suite.pool = herocache.NewTransactions(10, unittest.Logger(), metrics.NewNoopCollector())
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter)

	// insert a transaction referring genesis (now expired)
	tx1 := unittest.TransactionBodyFixture(func(tx *flow.TransactionBody) {
		tx.ReferenceBlockID = genesis.ID()
		tx.ProposalKey.SequenceNumber = 0
	})
	added := suite.pool.Add(&tx1)
	suite.Assert().True(added)

	// insert a transaction referencing the head (valid)
	tx2 := unittest.TransactionBodyFixture(func(tx *flow.TransactionBody) {
		tx.ReferenceBlockID = head.ID()
		tx.ProposalKey.SequenceNumber = 1
	})
	added = suite.pool.Add(&tx2)
	suite.Assert().True(added)

	suite.T().Log("tx1: ", tx1.ID())
	suite.T().Log("tx2: ", tx2.ID())

	// build a block
	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	// retrieve the built block from storage
	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Require().NoError(err)
	builtCollection := built.Payload.Collection

	// the block should only contain the un-expired transaction
	suite.Assert().False(collectionContains(builtCollection, tx1.ID()))
	suite.Assert().True(collectionContains(builtCollection, tx2.ID()))
	// the expired transaction should have been removed from the mempool
	suite.Assert().False(suite.pool.Has(tx1.ID()))
}

func (suite *BuilderSuite) TestBuildOn_EmptyMempool() {

	// start with an empty mempool
	suite.pool = herocache.NewTransactions(1000, unittest.Logger(), metrics.NewNoopCollector())
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter)

	header, err := suite.builder.BuildOn(suite.genesis.ID(), noopSetter, noopSigner)
	suite.Require().NoError(err)

	var built model.Block
	err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
	suite.Require().NoError(err)

	// should reference a valid reference block
	// (since genesis is the only block, it's the only valid reference)
	mainGenesis, err := suite.protoState.AtHeight(0).Head()
	suite.Assert().NoError(err)
	suite.Assert().Equal(mainGenesis.ID(), built.Payload.ReferenceBlockID)

	// the payload should be empty
	suite.Assert().Equal(0, built.Payload.Collection.Len())
}

// With rate limiting turned off, we should fill collections as fast as we can
// regardless of how many transactions with the same payer we include.
func (suite *BuilderSuite) TestBuildOn_NoRateLimiting() {

	// start with an empty mempool
	suite.ClearPool()

	// create builder with no rate limit and max 10 tx/collection
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter,
		builder.WithMaxCollectionSize(10),
		builder.WithMaxPayerTransactionRate(0),
	)

	// fill the pool with 100 transactions from the same payer
	payer := unittest.RandomAddressFixture()
	create := func() *flow.TransactionBody {
		tx := unittest.TransactionBodyFixture()
		tx.ReferenceBlockID = suite.ProtoStateRoot().ID()
		tx.Payer = payer
		return &tx
	}
	suite.FillPool(100, create)

	// since we have no rate limiting we should fill all collections and in 10 blocks
	parentID := suite.genesis.ID()
	for i := 0; i < 10; i++ {
		header, err := suite.builder.BuildOn(parentID, noopSetter, noopSigner)
		suite.Require().NoError(err)
		parentID = header.ID()

		// each collection should be full with 10 transactions
		var built model.Block
		err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
		suite.Assert().NoError(err)
		suite.Assert().Len(built.Payload.Collection.Transactions, 10)
	}
}

// With rate limiting turned on, we should be able to fill transactions as fast
// as possible so long as per-payer limits are not reached. This test generates
// transactions such that the number of transactions with a given proposer exceeds
// the rate limit -- since it's the proposer not the payer, it shouldn't limit
// our collections.
func (suite *BuilderSuite) TestBuildOn_RateLimitNonPayer() {

	// start with an empty mempool
	suite.ClearPool()

	// create builder with 5 tx/payer and max 10 tx/collection
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter,
		builder.WithMaxCollectionSize(10),
		builder.WithMaxPayerTransactionRate(5),
	)

	// fill the pool with 100 transactions with the same proposer
	// since it's not the same payer, rate limit does not apply
	proposer := unittest.RandomAddressFixture()
	create := func() *flow.TransactionBody {
		tx := unittest.TransactionBodyFixture()
		tx.ReferenceBlockID = suite.ProtoStateRoot().ID()
		tx.Payer = unittest.RandomAddressFixture()
		tx.ProposalKey = flow.ProposalKey{
			Address:        proposer,
			KeyIndex:       rand.Uint32(),
			SequenceNumber: rand.Uint64(),
		}
		return &tx
	}
	suite.FillPool(100, create)

	// since rate limiting does not apply to non-payer keys, we should fill all collections in 10 blocks
	parentID := suite.genesis.ID()
	for i := 0; i < 10; i++ {
		header, err := suite.builder.BuildOn(parentID, noopSetter, noopSigner)
		suite.Require().NoError(err)
		parentID = header.ID()

		// each collection should be full with 10 transactions
		var built model.Block
		err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
		suite.Assert().NoError(err)
		suite.Assert().Len(built.Payload.Collection.Transactions, 10)
	}
}

// When configured with a rate limit of k>1, we should be able to include up to
// k transactions with a given payer per collection
func (suite *BuilderSuite) TestBuildOn_HighRateLimit() {

	// start with an empty mempool
	suite.ClearPool()

	// create builder with 5 tx/payer and max 10 tx/collection
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter,
		builder.WithMaxCollectionSize(10),
		builder.WithMaxPayerTransactionRate(5),
	)

	// fill the pool with 50 transactions from the same payer
	payer := unittest.RandomAddressFixture()
	create := func() *flow.TransactionBody {
		tx := unittest.TransactionBodyFixture()
		tx.ReferenceBlockID = suite.ProtoStateRoot().ID()
		tx.Payer = payer
		return &tx
	}
	suite.FillPool(50, create)

	// rate-limiting should be applied, resulting in half-full collections (5/10)
	parentID := suite.genesis.ID()
	for i := 0; i < 10; i++ {
		header, err := suite.builder.BuildOn(parentID, noopSetter, noopSigner)
		suite.Require().NoError(err)
		parentID = header.ID()

		// each collection should be half-full with 5 transactions
		var built model.Block
		err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
		suite.Assert().NoError(err)
		suite.Assert().Len(built.Payload.Collection.Transactions, 5)
	}
}

// When configured with a rate limit of k<1, we should be able to include 1
// transactions with a given payer every ceil(1/k) collections
func (suite *BuilderSuite) TestBuildOn_LowRateLimit() {

	// start with an empty mempool
	suite.ClearPool()

	// create builder with .5 tx/payer and max 10 tx/collection
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter,
		builder.WithMaxCollectionSize(10),
		builder.WithMaxPayerTransactionRate(.5),
	)

	// fill the pool with 5 transactions from the same payer
	payer := unittest.RandomAddressFixture()
	create := func() *flow.TransactionBody {
		tx := unittest.TransactionBodyFixture()
		tx.ReferenceBlockID = suite.ProtoStateRoot().ID()
		tx.Payer = payer
		return &tx
	}
	suite.FillPool(5, create)

	// rate-limiting should be applied, resulting in every ceil(1/k) collections
	// having one transaction and empty collections otherwise
	parentID := suite.genesis.ID()
	for i := 0; i < 10; i++ {
		header, err := suite.builder.BuildOn(parentID, noopSetter, noopSigner)
		suite.Require().NoError(err)
		parentID = header.ID()

		// collections should either be empty or have 1 transaction
		var built model.Block
		err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
		suite.Assert().NoError(err)
		if i%2 == 0 {
			suite.Assert().Len(built.Payload.Collection.Transactions, 1)
		} else {
			suite.Assert().Len(built.Payload.Collection.Transactions, 0)
		}
	}
}
func (suite *BuilderSuite) TestBuildOn_UnlimitedPayer() {

	// start with an empty mempool
	suite.ClearPool()

	// create builder with 5 tx/payer and max 10 tx/collection
	// configure an unlimited payer
	payer := unittest.RandomAddressFixture()
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter,
		builder.WithMaxCollectionSize(10),
		builder.WithMaxPayerTransactionRate(5),
		builder.WithUnlimitedPayers(payer),
	)

	// fill the pool with 100 transactions from the same payer
	create := func() *flow.TransactionBody {
		tx := unittest.TransactionBodyFixture()
		tx.ReferenceBlockID = suite.ProtoStateRoot().ID()
		tx.Payer = payer
		return &tx
	}
	suite.FillPool(100, create)

	// rate-limiting should not be applied, since the payer is marked as unlimited
	parentID := suite.genesis.ID()
	for i := 0; i < 10; i++ {
		header, err := suite.builder.BuildOn(parentID, noopSetter, noopSigner)
		suite.Require().NoError(err)
		parentID = header.ID()

		// each collection should be full with 10 transactions
		var built model.Block
		err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
		suite.Assert().NoError(err)
		suite.Assert().Len(built.Payload.Collection.Transactions, 10)

	}
}

// TestBuildOn_RateLimitDryRun tests that rate limiting rules aren't enforced
// if dry-run is enabled.
func (suite *BuilderSuite) TestBuildOn_RateLimitDryRun() {

	// start with an empty mempool
	suite.ClearPool()

	// create builder with 5 tx/payer and max 10 tx/collection
	// configure an unlimited payer
	payer := unittest.RandomAddressFixture()
	suite.builder, _ = builder.NewBuilder(suite.db, trace.NewNoopTracer(), suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter,
		builder.WithMaxCollectionSize(10),
		builder.WithMaxPayerTransactionRate(5),
		builder.WithRateLimitDryRun(true),
	)

	// fill the pool with 100 transactions from the same payer
	create := func() *flow.TransactionBody {
		tx := unittest.TransactionBodyFixture()
		tx.ReferenceBlockID = suite.ProtoStateRoot().ID()
		tx.Payer = payer
		return &tx
	}
	suite.FillPool(100, create)

	// rate-limiting should not be applied, since dry-run setting is enabled
	parentID := suite.genesis.ID()
	for i := 0; i < 10; i++ {
		header, err := suite.builder.BuildOn(parentID, noopSetter, noopSigner)
		suite.Require().NoError(err)
		parentID = header.ID()

		// each collection should be full with 10 transactions
		var built model.Block
		err = suite.db.View(procedure.RetrieveClusterBlock(header.ID(), &built))
		suite.Assert().NoError(err)
		suite.Assert().Len(built.Payload.Collection.Transactions, 10)
	}
}

// helper to check whether a collection contains each of the given transactions.
func collectionContains(collection flow.Collection, txIDs ...flow.Identifier) bool {

	lookup := make(map[flow.Identifier]struct{}, len(txIDs))
	for _, tx := range collection.Transactions {
		lookup[tx.ID()] = struct{}{}
	}

	for _, txID := range txIDs {
		_, exists := lookup[txID]
		if !exists {
			return false
		}
	}

	return true
}

func BenchmarkBuildOn10(b *testing.B)     { benchmarkBuildOn(b, 10) }
func BenchmarkBuildOn100(b *testing.B)    { benchmarkBuildOn(b, 100) }
func BenchmarkBuildOn1000(b *testing.B)   { benchmarkBuildOn(b, 1000) }
func BenchmarkBuildOn10000(b *testing.B)  { benchmarkBuildOn(b, 10000) }
func BenchmarkBuildOn100000(b *testing.B) { benchmarkBuildOn(b, 100000) }

func benchmarkBuildOn(b *testing.B, size int) {
	b.StopTimer()
	b.ResetTimer()

	// re-use the builder suite
	suite := new(BuilderSuite)

	// Copied from SetupTest. We can't use that function because suite.Assert
	// is incompatible with benchmarks.
	// ref: https://github.com/stretchr/testify/issues/811
	{
		var err error

		suite.genesis = model.Genesis()
		suite.chainID = suite.genesis.Header.ChainID

		suite.pool = herocache.NewTransactions(1000, unittest.Logger(), metrics.NewNoopCollector())

		suite.dbdir = unittest.TempDir(b)
		suite.db = unittest.BadgerDB(b, suite.dbdir)
		defer func() {
			err = suite.db.Close()
			assert.NoError(b, err)
			err = os.RemoveAll(suite.dbdir)
			assert.NoError(b, err)
		}()

		metrics := metrics.NewNoopCollector()
		tracer := trace.NewNoopTracer()
		all := bstorage.InitAll(metrics, suite.db)
		suite.headers = all.Headers
		suite.blocks = all.Blocks
		suite.payloads = bstorage.NewClusterPayloads(metrics, suite.db)

		qc := unittest.QuorumCertificateFixture(unittest.QCWithRootBlockID(suite.genesis.ID()))
		stateRoot, err := clusterkv.NewStateRoot(suite.genesis, qc, suite.epochCounter)

		state, err := clusterkv.Bootstrap(suite.db, stateRoot)
		assert.NoError(b, err)

		suite.state, err = clusterkv.NewMutableState(state, tracer, suite.headers, suite.payloads)
		assert.NoError(b, err)

		// add some transactions to transaction pool
		for i := 0; i < 3; i++ {
			tx := unittest.TransactionBodyFixture()
			added := suite.pool.Add(&tx)
			assert.True(b, added)
		}

		// create the builder
		suite.builder, _ = builder.NewBuilder(suite.db, tracer, suite.protoState, suite.state, suite.headers, suite.headers, suite.payloads, suite.pool, unittest.Logger(), suite.epochCounter)
	}

	// create a block history to test performance against
	final := suite.genesis
	for i := 0; i < size; i++ {
		block := unittest.ClusterBlockWithParent(final)
		err := suite.db.Update(procedure.InsertClusterBlock(&block))
		require.NoError(b, err)

		// finalize the block 80% of the time, resulting in a fork-rate of 20%
		if rand.Intn(100) < 80 {
			err = suite.db.Update(procedure.FinalizeClusterBlock(block.ID()))
			require.NoError(b, err)
			final = &block
		}
	}

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		_, err := suite.builder.BuildOn(final.ID(), noopSetter, noopSigner)
		assert.NoError(b, err)
	}
}
