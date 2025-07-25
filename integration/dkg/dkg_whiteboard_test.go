package dkg

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"golang.org/x/exp/slices"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/crypto"

	dkgeng "github.com/onflow/flow-go/engine/consensus/dkg"
	"github.com/onflow/flow-go/engine/testutil"
	"github.com/onflow/flow-go/model/bootstrap"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/dkg"
	"github.com/onflow/flow-go/module/metrics"
	msig "github.com/onflow/flow-go/module/signature"
	"github.com/onflow/flow-go/network/stub"
	"github.com/onflow/flow-go/state/protocol/events/gadgets"
	protocolmock "github.com/onflow/flow-go/state/protocol/mock"
	"github.com/onflow/flow-go/storage/badger"
	"github.com/onflow/flow-go/utils/unittest"
	"github.com/onflow/flow-go/utils/unittest/mocks"
)

// create a set of nodes that share the same hub for networking, the same
// whiteboard for exchanging DKG broadcast messages, and a mocked state
// containing the expected next EpochSetup event
func createNodes(
	t *testing.T,
	hub *stub.Hub,
	chainID flow.ChainID,
	whiteboard *whiteboard,
	conIdentities []bootstrap.NodeInfo,
	currentEpochSetup flow.EpochSetup,
	nextEpochSetup flow.EpochSetup,
	firstBlock *flow.Header) []*node {

	identities := make(flow.IdentityList, 0, len(conIdentities))
	for _, identity := range conIdentities {
		identities = append(identities, identity.Identity())
	}

	// We need to initialise the nodes with a list of identities that contain
	// all roles, otherwise there would be an error initialising the first epoch
	identities = unittest.CompleteIdentitySet(identities...)

	nodes := []*node{}
	for _, id := range conIdentities {
		nodes = append(nodes, createNode(t,
			id,
			identities,
			hub,
			chainID,
			whiteboard,
			currentEpochSetup,
			nextEpochSetup,
			firstBlock))
	}

	return nodes
}

