package computation

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/onflow/cadence/runtime"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/engine/execution/computation/committer"
	"github.com/onflow/flow-go/engine/execution/computation/computer"
	"github.com/onflow/flow-go/engine/execution/state/delta"
	"github.com/onflow/flow-go/engine/execution/testutil"
	"github.com/onflow/flow-go/fvm"
	"github.com/onflow/flow-go/fvm/programs"
	reusableRuntime "github.com/onflow/flow-go/fvm/runtime"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/executiondatasync/execution_data"
	exedataprovider "github.com/onflow/flow-go/module/executiondatasync/provider"
	"github.com/onflow/flow-go/module/executiondatasync/tracker"
	mocktracker "github.com/onflow/flow-go/module/executiondatasync/tracker/mock"
	"github.com/onflow/flow-go/module/mempool/entity"
	"github.com/onflow/flow-go/module/metrics"
	module "github.com/onflow/flow-go/module/mock"
	requesterunit "github.com/onflow/flow-go/module/state_synchronization/requester/unittest"
	"github.com/onflow/flow-go/module/trace"
	"github.com/onflow/flow-go/utils/unittest"
)

type testAccount struct {
	address    flow.Address
	privateKey flow.AccountPrivateKey
}

type testAccounts struct {
	accounts []testAccount
	seq      uint64
}

func createAccounts(b *testing.B, vm *fvm.VirtualMachine, ledger state.View, num int) *testAccounts {
	privateKeys, err := testutil.GenerateAccountPrivateKeys(num)
	require.NoError(b, err)

	addresses, err := testutil.CreateAccounts(vm, ledger, programs.NewEmptyBlockPrograms(), privateKeys, chain)
	require.NoError(b, err)

	accs := &testAccounts{
		accounts: make([]testAccount, num),
	}
	for i := 0; i < num; i++ {
		accs.accounts[i] = testAccount{
			address:    addresses[i],
			privateKey: privateKeys[i],
		}
	}
	return accs
}

func mustFundAccounts(
	b *testing.B,
	vm *fvm.VirtualMachine,
	ledger state.View,
	execCtx fvm.Context,
	accs *testAccounts,
) {
	blockPrograms := programs.NewEmptyBlockPrograms()
	execCtx = fvm.NewContextFromParent(
		execCtx,
		fvm.WithBlockPrograms(blockPrograms))

	var err error
	for _, acc := range accs.accounts {
		transferTx := testutil.CreateTokenTransferTransaction(chain, 1_000_000, acc.address, chain.ServiceAddress())
		err = testutil.SignTransactionAsServiceAccount(transferTx, accs.seq, chain)
		require.NoError(b, err)
		accs.seq++

		tx := fvm.Transaction(
			transferTx,
			blockPrograms.NextTxIndexForTestingOnly())
		err = vm.RunV2(execCtx, tx, ledger)
		require.NoError(b, err)
		require.NoError(b, tx.Err)
	}
}