// createNode instantiates a node with a network hub, a whiteboard reference,
// and a pre-set EpochSetup that will be used to trigger the next DKG run.
func createNode(
	t *testing.T,
	id bootstrap.NodeInfo,
	ids []*flow.Identity,
	hub *stub.Hub,
	chainID flow.ChainID,
	whiteboard *whiteboard,
	currentSetup flow.EpochSetup,
	nextSetup flow.EpochSetup,
	firstBlock *flow.Header) *node {

	core := testutil.GenericNodeFromParticipants(t, hub, id, ids, chainID)
	core.Log = zerolog.New(os.Stdout).Level(zerolog.WarnLevel)

	// the viewsObserver is used by the reactor engine to subscribe to when
	// blocks are finalized that are in a new view
	viewsObserver := gadgets.NewViews()
	core.ProtocolEvents.AddConsumer(viewsObserver)

	// keyKeys is used to store the private key resulting from the node's
	// participation in the DKG run
	dkgState, err := badger.NewRecoverableRandomBeaconStateMachine(core.Metrics, core.SecretsDB, core.Me.NodeID())
	require.NoError(t, err)

	// configure the state snapshot at firstBlock to return the desired
	// Epochs
	currentEpoch := new(protocolmock.CommittedEpoch)
	currentEpoch.On("Counter").Return(currentSetup.Counter)
	currentEpoch.On("InitialIdentities").Return(currentSetup.Participants)
	currentEpoch.On("DKGPhase1FinalView").Return(currentSetup.DKGPhase1FinalView)
	currentEpoch.On("DKGPhase2FinalView").Return(currentSetup.DKGPhase2FinalView)
	currentEpoch.On("DKGPhase3FinalView").Return(currentSetup.DKGPhase3FinalView)
	currentEpoch.On("RandomSource").Return(nextSetup.RandomSource)

	nextEpoch := new(protocolmock.TentativeEpoch)
	nextEpoch.On("Counter").Return(nextSetup.Counter)
	nextEpoch.On("InitialIdentities").Return(nextSetup.Participants)

	epochQuery := mocks.NewEpochQuery(t, currentSetup.Counter)
	epochQuery.AddCommitted(currentEpoch)
	epochQuery.AddTentative(nextEpoch)
	snapshot := new(protocolmock.Snapshot)
	snapshot.On("Epochs").Return(epochQuery)
	snapshot.On("EpochPhase").Return(flow.EpochPhaseStaking, nil)
	snapshot.On("Head").Return(firstBlock, nil)
	state := new(protocolmock.ParticipantState)
	state.On("AtBlockID", firstBlock.ID()).Return(snapshot)
	state.On("Final").Return(snapshot)
	core.State = state

	// brokerTunnel is used to communicate between the messaging engine and the
	// DKG broker/controller
	brokerTunnel := dkg.NewBrokerTunnel()

	// messagingEngine is a network engine that is used by nodes to exchange
	// private DKG messages
	messagingEngine, err := dkgeng.NewMessagingEngine(
		core.Log,
		core.Net,
		core.Me,
		brokerTunnel,
		metrics.NewNoopCollector(),
		dkgeng.DefaultMessagingEngineConfig(),
	)
	require.NoError(t, err)

	// We add a hook to the logger such that the test fails if the broker writes
	// a Warn log, which happens when it flags or disqualifies a node
	hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		if level == zerolog.WarnLevel {
			t.Fatal("DKG flagging misbehaviour")
		}
	})
	controllerFactoryLogger := zerolog.New(os.Stdout).Hook(hook)

	// the reactor engine reacts to new views being finalized and drives the
	// DKG protocol
	reactorEngine := dkgeng.NewReactorEngine(
		core.Log,
		core.Me,
		core.State,
		dkgState,
		dkg.NewControllerFactory(
			controllerFactoryLogger,
			core.Me,
			[]module.DKGContractClient{NewWhiteboardClient(id.NodeID, whiteboard)},
			brokerTunnel,
		),
		viewsObserver,
	)

	// reactorEngine consumes the EpochSetupPhaseStarted event
	core.ProtocolEvents.AddConsumer(reactorEngine)

	node := node{
		t:               t,
		GenericNode:     core,
		dkgState:        dkgState,
		messagingEngine: messagingEngine,
		reactorEngine:   reactorEngine,
	}

	return &node
}

// TestWithWhiteboard tests the DKG protocol against a mocked out DKG smart contract (whiteboard).
func TestWithWhiteboard(t *testing.T) {

	// hub is an in-memory test network that enables nodes to communicate using
	// the DKG messaging engine
	hub := stub.NewNetworkHub()

	// whiteboard is a shared object where DKG nodes can publish/read broadcast
	// messages, as well as publish end results, using a special
	// DKGContractClient.
	whiteboard := newWhiteboard()

	chainID := flow.Testnet

	// we run the DKG protocol with N consensus nodes
	N := 10
	bootstrapNodesInfo := unittest.PrivateNodeInfosFixture(N, unittest.WithRole(flow.RoleConsensus))
	slices.SortFunc(bootstrapNodesInfo, func(lhs, rhs bootstrap.NodeInfo) int {
		return flow.IdentifierCanonical(lhs.NodeID, rhs.NodeID)
	})
	conIdentities := make(flow.IdentitySkeletonList, 0, len(bootstrapNodesInfo))
	for _, identity := range bootstrapNodesInfo {
		conIdentities = append(conIdentities, &identity.Identity().IdentitySkeleton)
	}

	// The EpochSetup event is received at view 100. The phase transitions are
	// at views 150, 200, and 250. In between phase transitions, the controller
	// calls the DKG smart-contract every 10 views.
	//
	// VIEWS
	// setup      : 100
	// polling    : 110 120 130 140 150
	// Phase1Final: 150
	// polling    : 160 170 180 190 200
	// Phase2Final: 200
	// polling    : 210 220 230 240 250
	// Phase3Final: 250
	// final      : 300

	// create and record relevant blocks
	blocks := make(map[uint64]*flow.Header)
	var view uint64
	for view = 100; view <= 250; view += dkgeng.DefaultPollStep {
		blocks[view] = &flow.Header{View: view}
	}
	firstBlock := blocks[100]

	// we arbitrarily use 999 as the current epoch counter
	currentCounter := uint64(999)

	currentEpochSetup := flow.EpochSetup{
		Counter:            currentCounter,
		DKGPhase1FinalView: 150,
		DKGPhase2FinalView: 200,
		DKGPhase3FinalView: 250,
		FinalView:          300,
		Participants:       conIdentities.ToSkeleton(),
		RandomSource:       unittest.EpochSetupRandomSourceFixture(),
	}

	// create the EpochSetup that will trigger the next DKG run with all the
	// desired parameters
	nextEpochSetup := flow.EpochSetup{
		Counter:      currentCounter + 1,
		Participants: conIdentities.ToSkeleton(),
		RandomSource: unittest.EpochSetupRandomSourceFixture(),
	}

	nodes := createNodes(
		t,
		hub,
		chainID,
		whiteboard,
		bootstrapNodesInfo,
		currentEpochSetup,
		nextEpochSetup,
		firstBlock,
	)

	for _, node := range nodes {
		node.Start()
		unittest.RequireCloseBefore(t, node.Ready(), time.Second, "failed to start up")
	}
	// trigger the EpochSetupPhaseStarted event for all nodes, effectively
	// starting the next DKG run
	for _, n := range nodes {
		n.ProtocolEvents.EpochSetupPhaseStarted(currentCounter, firstBlock)
	}

	// trigger the BlockFinalized events for each view of interest, effectively
	// causing the DKG state machine to progress
	for view = 100; view <= 250; view += dkgeng.DefaultPollStep {
		time.Sleep(300 * time.Millisecond)
		hub.DeliverAll()
		for _, n := range nodes {
			n.ProtocolEvents.BlockFinalized(blocks[view])
		}
	}

	for _, n := range nodes {
		n.Stop()
		unittest.RequireCloseBefore(t, n.Done(), time.Second, "nodes did not shutdown")
	}

	t.Logf("there are %d result(s)", len(whiteboard.results))
	assert.Equal(t, 1, len(whiteboard.results))
	tag := "some tag"
	hasher := msig.NewBLSHasher(tag)

	for _, result := range whiteboard.results {
		signers := whiteboard.resultSubmitters[flow.MakeID(result)]
		t.Logf("result %s has %d proposers", flow.MakeID(result), len(signers))
		assert.Equal(t, N, len(signers))
	}

	// create and test a threshold signature with the keys computed by dkg
	sigData := []byte("message to be signed")
	signatures := []crypto.Signature{}
	indices := []int{}
	for i, n := range nodes {
		beaconKey, err := n.dkgState.UnsafeRetrieveMyBeaconPrivateKey(nextEpochSetup.Counter)
		require.NoError(t, err)

		signature, err := beaconKey.Sign(sigData, hasher)
		require.NoError(t, err)

		signatures = append(signatures, signature)
		indices = append(indices, i)
	}

	// shuffle the signatures and indices before constructing the group
	// signature (since it only uses the first half signatures)
	rand.Shuffle(len(signatures), func(i, j int) {
		signatures[i], signatures[j] = signatures[j], signatures[i]
		indices[i], indices[j] = indices[j], indices[i]
	})

	threshold := msig.RandomBeaconThreshold(len(nodes))
	groupSignature, err := crypto.BLSReconstructThresholdSignature(len(nodes), threshold, signatures, indices)
	require.NoError(t, err)

	result := whiteboard.resultBySubmitter[nodes[0].Me.NodeID()]
	groupPk := result.groupKey
	ok, err := groupPk.Verify(groupSignature, sigData, hasher)
	require.NoError(t, err)
	assert.True(t, ok, "failed to verify threshold signature")
}