func BenchmarkComputeBlock(b *testing.B) {
	b.StopTimer()

	tracer, err := trace.NewTracer(zerolog.Nop(), "", "", 4)
	require.NoError(b, err)

	vm := fvm.NewVM()

	chain := flow.Emulator.Chain()
	execCtx := fvm.NewContext(
		fvm.WithChain(chain),
		fvm.WithAccountStorageLimit(true),
		fvm.WithTransactionFeesEnabled(true),
		fvm.WithTracer(tracer),
		fvm.WithReusableCadenceRuntimePool(
			reusableRuntime.NewReusableCadenceRuntimePool(
				ReusableCadenceRuntimePoolSize,
				runtime.Config{})),
	)
	ledger := testutil.RootBootstrappedLedger(
		vm,
		execCtx,
		fvm.WithAccountCreationFee(fvm.DefaultAccountCreationFee),
		fvm.WithMinimumStorageReservation(fvm.DefaultMinimumStorageReservation),
		fvm.WithTransactionFee(fvm.DefaultTransactionFees),
		fvm.WithStorageMBPerFLOW(fvm.DefaultStorageMBPerFLOW),
	)
	accs := createAccounts(b, vm, ledger, 1000)
	mustFundAccounts(b, vm, ledger, execCtx, accs)

	me := new(module.Local)
	me.On("NodeID").Return(flow.ZeroID)

	bservice := requesterunit.MockBlobService(blockstore.NewBlockstore(dssync.MutexWrap(datastore.NewMapDatastore())))
	trackerStorage := new(mocktracker.Storage)
	trackerStorage.On("Update", mock.Anything).Return(func(fn tracker.UpdateFn) error {
		return fn(func(uint64, ...cid.Cid) error { return nil })
	})

	prov := exedataprovider.NewProvider(
		zerolog.Nop(),
		metrics.NewNoopCollector(),
		execution_data.DefaultSerializer,
		bservice,
		trackerStorage,
	)

	// TODO(rbtz): add real ledger
	blockComputer, err := computer.NewBlockComputer(vm, execCtx, metrics.NewNoopCollector(), tracer, zerolog.Nop(), committer.NewNoopViewCommitter(), prov)
	require.NoError(b, err)

	programsCache, err := programs.NewChainPrograms(
		programs.DefaultProgramsCacheSize)
	require.NoError(b, err)

	engine := &Manager{
		blockComputer: blockComputer,
		tracer:        tracer,
		me:            me,
		programsCache: programsCache,
	}

	view := delta.NewView(ledger.Get)
	blockView := view.NewChild()

	b.SetParallelism(1)

	parentBlock := &flow.Block{
		Header:  &flow.Header{},
		Payload: &flow.Payload{},
	}

	const (
		cols = 16
		txes = 128
	)

	b.Run(fmt.Sprintf("%d/cols/%d/txes", cols, txes), func(b *testing.B) {
		b.StopTimer()
		b.ResetTimer()

		var elapsed time.Duration
		for i := 0; i < b.N; i++ {
			executableBlock := createBlock(b, parentBlock, accs, cols, txes)
			parentBlock = executableBlock.Block

			b.StartTimer()
			start := time.Now()
			res, err := engine.ComputeBlock(context.Background(), executableBlock, blockView)
			elapsed += time.Since(start)
			b.StopTimer()

			require.NoError(b, err)
			for j, r := range res.TransactionResults {
				// skip system transactions
				if j >= cols*txes {
					break
				}
				require.Emptyf(b, r.ErrorMessage, "Transaction %d failed", j)
			}
		}
		totalTxes := int64(cols) * int64(txes) * int64(b.N)
		b.ReportMetric(float64(elapsed.Nanoseconds()/totalTxes/int64(time.Microsecond)), "us/tx")
	})
}

func createBlock(b *testing.B, parentBlock *flow.Block, accs *testAccounts, colNum int, txNum int) *entity.ExecutableBlock {
	completeCollections := make(map[flow.Identifier]*entity.CompleteCollection, colNum)
	collections := make([]*flow.Collection, colNum)
	guarantees := make([]*flow.CollectionGuarantee, colNum)

	for c := 0; c < colNum; c++ {
		transactions := make([]*flow.TransactionBody, txNum)
		for t := 0; t < txNum; t++ {
			transactions[t] = createTokenTransferTransaction(b, accs)
		}

		collection := &flow.Collection{Transactions: transactions}
		guarantee := &flow.CollectionGuarantee{CollectionID: collection.ID()}

		collections[c] = collection
		guarantees[c] = guarantee
		completeCollections[guarantee.ID()] = &entity.CompleteCollection{
			Guarantee:    guarantee,
			Transactions: transactions,
		}
	}

	block := flow.Block{
		Header: &flow.Header{
			ParentID: parentBlock.ID(),
			View:     parentBlock.Header.Height + 1,
		},
		Payload: &flow.Payload{
			Guarantees: guarantees,
		},
	}

	return &entity.ExecutableBlock{
		Block:               &block,
		CompleteCollections: completeCollections,
		StartState:          unittest.StateCommitmentPointerFixture(),
	}
}

func createTokenTransferTransaction(b *testing.B, accs *testAccounts) *flow.TransactionBody {
	var err error

	rnd := rand.Intn(len(accs.accounts))
	src := accs.accounts[rnd]
	dst := accs.accounts[(rnd+1)%len(accs.accounts)]

	tx := testutil.CreateTokenTransferTransaction(chain, 1, dst.address, src.address)
	tx.SetProposalKey(chain.ServiceAddress(), 0, accs.seq).
		SetGasLimit(1000).
		SetPayer(chain.ServiceAddress())
	accs.seq++

	err = testutil.SignPayload(tx, src.address, src.privateKey)
	require.NoError(b, err)

	err = testutil.SignEnvelope(tx, chain.ServiceAddress(), unittest.ServiceAccountPrivateKey)
	require.NoError(b, err)

	return tx
}
