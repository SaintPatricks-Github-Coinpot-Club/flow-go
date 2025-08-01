package node_builder

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ipfs/boxo/bitswap"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/onflow/crypto"
	"github.com/onflow/flow/protobuf/go/flow/access"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	txvalidator "github.com/onflow/flow-go/access/validator"
	"github.com/onflow/flow-go/admin/commands"
	stateSyncCommands "github.com/onflow/flow-go/admin/commands/state_synchronization"
	storageCommands "github.com/onflow/flow-go/admin/commands/storage"
	"github.com/onflow/flow-go/cmd"
	"github.com/onflow/flow-go/cmd/build"
	"github.com/onflow/flow-go/consensus"
	"github.com/onflow/flow-go/consensus/hotstuff"
	"github.com/onflow/flow-go/consensus/hotstuff/committees"
	"github.com/onflow/flow-go/consensus/hotstuff/notifications"
	consensuspubsub "github.com/onflow/flow-go/consensus/hotstuff/notifications/pubsub"
	"github.com/onflow/flow-go/consensus/hotstuff/signature"
	hotstuffvalidator "github.com/onflow/flow-go/consensus/hotstuff/validator"
	"github.com/onflow/flow-go/consensus/hotstuff/verification"
	recovery "github.com/onflow/flow-go/consensus/recovery/protocol"
	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/engine/access/index"
	"github.com/onflow/flow-go/engine/access/ingestion"
	"github.com/onflow/flow-go/engine/access/ingestion/tx_error_messages"
	pingeng "github.com/onflow/flow-go/engine/access/ping"
	"github.com/onflow/flow-go/engine/access/rest"
	commonrest "github.com/onflow/flow-go/engine/access/rest/common"
	"github.com/onflow/flow-go/engine/access/rest/router"
	"github.com/onflow/flow-go/engine/access/rest/websockets"
	"github.com/onflow/flow-go/engine/access/rpc"
	"github.com/onflow/flow-go/engine/access/rpc/backend"
	"github.com/onflow/flow-go/engine/access/rpc/backend/events"
	"github.com/onflow/flow-go/engine/access/rpc/backend/node_communicator"
	"github.com/onflow/flow-go/engine/access/rpc/backend/query_mode"
	"github.com/onflow/flow-go/engine/access/rpc/backend/scripts"
	"github.com/onflow/flow-go/engine/access/rpc/backend/transactions/error_message_provider"
	rpcConnection "github.com/onflow/flow-go/engine/access/rpc/connection"
	"github.com/onflow/flow-go/engine/access/state_stream"
	statestreambackend "github.com/onflow/flow-go/engine/access/state_stream/backend"
	"github.com/onflow/flow-go/engine/access/subscription"
	subscriptiontracker "github.com/onflow/flow-go/engine/access/subscription/tracker"
	followereng "github.com/onflow/flow-go/engine/common/follower"
	"github.com/onflow/flow-go/engine/common/requester"
	commonrpc "github.com/onflow/flow-go/engine/common/rpc"
	"github.com/onflow/flow-go/engine/common/stop"
	synceng "github.com/onflow/flow-go/engine/common/synchronization"
	"github.com/onflow/flow-go/engine/common/version"
	"github.com/onflow/flow-go/engine/execution/computation"
	"github.com/onflow/flow-go/engine/execution/computation/query"
	"github.com/onflow/flow-go/fvm/storage/derived"
	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/complete/wal"
	"github.com/onflow/flow-go/model/bootstrap"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/model/flow/filter"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/blobs"
	"github.com/onflow/flow-go/module/chainsync"
	"github.com/onflow/flow-go/module/counters"
	"github.com/onflow/flow-go/module/execution"
	"github.com/onflow/flow-go/module/executiondatasync/execution_data"
	execdatacache "github.com/onflow/flow-go/module/executiondatasync/execution_data/cache"
	"github.com/onflow/flow-go/module/executiondatasync/pruner"
	edstorage "github.com/onflow/flow-go/module/executiondatasync/storage"
	"github.com/onflow/flow-go/module/executiondatasync/tracker"
	finalizer "github.com/onflow/flow-go/module/finalizer/consensus"
	"github.com/onflow/flow-go/module/grpcserver"
	"github.com/onflow/flow-go/module/id"
	"github.com/onflow/flow-go/module/mempool/herocache"
	"github.com/onflow/flow-go/module/mempool/stdmap"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/module/metrics/unstaked"
	"github.com/onflow/flow-go/module/state_synchronization"
	"github.com/onflow/flow-go/module/state_synchronization/indexer"
	edrequester "github.com/onflow/flow-go/module/state_synchronization/requester"
	"github.com/onflow/flow-go/network"
	alspmgr "github.com/onflow/flow-go/network/alsp/manager"
	netcache "github.com/onflow/flow-go/network/cache"
	"github.com/onflow/flow-go/network/channels"
	cborcodec "github.com/onflow/flow-go/network/codec/cbor"
	"github.com/onflow/flow-go/network/p2p"
	"github.com/onflow/flow-go/network/p2p/blob"
	p2pbuilder "github.com/onflow/flow-go/network/p2p/builder"
	p2pbuilderconfig "github.com/onflow/flow-go/network/p2p/builder/config"
	"github.com/onflow/flow-go/network/p2p/cache"
	"github.com/onflow/flow-go/network/p2p/conduit"
	"github.com/onflow/flow-go/network/p2p/connection"
	"github.com/onflow/flow-go/network/p2p/dht"
	networkingsubscription "github.com/onflow/flow-go/network/p2p/subscription"
	"github.com/onflow/flow-go/network/p2p/translator"
	"github.com/onflow/flow-go/network/p2p/unicast/protocols"
	relaynet "github.com/onflow/flow-go/network/relay"
	"github.com/onflow/flow-go/network/slashing"
	"github.com/onflow/flow-go/network/topology"
	"github.com/onflow/flow-go/network/underlay"
	"github.com/onflow/flow-go/network/validator"
	"github.com/onflow/flow-go/state/protocol"
	badgerState "github.com/onflow/flow-go/state/protocol/badger"
	"github.com/onflow/flow-go/state/protocol/blocktimer"
	"github.com/onflow/flow-go/storage"
	bstorage "github.com/onflow/flow-go/storage/badger"
	pstorage "github.com/onflow/flow-go/storage/pebble"
	"github.com/onflow/flow-go/storage/store"
	"github.com/onflow/flow-go/utils/grpcutils"
)

// AccessNodeBuilder extends cmd.NodeBuilder and declares additional functions needed to bootstrap an Access node.
// The private network allows the staked nodes to communicate among themselves, while the public network allows the
// Observers and an Access node to communicate.
//
//                                 public network                           private network
//  +------------------------+
//  | Observer             1 |<--------------------------|
//  +------------------------+                           v
//  +------------------------+                         +--------------------+                 +------------------------+
//  | Observer             2 |<----------------------->| Staked Access Node |<--------------->| All other staked Nodes |
//  +------------------------+                         +--------------------+                 +------------------------+
//  +------------------------+                           ^
//  | Observer             3 |<--------------------------|
//  +------------------------+

// AccessNodeConfig defines all the user defined parameters required to bootstrap an access node
// For a node running as a standalone process, the config fields will be populated from the command line params,
// while for a node running as a library, the config fields are expected to be initialized by the caller.
type AccessNodeConfig struct {
	supportsObserver                     bool // True if this is an Access node that supports observers and consensus follower engines
	collectionGRPCPort                   uint
	executionGRPCPort                    uint
	pingEnabled                          bool
	nodeInfoFile                         string
	apiRatelimits                        map[string]int
	apiBurstlimits                       map[string]int
	rpcConf                              rpc.Config
	stateStreamConf                      statestreambackend.Config
	stateStreamFilterConf                map[string]int
	ExecutionNodeAddress                 string // deprecated
	HistoricalAccessRPCs                 []access.AccessAPIClient
	logTxTimeToFinalized                 bool
	logTxTimeToExecuted                  bool
	logTxTimeToFinalizedExecuted         bool
	logTxTimeToSealed                    bool
	retryEnabled                         bool
	rpcMetricsEnabled                    bool
	executionDataSyncEnabled             bool
	publicNetworkExecutionDataEnabled    bool
	executionDataDBMode                  string
	executionDataPrunerHeightRangeTarget uint64
	executionDataPrunerThreshold         uint64
	executionDataPruningInterval         time.Duration
	executionDataDir                     string
	executionDataStartHeight             uint64
	executionDataConfig                  edrequester.ExecutionDataConfig
	PublicNetworkConfig                  PublicNetworkConfig
	TxResultCacheSize                    uint
	executionDataIndexingEnabled         bool
	registersDBPath                      string
	checkpointFile                       string
	scriptExecutorConfig                 query.QueryConfig
	scriptExecMinBlock                   uint64
	scriptExecMaxBlock                   uint64
	registerCacheType                    string
	registerCacheSize                    uint
	programCacheSize                     uint
	checkPayerBalanceMode                string
	versionControlEnabled                bool
	storeTxResultErrorMessages           bool
	stopControlEnabled                   bool
	registerDBPruneThreshold             uint64
}

type PublicNetworkConfig struct {
	// NetworkKey crypto.PublicKey // TODO: do we need a different key for the public network?
	BindAddress string
	Network     network.EngineRegistry
	Metrics     module.NetworkMetrics
}

// DefaultAccessNodeConfig defines all the default values for the AccessNodeConfig
func DefaultAccessNodeConfig() *AccessNodeConfig {
	homedir, _ := os.UserHomeDir()
	return &AccessNodeConfig{
		supportsObserver:   false,
		collectionGRPCPort: 9000,
		executionGRPCPort:  9000,
		rpcConf: rpc.Config{
			UnsecureGRPCListenAddr: "0.0.0.0:9000",
			SecureGRPCListenAddr:   "0.0.0.0:9001",
			HTTPListenAddr:         "0.0.0.0:8000",
			CollectionAddr:         "",
			HistoricalAccessAddrs:  "",
			BackendConfig: backend.Config{
				CollectionClientTimeout:   3 * time.Second,
				ExecutionClientTimeout:    3 * time.Second,
				ConnectionPoolSize:        backend.DefaultConnectionPoolSize,
				MaxHeightRange:            events.DefaultMaxHeightRange,
				PreferredExecutionNodeIDs: nil,
				FixedExecutionNodeIDs:     nil,
				CircuitBreakerConfig: rpcConnection.CircuitBreakerConfig{
					Enabled:        false,
					RestoreTimeout: 60 * time.Second,
					MaxFailures:    5,
					MaxRequests:    1,
				},
				ScriptExecutionMode: query_mode.IndexQueryModeExecutionNodesOnly.String(), // default to ENs only for now
				EventQueryMode:      query_mode.IndexQueryModeExecutionNodesOnly.String(), // default to ENs only for now
				TxResultQueryMode:   query_mode.IndexQueryModeExecutionNodesOnly.String(), // default to ENs only for now
			},
			RestConfig: rest.Config{
				ListenAddress:  "",
				WriteTimeout:   rest.DefaultWriteTimeout,
				ReadTimeout:    rest.DefaultReadTimeout,
				IdleTimeout:    rest.DefaultIdleTimeout,
				MaxRequestSize: commonrest.DefaultMaxRequestSize,
			},
			MaxMsgSize:                grpcutils.DefaultMaxMsgSize,
			CompressorName:            grpcutils.NoCompressor,
			WebSocketConfig:           websockets.NewDefaultWebsocketConfig(),
			EnableWebSocketsStreamAPI: true,
		},
		stateStreamConf: statestreambackend.Config{
			MaxExecutionDataMsgSize: grpcutils.DefaultMaxMsgSize,
			ExecutionDataCacheSize:  subscription.DefaultCacheSize,
			ClientSendTimeout:       subscription.DefaultSendTimeout,
			ClientSendBufferSize:    subscription.DefaultSendBufferSize,
			MaxGlobalStreams:        subscription.DefaultMaxGlobalStreams,
			EventFilterConfig:       state_stream.DefaultEventFilterConfig,
			RegisterIDsRequestLimit: state_stream.DefaultRegisterIDsRequestLimit,
			ResponseLimit:           subscription.DefaultResponseLimit,
			HeartbeatInterval:       subscription.DefaultHeartbeatInterval,
		},
		stateStreamFilterConf:        nil,
		ExecutionNodeAddress:         "localhost:9000",
		logTxTimeToFinalized:         false,
		logTxTimeToExecuted:          false,
		logTxTimeToFinalizedExecuted: false,
		logTxTimeToSealed:            false,
		pingEnabled:                  false,
		retryEnabled:                 false,
		rpcMetricsEnabled:            false,
		nodeInfoFile:                 "",
		apiRatelimits:                nil,
		apiBurstlimits:               nil,
		TxResultCacheSize:            0,
		PublicNetworkConfig: PublicNetworkConfig{
			BindAddress: cmd.NotSet,
			Metrics:     metrics.NewNoopCollector(),
		},
		executionDataSyncEnabled:          true,
		publicNetworkExecutionDataEnabled: false,
		executionDataDir:                  filepath.Join(homedir, ".flow", "execution_data"),
		executionDataStartHeight:          0,
		executionDataConfig: edrequester.ExecutionDataConfig{
			InitialBlockHeight: 0,
			MaxSearchAhead:     edrequester.DefaultMaxSearchAhead,
			FetchTimeout:       edrequester.DefaultFetchTimeout,
			MaxFetchTimeout:    edrequester.DefaultMaxFetchTimeout,
			RetryDelay:         edrequester.DefaultRetryDelay,
			MaxRetryDelay:      edrequester.DefaultMaxRetryDelay,
		},
		executionDataIndexingEnabled:         false,
		executionDataDBMode:                  execution_data.ExecutionDataDBModeBadger.String(),
		executionDataPrunerHeightRangeTarget: 0,
		executionDataPrunerThreshold:         pruner.DefaultThreshold,
		executionDataPruningInterval:         pruner.DefaultPruningInterval,
		registersDBPath:                      filepath.Join(homedir, ".flow", "execution_state"),
		checkpointFile:                       cmd.NotSet,
		scriptExecutorConfig:                 query.NewDefaultConfig(),
		scriptExecMinBlock:                   0,
		scriptExecMaxBlock:                   math.MaxUint64,
		registerCacheType:                    pstorage.CacheTypeTwoQueue.String(),
		registerCacheSize:                    0,
		programCacheSize:                     0,
		checkPayerBalanceMode:                txvalidator.Disabled.String(),
		versionControlEnabled:                true,
		storeTxResultErrorMessages:           false,
		stopControlEnabled:                   false,
		registerDBPruneThreshold:             0,
	}
}

// FlowAccessNodeBuilder provides the common functionality needed to bootstrap a Flow access node
// It is composed of the FlowNodeBuilder, the AccessNodeConfig and contains all the components and modules needed for the
// access nodes
type FlowAccessNodeBuilder struct {
	*cmd.FlowNodeBuilder
	*AccessNodeConfig

	// components
	FollowerState                protocol.FollowerState
	SyncCore                     *chainsync.Core
	RpcEng                       *rpc.Engine
	FollowerDistributor          *consensuspubsub.FollowerDistributor
	CollectionRPC                access.AccessAPIClient
	TransactionTimings           *stdmap.TransactionTimings
	CollectionsToMarkFinalized   *stdmap.Times
	CollectionsToMarkExecuted    *stdmap.Times
	BlocksToMarkExecuted         *stdmap.Times
	BlockTransactions            *stdmap.IdentifierMap
	TransactionMetrics           *metrics.TransactionCollector
	TransactionValidationMetrics *metrics.TransactionValidationCollector
	RestMetrics                  *metrics.RestCollector
	AccessMetrics                module.AccessMetrics
	PingMetrics                  module.PingMetrics
	Committee                    hotstuff.DynamicCommittee
	Finalized                    *flow.Header // latest finalized block that the node knows of at startup time
	Pending                      []*flow.Header
	FollowerCore                 module.HotStuffFollower
	Validator                    hotstuff.Validator
	ExecutionDataDownloader      execution_data.Downloader
	PublicBlobService            network.BlobService
	ExecutionDataRequester       state_synchronization.ExecutionDataRequester
	ExecutionDataStore           execution_data.ExecutionDataStore
	ExecutionDataBlobstore       blobs.Blobstore
	ExecutionDataCache           *execdatacache.ExecutionDataCache
	ExecutionIndexer             *indexer.Indexer
	ExecutionIndexerCore         *indexer.IndexerCore
	ScriptExecutor               *scripts.ScriptExecutor
	RegistersAsyncStore          *execution.RegistersAsyncStore
	Reporter                     *index.Reporter
	EventsIndex                  *index.EventsIndex
	TxResultsIndex               *index.TransactionResultsIndex
	IndexerDependencies          *cmd.DependencyList
	collectionExecutedMetric     module.CollectionExecutedMetric
	ExecutionDataPruner          *pruner.Pruner
	ExecutionDatastoreManager    edstorage.DatastoreManager
	ExecutionDataTracker         tracker.Storage
	VersionControl               *version.VersionControl
	StopControl                  *stop.StopControl

	// storage
	events                         storage.Events
	lightTransactionResults        storage.LightTransactionResults
	transactionResultErrorMessages storage.TransactionResultErrorMessages
	transactions                   storage.Transactions
	collections                    storage.Collections

	// The sync engine participants provider is the libp2p peer store for the access node
	// which is not available until after the network has started.
	// Hence, a factory function that needs to be called just before creating the sync engine
	SyncEngineParticipantsProviderFactory func() module.IdentifierProvider

	// engines
	IngestEng      *ingestion.Engine
	RequestEng     *requester.Engine
	FollowerEng    *followereng.ComplianceEngine
	SyncEng        *synceng.Engine
	StateStreamEng *statestreambackend.Engine

	// grpc servers
	secureGrpcServer      *grpcserver.GrpcServer
	unsecureGrpcServer    *grpcserver.GrpcServer
	stateStreamGrpcServer *grpcserver.GrpcServer

	stateStreamBackend *statestreambackend.StateStreamBackend
	nodeBackend        *backend.Backend

	ExecNodeIdentitiesProvider   *commonrpc.ExecutionNodeIdentitiesProvider
	TxResultErrorMessagesCore    *tx_error_messages.TxErrorMessagesCore
	txResultErrorMessageProvider error_message_provider.TxErrorMessageProvider
}

func (builder *FlowAccessNodeBuilder) buildFollowerState() *FlowAccessNodeBuilder {
	builder.Module("mutable follower state", func(node *cmd.NodeConfig) error {
		// For now, we only support state implementations from package badger.
		// If we ever support different implementations, the following can be replaced by a type-aware factory
		state, ok := node.State.(*badgerState.State)
		if !ok {
			return fmt.Errorf("only implementations of type badger.State are currently supported but read-only state has type %T", node.State)
		}

		followerState, err := badgerState.NewFollowerState(
			node.Logger,
			node.Tracer,
			node.ProtocolEvents,
			state,
			node.Storage.Index,
			node.Storage.Payloads,
			blocktimer.DefaultBlockTimer,
		)
		builder.FollowerState = followerState

		return err
	})

	return builder
}

func (builder *FlowAccessNodeBuilder) buildSyncCore() *FlowAccessNodeBuilder {
	builder.Module("sync core", func(node *cmd.NodeConfig) error {
		syncCore, err := chainsync.New(node.Logger, node.SyncCoreConfig, metrics.NewChainSyncCollector(node.RootChainID), node.RootChainID)
		builder.SyncCore = syncCore

		return err
	})

	return builder
}

func (builder *FlowAccessNodeBuilder) buildCommittee() *FlowAccessNodeBuilder {
	builder.Component("committee", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		// initialize consensus committee's membership state
		// This committee state is for the HotStuff follower, which follows the MAIN CONSENSUS committee
		// Note: node.Me.NodeID() is not part of the consensus committee
		committee, err := committees.NewConsensusCommittee(node.State, node.Me.NodeID())
		node.ProtocolEvents.AddConsumer(committee)
		builder.Committee = committee

		return committee, err
	})

	return builder
}

func (builder *FlowAccessNodeBuilder) buildLatestHeader() *FlowAccessNodeBuilder {
	builder.Module("latest header", func(node *cmd.NodeConfig) error {
		finalized, pending, err := recovery.FindLatest(node.State, node.Storage.Headers)
		builder.Finalized, builder.Pending = finalized, pending

		return err
	})

	return builder
}

func (builder *FlowAccessNodeBuilder) buildFollowerCore() *FlowAccessNodeBuilder {
	builder.Component("follower core", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		// create a finalizer that will handle updating the protocol
		// state when the follower detects newly finalized blocks
		final := finalizer.NewFinalizer(node.DB, node.Storage.Headers, builder.FollowerState, node.Tracer)

		packer := signature.NewConsensusSigDataPacker(builder.Committee)
		// initialize the verifier for the protocol consensus
		verifier := verification.NewCombinedVerifier(builder.Committee, packer)
		builder.Validator = hotstuffvalidator.New(builder.Committee, verifier)

		followerCore, err := consensus.NewFollower(
			node.Logger,
			node.Metrics.Mempool,
			node.Storage.Headers,
			final,
			builder.FollowerDistributor,
			node.FinalizedRootBlock.Header,
			node.RootQC,
			builder.Finalized,
			builder.Pending,
		)
		if err != nil {
			return nil, fmt.Errorf("could not initialize follower core: %w", err)
		}
		builder.FollowerCore = followerCore

		return builder.FollowerCore, nil
	})

	return builder
}

func (builder *FlowAccessNodeBuilder) buildFollowerEngine() *FlowAccessNodeBuilder {
	builder.Component("follower engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		var heroCacheCollector module.HeroCacheMetrics = metrics.NewNoopCollector()
		if node.HeroCacheMetricsEnable {
			heroCacheCollector = metrics.FollowerCacheMetrics(node.MetricsRegisterer)
		}

		core, err := followereng.NewComplianceCore(
			node.Logger,
			node.Metrics.Mempool,
			heroCacheCollector,
			builder.FollowerDistributor,
			builder.FollowerState,
			builder.FollowerCore,
			builder.Validator,
			builder.SyncCore,
			node.Tracer,
		)
		if err != nil {
			return nil, fmt.Errorf("could not create follower core: %w", err)
		}

		builder.FollowerEng, err = followereng.NewComplianceLayer(
			node.Logger,
			node.EngineRegistry,
			node.Me,
			node.Metrics.Engine,
			node.Storage.Headers,
			builder.Finalized,
			core,
			node.ComplianceConfig,
		)
		if err != nil {
			return nil, fmt.Errorf("could not create follower engine: %w", err)
		}
		builder.FollowerDistributor.AddOnBlockFinalizedConsumer(builder.FollowerEng.OnFinalizedBlock)

		return builder.FollowerEng, nil
	})

	return builder
}

func (builder *FlowAccessNodeBuilder) buildSyncEngine() *FlowAccessNodeBuilder {
	builder.Component("sync engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		spamConfig, err := synceng.NewSpamDetectionConfig()
		if err != nil {
			return nil, fmt.Errorf("could not initialize spam detection config: %w", err)
		}
		sync, err := synceng.New(
			node.Logger,
			node.Metrics.Engine,
			node.EngineRegistry,
			node.Me,
			node.State,
			node.Storage.Blocks,
			builder.FollowerEng,
			builder.SyncCore,
			builder.SyncEngineParticipantsProviderFactory(),
			spamConfig,
		)
		if err != nil {
			return nil, fmt.Errorf("could not create synchronization engine: %w", err)
		}
		builder.SyncEng = sync
		builder.FollowerDistributor.AddFinalizationConsumer(sync)

		return builder.SyncEng, nil
	})

	return builder
}

func (builder *FlowAccessNodeBuilder) BuildConsensusFollower() *FlowAccessNodeBuilder {
	builder.
		buildFollowerState().
		buildSyncCore().
		buildCommittee().
		buildLatestHeader().
		buildFollowerCore().
		buildFollowerEngine().
		buildSyncEngine()

	return builder
}

func (builder *FlowAccessNodeBuilder) BuildExecutionSyncComponents() *FlowAccessNodeBuilder {
	var bs network.BlobService
	var processedBlockHeight storage.ConsumerProgressInitializer
	var processedNotifications storage.ConsumerProgressInitializer
	var bsDependable *module.ProxiedReadyDoneAware
	var execDataDistributor *edrequester.ExecutionDataDistributor
	var execDataCacheBackend *herocache.BlockExecutionData
	var executionDataStoreCache *execdatacache.ExecutionDataCache

	// setup dependency chain to ensure indexer starts after the requester
	requesterDependable := module.NewProxiedReadyDoneAware()
	builder.IndexerDependencies.Add(requesterDependable)

	executionDataPrunerEnabled := builder.executionDataPrunerHeightRangeTarget != 0

	builder.
		AdminCommand("read-execution-data", func(config *cmd.NodeConfig) commands.AdminCommand {
			return stateSyncCommands.NewReadExecutionDataCommand(builder.ExecutionDataStore)
		}).
		Module("transactions and collections storage", func(node *cmd.NodeConfig) error {

			dbStore := cmd.GetStorageMultiDBStoreIfNeeded(node)

			transactions := store.NewTransactions(node.Metrics.Cache, dbStore)
			collections := store.NewCollections(dbStore, transactions)
			builder.transactions = transactions
			builder.collections = collections

			return nil
		}).
		Module("execution data datastore and blobstore", func(node *cmd.NodeConfig) error {
			var err error
			builder.ExecutionDatastoreManager, err = edstorage.CreateDatastoreManager(
				node.Logger, builder.executionDataDir, builder.executionDataDBMode)
			if err != nil {
				return fmt.Errorf("could not create execution data datastore manager: %w", err)
			}

			builder.ShutdownFunc(builder.ExecutionDatastoreManager.Close)

			return nil
		}).
		Module("processed block height consumer progress", func(node *cmd.NodeConfig) error {
			// Note: progress is stored in the datastore's DB since that is where the jobqueue
			// writes execution data to.
			db := builder.ExecutionDatastoreManager.DB()

			processedBlockHeight = store.NewConsumerProgress(db, module.ConsumeProgressExecutionDataRequesterBlockHeight)
			return nil
		}).
		Module("processed notifications consumer progress", func(node *cmd.NodeConfig) error {
			// Note: progress is stored in the datastore's DB since that is where the jobqueue
			// writes execution data to.
			db := builder.ExecutionDatastoreManager.DB()
			processedNotifications = store.NewConsumerProgress(db, module.ConsumeProgressExecutionDataRequesterNotification)
			return nil
		}).
		Module("blobservice peer manager dependencies", func(node *cmd.NodeConfig) error {
			bsDependable = module.NewProxiedReadyDoneAware()
			builder.PeerManagerDependencies.Add(bsDependable)
			return nil
		}).
		Module("execution datastore", func(node *cmd.NodeConfig) error {
			builder.ExecutionDataBlobstore = blobs.NewBlobstore(builder.ExecutionDatastoreManager.Datastore())
			builder.ExecutionDataStore = execution_data.NewExecutionDataStore(builder.ExecutionDataBlobstore, execution_data.DefaultSerializer)
			return nil
		}).
		Module("execution data cache", func(node *cmd.NodeConfig) error {
			var heroCacheCollector module.HeroCacheMetrics = metrics.NewNoopCollector()
			if builder.HeroCacheMetricsEnable {
				heroCacheCollector = metrics.AccessNodeExecutionDataCacheMetrics(builder.MetricsRegisterer)
			}

			execDataCacheBackend = herocache.NewBlockExecutionData(builder.stateStreamConf.ExecutionDataCacheSize, builder.Logger, heroCacheCollector)

			// Execution Data cache that uses a blobstore as the backend (instead of a downloader)
			// This ensures that it simply returns a not found error if the blob doesn't exist
			// instead of attempting to download it from the network.
			executionDataStoreCache = execdatacache.NewExecutionDataCache(
				builder.ExecutionDataStore,
				builder.Storage.Headers,
				builder.Storage.Seals,
				builder.Storage.Results,
				execDataCacheBackend,
			)

			return nil
		}).
		Component("execution data service", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			opts := []network.BlobServiceOption{
				blob.WithBitswapOptions(
					// Only allow block requests from staked ENs and ANs
					bitswap.WithPeerBlockRequestFilter(
						blob.AuthorizedRequester(nil, builder.IdentityProvider, builder.Logger),
					),
					bitswap.WithTracer(
						blob.NewTracer(node.Logger.With().Str("blob_service", channels.ExecutionDataService.String()).Logger()),
					),
				),
			}

			if !builder.BitswapReprovideEnabled {
				opts = append(opts, blob.WithReprovideInterval(-1))
			}

			var err error
			bs, err = node.EngineRegistry.RegisterBlobService(channels.ExecutionDataService, builder.ExecutionDatastoreManager.Datastore(), opts...)
			if err != nil {
				return nil, fmt.Errorf("could not register blob service: %w", err)
			}

			// add blobservice into ReadyDoneAware dependency passed to peer manager
			// this starts the blob service and configures peer manager to wait for the blobservice
			// to be ready before starting
			bsDependable.Init(bs)

			var downloaderOpts []execution_data.DownloaderOption

			if executionDataPrunerEnabled {
				sealed, err := node.State.Sealed().Head()
				if err != nil {
					return nil, fmt.Errorf("cannot get the sealed block: %w", err)
				}

				trackerDir := filepath.Join(builder.executionDataDir, "tracker")
				builder.ExecutionDataTracker, err = tracker.OpenStorage(
					trackerDir,
					sealed.Height,
					node.Logger,
					tracker.WithPruneCallback(func(c cid.Cid) error {
						// TODO: use a proper context here
						return builder.ExecutionDataBlobstore.DeleteBlob(context.TODO(), c)
					}),
				)
				if err != nil {
					return nil, fmt.Errorf("failed to create execution data tracker: %w", err)
				}

				downloaderOpts = []execution_data.DownloaderOption{
					execution_data.WithExecutionDataTracker(builder.ExecutionDataTracker, node.Storage.Headers),
				}
			}

			builder.ExecutionDataDownloader = execution_data.NewDownloader(bs, downloaderOpts...)
			return builder.ExecutionDataDownloader, nil
		}).
		Component("execution data requester", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			// Validation of the start block height needs to be done after loading state
			if builder.executionDataStartHeight > 0 {
				if builder.executionDataStartHeight <= builder.FinalizedRootBlock.Header.Height {
					return nil, fmt.Errorf(
						"execution data start block height (%d) must be greater than the root block height (%d)",
						builder.executionDataStartHeight, builder.FinalizedRootBlock.Header.Height)
				}

				latestSeal, err := builder.State.Sealed().Head()
				if err != nil {
					return nil, fmt.Errorf("failed to get latest sealed height")
				}

				// Note: since the root block of a spork is also sealed in the root protocol state, the
				// latest sealed height is always equal to the root block height. That means that at the
				// very beginning of a spork, this check will always fail. Operators should not specify
				// an InitialBlockHeight when starting from the beginning of a spork.
				if builder.executionDataStartHeight > latestSeal.Height {
					return nil, fmt.Errorf(
						"execution data start block height (%d) must be less than or equal to the latest sealed block height (%d)",
						builder.executionDataStartHeight, latestSeal.Height)
				}

				// executionDataStartHeight is provided as the first block to sync, but the
				// requester expects the initial last processed height, which is the first height - 1
				builder.executionDataConfig.InitialBlockHeight = builder.executionDataStartHeight - 1
			} else {
				builder.executionDataConfig.InitialBlockHeight = builder.SealedRootBlock.Header.Height
			}

			execDataDistributor = edrequester.NewExecutionDataDistributor()

			// Execution Data cache with a downloader as the backend. This is used by the requester
			// to download and cache execution data for each block. It shares a cache backend instance
			// with the datastore implementation.
			executionDataCache := execdatacache.NewExecutionDataCache(
				builder.ExecutionDataDownloader,
				builder.Storage.Headers,
				builder.Storage.Seals,
				builder.Storage.Results,
				execDataCacheBackend,
			)

			r, err := edrequester.New(
				builder.Logger,
				metrics.NewExecutionDataRequesterCollector(),
				builder.ExecutionDataDownloader,
				executionDataCache,
				processedBlockHeight,
				processedNotifications,
				builder.State,
				builder.Storage.Headers,
				builder.executionDataConfig,
				execDataDistributor,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create execution data requester: %w", err)
			}
			builder.ExecutionDataRequester = r

			builder.FollowerDistributor.AddOnBlockFinalizedConsumer(builder.ExecutionDataRequester.OnBlockFinalized)

			// add requester into ReadyDoneAware dependency passed to indexer. This allows the indexer
			// to wait for the requester to be ready before starting.
			requesterDependable.Init(builder.ExecutionDataRequester)

			return builder.ExecutionDataRequester, nil
		}).
		Component("execution data pruner", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			if !executionDataPrunerEnabled {
				return &module.NoopReadyDoneAware{}, nil
			}

			var prunerMetrics module.ExecutionDataPrunerMetrics = metrics.NewNoopCollector()
			if node.MetricsEnabled {
				prunerMetrics = metrics.NewExecutionDataPrunerCollector()
			}

			var err error
			builder.ExecutionDataPruner, err = pruner.NewPruner(
				node.Logger,
				prunerMetrics,
				builder.ExecutionDataTracker,
				pruner.WithPruneCallback(func(ctx context.Context) error {
					return builder.ExecutionDatastoreManager.CollectGarbage(ctx)
				}),
				pruner.WithHeightRangeTarget(builder.executionDataPrunerHeightRangeTarget),
				pruner.WithThreshold(builder.executionDataPrunerThreshold),
				pruner.WithPruningInterval(builder.executionDataPruningInterval),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create execution data pruner: %w", err)
			}

			builder.ExecutionDataPruner.RegisterHeightRecorder(builder.ExecutionDataDownloader)

			return builder.ExecutionDataPruner, nil
		})

	if builder.publicNetworkExecutionDataEnabled {
		var publicBsDependable *module.ProxiedReadyDoneAware

		builder.Module("public blobservice peer manager dependencies", func(node *cmd.NodeConfig) error {
			publicBsDependable = module.NewProxiedReadyDoneAware()
			builder.PeerManagerDependencies.Add(publicBsDependable)
			return nil
		})
		builder.Component("public network execution data service", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			opts := []network.BlobServiceOption{
				blob.WithBitswapOptions(
					bitswap.WithTracer(
						blob.NewTracer(node.Logger.With().Str("public_blob_service", channels.PublicExecutionDataService.String()).Logger()),
					),
				),
				blob.WithParentBlobService(bs),
			}

			net := builder.AccessNodeConfig.PublicNetworkConfig.Network

			var err error
			builder.PublicBlobService, err = net.RegisterBlobService(channels.PublicExecutionDataService, builder.ExecutionDatastoreManager.Datastore(), opts...)
			if err != nil {
				return nil, fmt.Errorf("could not register blob service: %w", err)
			}

			// add blobservice into ReadyDoneAware dependency passed to peer manager
			// this starts the blob service and configures peer manager to wait for the blobservice
			// to be ready before starting
			publicBsDependable.Init(builder.PublicBlobService)
			return &module.NoopReadyDoneAware{}, nil
		})
	}

	if builder.executionDataIndexingEnabled {
		var indexedBlockHeight storage.ConsumerProgressInitializer

		builder.
			AdminCommand("execute-script", func(config *cmd.NodeConfig) commands.AdminCommand {
				return stateSyncCommands.NewExecuteScriptCommand(builder.ScriptExecutor)
			}).
			Module("indexed block height consumer progress", func(node *cmd.NodeConfig) error {
				// Note: progress is stored in the MAIN db since that is where indexed execution data is stored.
				indexedBlockHeight = store.NewConsumerProgress(builder.ProtocolDB, module.ConsumeProgressExecutionDataIndexerBlockHeight)
				return nil
			}).
			Module("transaction results storage", func(node *cmd.NodeConfig) error {
				dbStore := cmd.GetStorageMultiDBStoreIfNeeded(node)
				builder.lightTransactionResults = store.NewLightTransactionResults(node.Metrics.Cache, dbStore, bstorage.DefaultCacheSize)
				return nil
			}).
			DependableComponent("execution data indexer", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
				// Note: using a DependableComponent here to ensure that the indexer does not block
				// other components from starting while bootstrapping the register db since it may
				// take hours to complete.

				pdb, err := pstorage.OpenRegisterPebbleDB(
					node.Logger.With().Str("pebbledb", "registers").Logger(),
					builder.registersDBPath)
				if err != nil {
					return nil, fmt.Errorf("could not open registers db: %w", err)
				}
				builder.ShutdownFunc(func() error {
					return pdb.Close()
				})

				bootstrapped, err := pstorage.IsBootstrapped(pdb)
				if err != nil {
					return nil, fmt.Errorf("could not check if registers db is bootstrapped: %w", err)
				}

				if !bootstrapped {
					checkpointFile := builder.checkpointFile
					if checkpointFile == cmd.NotSet {
						checkpointFile = path.Join(builder.BootstrapDir, bootstrap.PathRootCheckpoint)
					}

					// currently, the checkpoint must be from the root block.
					// read the root hash from the provided checkpoint and verify it matches the
					// state commitment from the root snapshot.
					err := wal.CheckpointHasRootHash(
						node.Logger,
						"", // checkpoint file already full path
						checkpointFile,
						ledger.RootHash(node.RootSeal.FinalState),
					)
					if err != nil {
						return nil, fmt.Errorf("could not verify checkpoint file: %w", err)
					}

					checkpointHeight := builder.SealedRootBlock.Header.Height

					if builder.SealedRootBlock.ID() != builder.RootSeal.BlockID {
						return nil, fmt.Errorf("mismatching sealed root block and root seal: %v != %v",
							builder.SealedRootBlock.ID(), builder.RootSeal.BlockID)
					}

					rootHash := ledger.RootHash(builder.RootSeal.FinalState)
					bootstrap, err := pstorage.NewRegisterBootstrap(pdb, checkpointFile, checkpointHeight, rootHash, builder.Logger)
					if err != nil {
						return nil, fmt.Errorf("could not create registers bootstrap: %w", err)
					}

					// TODO: find a way to hook a context up to this to allow a graceful shutdown
					workerCount := 10
					err = bootstrap.IndexCheckpointFile(context.Background(), workerCount)
					if err != nil {
						return nil, fmt.Errorf("could not load checkpoint file: %w", err)
					}
				}

				registers, err := pstorage.NewRegisters(pdb, builder.registerDBPruneThreshold)
				if err != nil {
					return nil, fmt.Errorf("could not create registers storage: %w", err)
				}

				if builder.registerCacheSize > 0 {
					cacheType, err := pstorage.ParseCacheType(builder.registerCacheType)
					if err != nil {
						return nil, fmt.Errorf("could not parse register cache type: %w", err)
					}
					cacheMetrics := metrics.NewCacheCollector(builder.RootChainID)
					registersCache, err := pstorage.NewRegistersCache(registers, cacheType, builder.registerCacheSize, cacheMetrics)
					if err != nil {
						return nil, fmt.Errorf("could not create registers cache: %w", err)
					}
					builder.Storage.RegisterIndex = registersCache
				} else {
					builder.Storage.RegisterIndex = registers
				}

				indexerDerivedChainData, queryDerivedChainData, err := builder.buildDerivedChainData()
				if err != nil {
					return nil, fmt.Errorf("could not create derived chain data: %w", err)
				}

				indexerCore, err := indexer.New(
					builder.Logger,
					metrics.NewExecutionStateIndexerCollector(),
					notNil(builder.ProtocolDB),
					notNil(builder.Storage.RegisterIndex),
					notNil(builder.Storage.Headers),
					notNil(builder.events),
					notNil(builder.collections),
					notNil(builder.transactions),
					notNil(builder.lightTransactionResults),
					builder.RootChainID.Chain(),
					indexerDerivedChainData,
					notNil(builder.collectionExecutedMetric),
				)
				if err != nil {
					return nil, err
				}
				builder.ExecutionIndexerCore = indexerCore

				// execution state worker uses a jobqueue to process new execution data and indexes it by using the indexer.
				builder.ExecutionIndexer, err = indexer.NewIndexer(
					builder.Logger,
					registers.FirstHeight(),
					registers,
					indexerCore,
					executionDataStoreCache,
					builder.ExecutionDataRequester.HighestConsecutiveHeight,
					indexedBlockHeight,
				)
				if err != nil {
					return nil, err
				}

				if executionDataPrunerEnabled {
					builder.ExecutionDataPruner.RegisterHeightRecorder(builder.ExecutionIndexer)
				}

				// setup requester to notify indexer when new execution data is received
				execDataDistributor.AddOnExecutionDataReceivedConsumer(builder.ExecutionIndexer.OnExecutionData)

				// create script execution module, this depends on the indexer being initialized and the
				// having the register storage bootstrapped
				scripts := execution.NewScripts(
					builder.Logger,
					metrics.NewExecutionCollector(builder.Tracer),
					builder.RootChainID,
					computation.NewProtocolStateWrapper(builder.State),
					builder.Storage.Headers,
					builder.ExecutionIndexerCore.RegisterValue,
					builder.scriptExecutorConfig,
					queryDerivedChainData,
					builder.programCacheSize > 0,
				)

				err = builder.ScriptExecutor.Initialize(builder.ExecutionIndexer, scripts, builder.VersionControl)
				if err != nil {
					return nil, err
				}

				err = builder.Reporter.Initialize(builder.ExecutionIndexer)
				if err != nil {
					return nil, err
				}

				err = builder.RegistersAsyncStore.Initialize(registers)
				if err != nil {
					return nil, err
				}

				if builder.stopControlEnabled {
					builder.StopControl.RegisterHeightRecorder(builder.ExecutionIndexer)
				}

				return builder.ExecutionIndexer, nil
			}, builder.IndexerDependencies)
	}

	if builder.stateStreamConf.ListenAddr != "" {
		builder.Component("exec state stream engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			for key, value := range builder.stateStreamFilterConf {
				switch key {
				case "EventTypes":
					builder.stateStreamConf.MaxEventTypes = value
				case "Addresses":
					builder.stateStreamConf.MaxAddresses = value
				case "Contracts":
					builder.stateStreamConf.MaxContracts = value
				case "AccountAddresses":
					builder.stateStreamConf.MaxAccountAddress = value
				}
			}
			builder.stateStreamConf.RpcMetricsEnabled = builder.rpcMetricsEnabled

			highestAvailableHeight, err := builder.ExecutionDataRequester.HighestConsecutiveHeight()
			if err != nil {
				return nil, fmt.Errorf("could not get highest consecutive height: %w", err)
			}
			broadcaster := engine.NewBroadcaster()

			eventQueryMode, err := query_mode.ParseIndexQueryMode(builder.rpcConf.BackendConfig.EventQueryMode)
			if err != nil {
				return nil, fmt.Errorf("could not parse event query mode: %w", err)
			}

			// use the events index for events if enabled and the node is configured to use it for
			// regular event queries
			useIndex := builder.executionDataIndexingEnabled &&
				eventQueryMode != query_mode.IndexQueryModeExecutionNodesOnly

			executionDataTracker := subscriptiontracker.NewExecutionDataTracker(
				builder.Logger,
				node.State,
				builder.executionDataConfig.InitialBlockHeight,
				node.Storage.Headers,
				broadcaster,
				highestAvailableHeight,
				builder.EventsIndex,
				useIndex,
			)

			builder.stateStreamBackend, err = statestreambackend.New(
				node.Logger,
				node.State,
				node.Storage.Headers,
				node.Storage.Seals,
				node.Storage.Results,
				builder.ExecutionDataStore,
				executionDataStoreCache,
				builder.RegistersAsyncStore,
				builder.EventsIndex,
				useIndex,
				int(builder.stateStreamConf.RegisterIDsRequestLimit),
				subscription.NewSubscriptionHandler(
					builder.Logger,
					broadcaster,
					builder.stateStreamConf.ClientSendTimeout,
					builder.stateStreamConf.ResponseLimit,
					builder.stateStreamConf.ClientSendBufferSize,
				),
				executionDataTracker,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create state stream backend: %w", err)
			}

			stateStreamEng, err := statestreambackend.NewEng(
				node.Logger,
				builder.stateStreamConf,
				executionDataStoreCache,
				node.Storage.Headers,
				node.RootChainID,
				builder.stateStreamGrpcServer,
				builder.stateStreamBackend,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create state stream engine: %w", err)
			}
			builder.StateStreamEng = stateStreamEng

			// setup requester to notify ExecutionDataTracker when new execution data is received
			execDataDistributor.AddOnExecutionDataReceivedConsumer(builder.stateStreamBackend.OnExecutionData)

			return builder.StateStreamEng, nil
		})
	}

	return builder
}

// buildDerivedChainData creates the derived chain data for the indexer and the query engine
// If program caching is disabled, the function will return nil for the indexer cache, and a
// derived chain data object for the query engine cache.
func (builder *FlowAccessNodeBuilder) buildDerivedChainData() (
	indexerCache *derived.DerivedChainData,
	queryCache *derived.DerivedChainData,
	err error,
) {
	cacheSize := builder.programCacheSize

	// the underlying cache requires size > 0. no data will be written so 1 is fine.
	if cacheSize == 0 {
		cacheSize = 1
	}

	derivedChainData, err := derived.NewDerivedChainData(cacheSize)
	if err != nil {
		return nil, nil, err
	}

	// writes are done by the indexer. using a nil value effectively disables writes to the cache.
	if builder.programCacheSize == 0 {
		return nil, derivedChainData, nil
	}

	return derivedChainData, derivedChainData, nil
}

func FlowAccessNode(nodeBuilder *cmd.FlowNodeBuilder) *FlowAccessNodeBuilder {
	dist := consensuspubsub.NewFollowerDistributor()
	dist.AddProposalViolationConsumer(notifications.NewSlashingViolationsConsumer(nodeBuilder.Logger))
	return &FlowAccessNodeBuilder{
		AccessNodeConfig:    DefaultAccessNodeConfig(),
		FlowNodeBuilder:     nodeBuilder,
		FollowerDistributor: dist,
		IndexerDependencies: cmd.NewDependencyList(),
	}
}

func (builder *FlowAccessNodeBuilder) ParseFlags() error {
	builder.BaseFlags()

	builder.extraFlags()

	return builder.ParseAndPrintFlags()
}

func (builder *FlowAccessNodeBuilder) extraFlags() {
	builder.ExtraFlags(func(flags *pflag.FlagSet) {
		defaultConfig := DefaultAccessNodeConfig()

		flags.UintVar(&builder.collectionGRPCPort, "collection-ingress-port", defaultConfig.collectionGRPCPort, "the grpc ingress port for all collection nodes")
		flags.UintVar(&builder.executionGRPCPort, "execution-ingress-port", defaultConfig.executionGRPCPort, "the grpc ingress port for all execution nodes")
		flags.StringVarP(&builder.rpcConf.UnsecureGRPCListenAddr,
			"rpc-addr",
			"r",
			defaultConfig.rpcConf.UnsecureGRPCListenAddr,
			"the address the unsecured gRPC server listens on")
		flags.StringVar(&builder.rpcConf.SecureGRPCListenAddr,
			"secure-rpc-addr",
			defaultConfig.rpcConf.SecureGRPCListenAddr,
			"the address the secure gRPC server listens on")
		flags.StringVar(&builder.stateStreamConf.ListenAddr,
			"state-stream-addr",
			defaultConfig.stateStreamConf.ListenAddr,
			"the address the state stream server listens on (if empty the server will not be started)")
		flags.StringVarP(&builder.rpcConf.HTTPListenAddr, "http-addr", "h", defaultConfig.rpcConf.HTTPListenAddr, "the address the http proxy server listens on")
		flags.StringVar(&builder.rpcConf.RestConfig.ListenAddress,
			"rest-addr",
			defaultConfig.rpcConf.RestConfig.ListenAddress,
			"the address the REST server listens on (if empty the REST server will not be started)")
		flags.DurationVar(&builder.rpcConf.RestConfig.WriteTimeout,
			"rest-write-timeout",
			defaultConfig.rpcConf.RestConfig.WriteTimeout,
			"timeout to use when writing REST response")
		flags.DurationVar(&builder.rpcConf.RestConfig.ReadTimeout,
			"rest-read-timeout",
			defaultConfig.rpcConf.RestConfig.ReadTimeout,
			"timeout to use when reading REST request headers")
		flags.DurationVar(&builder.rpcConf.RestConfig.IdleTimeout, "rest-idle-timeout", defaultConfig.rpcConf.RestConfig.IdleTimeout, "idle timeout for REST connections")
		flags.Int64Var(&builder.rpcConf.RestConfig.MaxRequestSize,
			"rest-max-request-size",
			defaultConfig.rpcConf.RestConfig.MaxRequestSize,
			"the maximum request size in bytes for payload sent over REST server")
		flags.StringVarP(&builder.rpcConf.CollectionAddr,
			"static-collection-ingress-addr",
			"",
			defaultConfig.rpcConf.CollectionAddr,
			"the address (of the collection node) to send transactions to")
		flags.StringVarP(&builder.ExecutionNodeAddress,
			"script-addr",
			"s",
			defaultConfig.ExecutionNodeAddress,
			"the address (of the execution node) forward the script to")
		flags.StringVarP(&builder.rpcConf.HistoricalAccessAddrs,
			"historical-access-addr",
			"",
			defaultConfig.rpcConf.HistoricalAccessAddrs,
			"comma separated rpc addresses for historical access nodes")
		flags.DurationVar(&builder.rpcConf.BackendConfig.CollectionClientTimeout,
			"collection-client-timeout",
			defaultConfig.rpcConf.BackendConfig.CollectionClientTimeout,
			"grpc client timeout for a collection node")
		flags.DurationVar(&builder.rpcConf.BackendConfig.ExecutionClientTimeout,
			"execution-client-timeout",
			defaultConfig.rpcConf.BackendConfig.ExecutionClientTimeout,
			"grpc client timeout for an execution node")
		flags.UintVar(&builder.rpcConf.BackendConfig.ConnectionPoolSize,
			"connection-pool-size",
			defaultConfig.rpcConf.BackendConfig.ConnectionPoolSize,
			"maximum number of connections allowed in the connection pool, size of 0 disables the connection pooling, and anything less than the default size will be overridden to use the default size")
		flags.UintVar(&builder.rpcConf.MaxMsgSize,
			"rpc-max-message-size",
			grpcutils.DefaultMaxMsgSize,
			"the maximum message size in bytes for messages sent or received over grpc")
		flags.UintVar(&builder.rpcConf.BackendConfig.MaxHeightRange,
			"rpc-max-height-range",
			defaultConfig.rpcConf.BackendConfig.MaxHeightRange,
			"maximum size for height range requests")
		flags.StringSliceVar(&builder.rpcConf.BackendConfig.PreferredExecutionNodeIDs,
			"preferred-execution-node-ids",
			defaultConfig.rpcConf.BackendConfig.PreferredExecutionNodeIDs,
			"comma separated list of execution nodes ids to choose from when making an upstream call e.g. b4a4dbdcd443d...,fb386a6a... etc.")
		flags.StringSliceVar(&builder.rpcConf.BackendConfig.FixedExecutionNodeIDs,
			"fixed-execution-node-ids",
			defaultConfig.rpcConf.BackendConfig.FixedExecutionNodeIDs,
			"comma separated list of execution nodes ids to choose from when making an upstream call if no matching preferred execution id is found e.g. b4a4dbdcd443d...,fb386a6a... etc.")
		flags.StringVar(&builder.rpcConf.CompressorName,
			"grpc-compressor",
			defaultConfig.rpcConf.CompressorName,
			"name of grpc compressor that will be used for requests to other nodes. One of (gzip, snappy, deflate)")
		flags.BoolVar(&builder.logTxTimeToFinalized, "log-tx-time-to-finalized", defaultConfig.logTxTimeToFinalized, "log transaction time to finalized")
		flags.BoolVar(&builder.logTxTimeToExecuted, "log-tx-time-to-executed", defaultConfig.logTxTimeToExecuted, "log transaction time to executed")
		flags.BoolVar(&builder.logTxTimeToFinalizedExecuted,
			"log-tx-time-to-finalized-executed",
			defaultConfig.logTxTimeToFinalizedExecuted,
			"log transaction time to finalized and executed")
		flags.BoolVar(&builder.logTxTimeToSealed,
			"log-tx-time-to-sealed",
			defaultConfig.logTxTimeToSealed,
			"log transaction time to sealed")
		flags.BoolVar(&builder.pingEnabled,
			"ping-enabled",
			defaultConfig.pingEnabled,
			"whether to enable the ping process that pings all other peers and report the connectivity to metrics")
		flags.BoolVar(&builder.retryEnabled, "retry-enabled", defaultConfig.retryEnabled, "whether to enable the retry mechanism at the access node level")
		flags.BoolVar(&builder.rpcMetricsEnabled, "rpc-metrics-enabled", defaultConfig.rpcMetricsEnabled, "whether to enable the rpc metrics")
		flags.UintVar(&builder.TxResultCacheSize, "transaction-result-cache-size", defaultConfig.TxResultCacheSize, "transaction result cache size.(Disabled by default i.e 0)")
		flags.StringVarP(&builder.nodeInfoFile,
			"node-info-file",
			"",
			defaultConfig.nodeInfoFile,
			"full path to a json file which provides more details about nodes when reporting its reachability metrics")
		flags.StringToIntVar(&builder.apiRatelimits, "api-rate-limits", defaultConfig.apiRatelimits, "per second rate limits for Access API methods e.g. Ping=300,GetTransaction=500 etc.")
		flags.StringToIntVar(&builder.apiBurstlimits, "api-burst-limits", defaultConfig.apiBurstlimits, "burst limits for Access API methods e.g. Ping=100,GetTransaction=100 etc.")
		flags.BoolVar(&builder.supportsObserver, "supports-observer", defaultConfig.supportsObserver, "true if this staked access node supports observer or follower connections")
		flags.StringVar(&builder.PublicNetworkConfig.BindAddress, "public-network-address", defaultConfig.PublicNetworkConfig.BindAddress, "staked access node's public network bind address")
		flags.BoolVar(&builder.rpcConf.BackendConfig.CircuitBreakerConfig.Enabled,
			"circuit-breaker-enabled",
			defaultConfig.rpcConf.BackendConfig.CircuitBreakerConfig.Enabled,
			"specifies whether the circuit breaker is enabled for collection and execution API clients.")
		flags.DurationVar(&builder.rpcConf.BackendConfig.CircuitBreakerConfig.RestoreTimeout,
			"circuit-breaker-restore-timeout",
			defaultConfig.rpcConf.BackendConfig.CircuitBreakerConfig.RestoreTimeout,
			"duration after which the circuit breaker will restore the connection to the client after closing it due to failures. Default value is 60s")
		flags.Uint32Var(&builder.rpcConf.BackendConfig.CircuitBreakerConfig.MaxFailures,
			"circuit-breaker-max-failures",
			defaultConfig.rpcConf.BackendConfig.CircuitBreakerConfig.MaxFailures,
			"maximum number of failed calls to the client that will cause the circuit breaker to close the connection. Default value is 5")
		flags.Uint32Var(&builder.rpcConf.BackendConfig.CircuitBreakerConfig.MaxRequests,
			"circuit-breaker-max-requests",
			defaultConfig.rpcConf.BackendConfig.CircuitBreakerConfig.MaxRequests,
			"maximum number of requests to check if connection restored after timeout. Default value is 1")
		flags.BoolVar(&builder.versionControlEnabled,
			"version-control-enabled",
			defaultConfig.versionControlEnabled,
			"whether to enable the version control feature. Default value is true")
		flags.BoolVar(&builder.stopControlEnabled,
			"stop-control-enabled",
			defaultConfig.stopControlEnabled,
			"whether to enable the stop control feature. Default value is false")
		// ExecutionDataRequester config
		flags.BoolVar(&builder.executionDataSyncEnabled,
			"execution-data-sync-enabled",
			defaultConfig.executionDataSyncEnabled,
			"whether to enable the execution data sync protocol")
		flags.BoolVar(&builder.publicNetworkExecutionDataEnabled,
			"public-network-execution-data-sync-enabled",
			defaultConfig.publicNetworkExecutionDataEnabled,
			"[experimental] whether to enable the execution data sync protocol on public network")
		flags.StringVar(&builder.executionDataDir, "execution-data-dir", defaultConfig.executionDataDir, "directory to use for Execution Data database")
		flags.Uint64Var(&builder.executionDataStartHeight,
			"execution-data-start-height",
			defaultConfig.executionDataStartHeight,
			"height of first block to sync execution data from when starting with an empty Execution Data database")
		flags.Uint64Var(&builder.executionDataConfig.MaxSearchAhead,
			"execution-data-max-search-ahead",
			defaultConfig.executionDataConfig.MaxSearchAhead,
			"max number of heights to search ahead of the lowest outstanding execution data height")
		flags.DurationVar(&builder.executionDataConfig.FetchTimeout,
			"execution-data-fetch-timeout",
			defaultConfig.executionDataConfig.FetchTimeout,
			"initial timeout to use when fetching execution data from the network. timeout increases using an incremental backoff until execution-data-max-fetch-timeout. e.g. 30s")
		flags.DurationVar(&builder.executionDataConfig.MaxFetchTimeout,
			"execution-data-max-fetch-timeout",
			defaultConfig.executionDataConfig.MaxFetchTimeout,
			"maximum timeout to use when fetching execution data from the network e.g. 300s")
		flags.DurationVar(&builder.executionDataConfig.RetryDelay,
			"execution-data-retry-delay",
			defaultConfig.executionDataConfig.RetryDelay,
			"initial delay for exponential backoff when fetching execution data fails e.g. 10s")
		flags.DurationVar(&builder.executionDataConfig.MaxRetryDelay,
			"execution-data-max-retry-delay",
			defaultConfig.executionDataConfig.MaxRetryDelay,
			"maximum delay for exponential backoff when fetching execution data fails e.g. 5m")
		flags.StringVar(&builder.executionDataDBMode,
			"execution-data-db",
			defaultConfig.executionDataDBMode,
			"[experimental] the DB type for execution datastore. One of [badger, pebble]")
		flags.Uint64Var(&builder.executionDataPrunerHeightRangeTarget,
			"execution-data-height-range-target",
			defaultConfig.executionDataPrunerHeightRangeTarget,
			"number of blocks of Execution Data to keep on disk. older data is pruned")
		flags.Uint64Var(&builder.executionDataPrunerThreshold,
			"execution-data-height-range-threshold",
			defaultConfig.executionDataPrunerThreshold,
			"number of unpruned blocks of Execution Data beyond the height range target to allow before pruning")
		flags.DurationVar(&builder.executionDataPruningInterval,
			"execution-data-pruning-interval",
			defaultConfig.executionDataPruningInterval,
			"duration after which the pruner tries to prune execution data. The default value is 10 minutes")

		// Execution State Streaming API
		flags.Uint32Var(&builder.stateStreamConf.ExecutionDataCacheSize, "execution-data-cache-size", defaultConfig.stateStreamConf.ExecutionDataCacheSize, "block execution data cache size")
		flags.Uint32Var(&builder.stateStreamConf.MaxGlobalStreams, "state-stream-global-max-streams", defaultConfig.stateStreamConf.MaxGlobalStreams, "global maximum number of concurrent streams")
		flags.UintVar(&builder.stateStreamConf.MaxExecutionDataMsgSize,
			"state-stream-max-message-size",
			defaultConfig.stateStreamConf.MaxExecutionDataMsgSize,
			"maximum size for a gRPC message containing block execution data")
		flags.StringToIntVar(&builder.stateStreamFilterConf,
			"state-stream-event-filter-limits",
			defaultConfig.stateStreamFilterConf,
			"event filter limits for ExecutionData SubscribeEvents API e.g. EventTypes=100,Addresses=100,Contracts=100 etc.")
		flags.DurationVar(&builder.stateStreamConf.ClientSendTimeout,
			"state-stream-send-timeout",
			defaultConfig.stateStreamConf.ClientSendTimeout,
			"maximum wait before timing out while sending a response to a streaming client e.g. 30s")
		flags.UintVar(&builder.stateStreamConf.ClientSendBufferSize,
			"state-stream-send-buffer-size",
			defaultConfig.stateStreamConf.ClientSendBufferSize,
			"maximum number of responses to buffer within a stream")
		flags.Float64Var(&builder.stateStreamConf.ResponseLimit,
			"state-stream-response-limit",
			defaultConfig.stateStreamConf.ResponseLimit,
			"max number of responses per second to send over streaming endpoints. this helps manage resources consumed by each client querying data not in the cache e.g. 3 or 0.5. 0 means no limit")
		flags.Uint64Var(&builder.stateStreamConf.HeartbeatInterval,
			"state-stream-heartbeat-interval",
			defaultConfig.stateStreamConf.HeartbeatInterval,
			"default interval in blocks at which heartbeat messages should be sent. applied when client did not specify a value.")
		flags.Uint32Var(&builder.stateStreamConf.RegisterIDsRequestLimit,
			"state-stream-max-register-values",
			defaultConfig.stateStreamConf.RegisterIDsRequestLimit,
			"maximum number of register ids to include in a single request to the GetRegisters endpoint")

		// Execution Data Indexer
		flags.BoolVar(&builder.executionDataIndexingEnabled,
			"execution-data-indexing-enabled",
			defaultConfig.executionDataIndexingEnabled,
			"whether to enable the execution data indexing")
		flags.StringVar(&builder.registersDBPath, "execution-state-dir", defaultConfig.registersDBPath, "directory to use for execution-state database")
		flags.StringVar(&builder.checkpointFile, "execution-state-checkpoint", defaultConfig.checkpointFile, "execution-state checkpoint file")

		flags.StringVar(&builder.rpcConf.BackendConfig.EventQueryMode,
			"event-query-mode",
			defaultConfig.rpcConf.BackendConfig.EventQueryMode,
			"mode to use when querying events. one of [local-only, execution-nodes-only(default), failover]")

		flags.StringVar(&builder.rpcConf.BackendConfig.TxResultQueryMode,
			"tx-result-query-mode",
			defaultConfig.rpcConf.BackendConfig.TxResultQueryMode,
			"mode to use when querying transaction results. one of [local-only, execution-nodes-only(default), failover]")
		flags.BoolVar(&builder.storeTxResultErrorMessages,
			"store-tx-result-error-messages",
			defaultConfig.storeTxResultErrorMessages,
			"whether to enable storing transaction error messages into the db")
		// Script Execution
		flags.StringVar(&builder.rpcConf.BackendConfig.ScriptExecutionMode,
			"script-execution-mode",
			defaultConfig.rpcConf.BackendConfig.ScriptExecutionMode,
			"mode to use when executing scripts. one of (local-only, execution-nodes-only, failover, compare)")
		flags.Uint64Var(&builder.scriptExecutorConfig.ComputationLimit,
			"script-execution-computation-limit",
			defaultConfig.scriptExecutorConfig.ComputationLimit,
			"maximum number of computation units a locally executed script can use. default: 100000")
		flags.IntVar(&builder.scriptExecutorConfig.MaxErrorMessageSize,
			"script-execution-max-error-length",
			defaultConfig.scriptExecutorConfig.MaxErrorMessageSize,
			"maximum number characters to include in error message strings. additional characters are truncated. default: 1000")
		flags.DurationVar(&builder.scriptExecutorConfig.LogTimeThreshold,
			"script-execution-log-time-threshold",
			defaultConfig.scriptExecutorConfig.LogTimeThreshold,
			"emit a log for any scripts that take over this threshold. default: 1s")
		flags.DurationVar(&builder.scriptExecutorConfig.ExecutionTimeLimit,
			"script-execution-timeout",
			defaultConfig.scriptExecutorConfig.ExecutionTimeLimit,
			"timeout value for locally executed scripts. default: 10s")
		flags.Uint64Var(&builder.scriptExecMinBlock,
			"script-execution-min-height",
			defaultConfig.scriptExecMinBlock,
			"lowest block height to allow for script execution. default: no limit")
		flags.Uint64Var(&builder.scriptExecMaxBlock,
			"script-execution-max-height",
			defaultConfig.scriptExecMaxBlock,
			"highest block height to allow for script execution. default: no limit")
		flags.StringVar(&builder.registerCacheType,
			"register-cache-type",
			defaultConfig.registerCacheType,
			"type of backend cache to use for registers (lru, arc, 2q)")
		flags.UintVar(&builder.registerCacheSize,
			"register-cache-size",
			defaultConfig.registerCacheSize,
			"number of registers to cache for script execution. default: 0 (no cache)")
		flags.UintVar(&builder.programCacheSize,
			"program-cache-size",
			defaultConfig.programCacheSize,
			"[experimental] number of blocks to cache for cadence programs. use 0 to disable cache. default: 0. Note: this is an experimental feature and may cause nodes to become unstable under certain workloads. Use with caution.")

		// Payer Balance
		flags.StringVar(&builder.checkPayerBalanceMode,
			"check-payer-balance-mode",
			defaultConfig.checkPayerBalanceMode,
			"flag for payer balance validation that specifies whether or not to enforce the balance check. one of [disabled(default), warn, enforce]")

		// Register DB Pruning
		flags.Uint64Var(&builder.registerDBPruneThreshold,
			"registerdb-pruning-threshold",
			defaultConfig.registerDBPruneThreshold,
			fmt.Sprintf("specifies the number of blocks below the latest stored block height to keep in register db. default: %d", defaultConfig.registerDBPruneThreshold))

		// websockets config
		flags.DurationVar(
			&builder.rpcConf.WebSocketConfig.InactivityTimeout,
			"websocket-inactivity-timeout",
			defaultConfig.rpcConf.WebSocketConfig.InactivityTimeout,
			"the duration a WebSocket connection can remain open without any active subscriptions before being automatically closed",
		)
		flags.Uint64Var(
			&builder.rpcConf.WebSocketConfig.MaxSubscriptionsPerConnection,
			"websocket-max-subscriptions-per-connection",
			defaultConfig.rpcConf.WebSocketConfig.MaxSubscriptionsPerConnection,
			"the maximum number of active WebSocket subscriptions allowed per connection",
		)
		flags.Float64Var(
			&builder.rpcConf.WebSocketConfig.MaxResponsesPerSecond,
			"websocket-max-responses-per-second",
			defaultConfig.rpcConf.WebSocketConfig.MaxResponsesPerSecond,
			fmt.Sprintf("the maximum number of responses that can be sent to a single client per second. Default: %f. if set to 0, no limit is applied to the number of responses per second.", defaultConfig.rpcConf.WebSocketConfig.MaxResponsesPerSecond),
		)
		flags.BoolVar(
			&builder.rpcConf.EnableWebSocketsStreamAPI,
			"websockets-stream-api-enabled",
			defaultConfig.rpcConf.EnableWebSocketsStreamAPI,
			"whether to enable the WebSockets Stream API.",
		)
	}).ValidateFlags(func() error {
		if builder.supportsObserver && (builder.PublicNetworkConfig.BindAddress == cmd.NotSet || builder.PublicNetworkConfig.BindAddress == "") {
			return errors.New("public-network-address must be set if supports-observer is true")
		}
		if builder.executionDataSyncEnabled {
			if builder.executionDataConfig.FetchTimeout <= 0 {
				return errors.New("execution-data-fetch-timeout must be greater than 0")
			}
			if builder.executionDataConfig.MaxFetchTimeout < builder.executionDataConfig.FetchTimeout {
				return errors.New("execution-data-max-fetch-timeout must be greater than execution-data-fetch-timeout")
			}
			if builder.executionDataConfig.RetryDelay <= 0 {
				return errors.New("execution-data-retry-delay must be greater than 0")
			}
			if builder.executionDataConfig.MaxRetryDelay < builder.executionDataConfig.RetryDelay {
				return errors.New("execution-data-max-retry-delay must be greater than or equal to execution-data-retry-delay")
			}
			if builder.executionDataConfig.MaxSearchAhead == 0 {
				return errors.New("execution-data-max-search-ahead must be greater than 0")
			}
		}
		if builder.stateStreamConf.ListenAddr != "" {
			if builder.stateStreamConf.ExecutionDataCacheSize == 0 {
				return errors.New("execution-data-cache-size must be greater than 0")
			}
			if builder.stateStreamConf.ClientSendBufferSize == 0 {
				return errors.New("state-stream-send-buffer-size must be greater than 0")
			}
			if len(builder.stateStreamFilterConf) > 4 {
				return errors.New("state-stream-event-filter-limits must have at most 3 keys (EventTypes, Addresses, Contracts, AccountAddresses)")
			}
			for key, value := range builder.stateStreamFilterConf {
				switch key {
				case "EventTypes", "Addresses", "Contracts", "AccountAddresses":
					if value <= 0 {
						return fmt.Errorf("state-stream-event-filter-limits %s must be greater than 0", key)
					}
				default:
					return errors.New("state-stream-event-filter-limits may only contain the keys EventTypes, Addresses, Contracts, AccountAddresses")
				}
			}
			if builder.stateStreamConf.ResponseLimit < 0 {
				return errors.New("state-stream-response-limit must be greater than or equal to 0")
			}
			if builder.stateStreamConf.RegisterIDsRequestLimit <= 0 {
				return errors.New("state-stream-max-register-values must be greater than 0")
			}
		}
		if builder.rpcConf.BackendConfig.CircuitBreakerConfig.Enabled {
			if builder.rpcConf.BackendConfig.CircuitBreakerConfig.MaxFailures == 0 {
				return errors.New("circuit-breaker-max-failures must be greater than 0")
			}
			if builder.rpcConf.BackendConfig.CircuitBreakerConfig.MaxRequests == 0 {
				return errors.New("circuit-breaker-max-requests must be greater than 0")
			}
			if builder.rpcConf.BackendConfig.CircuitBreakerConfig.RestoreTimeout <= 0 {
				return errors.New("circuit-breaker-restore-timeout must be greater than 0")
			}
		}

		if builder.checkPayerBalanceMode != txvalidator.Disabled.String() && !builder.executionDataIndexingEnabled {
			return errors.New("execution-data-indexing-enabled must be set if check-payer-balance is enabled")
		}

		if builder.rpcConf.RestConfig.MaxRequestSize <= 0 {
			return errors.New("rest-max-request-size must be greater than 0")
		}

		return nil
	})
}

func publicNetworkMsgValidators(log zerolog.Logger, idProvider module.IdentityProvider, selfID flow.Identifier) []network.MessageValidator {
	return []network.MessageValidator{
		// filter out messages sent by this node itself
		validator.ValidateNotSender(selfID),
		validator.NewAnyValidator(
			// message should be either from a valid staked node
			validator.NewOriginValidator(
				id.NewIdentityFilterIdentifierProvider(filter.IsValidCurrentEpochParticipant, idProvider),
			),
			// or the message should be specifically targeted for this node
			validator.ValidateTarget(log, selfID),
		),
	}
}

func (builder *FlowAccessNodeBuilder) InitIDProviders() {
	builder.Module("id providers", func(node *cmd.NodeConfig) error {
		idCache, err := cache.NewProtocolStateIDCache(node.Logger, node.State, node.ProtocolEvents)
		if err != nil {
			return fmt.Errorf("could not initialize ProtocolStateIDCache: %w", err)
		}
		builder.IDTranslator = translator.NewHierarchicalIDTranslator(idCache, translator.NewPublicNetworkIDTranslator())

		// The following wrapper allows to disallow-list byzantine nodes via an admin command:
		// the wrapper overrides the 'Ejected' flag of disallow-listed nodes to true
		disallowListWrapper, err := cache.NewNodeDisallowListWrapper(
			idCache,
			node.ProtocolDB,
			func() network.DisallowListNotificationConsumer {
				return builder.NetworkUnderlay
			},
		)
		if err != nil {
			return fmt.Errorf("could not initialize NodeDisallowListWrapper: %w", err)
		}
		builder.IdentityProvider = disallowListWrapper

		// register the wrapper for dynamic configuration via admin command
		err = node.ConfigManager.RegisterIdentifierListConfig("network-id-provider-blocklist",
			disallowListWrapper.GetDisallowList, disallowListWrapper.Update)
		if err != nil {
			return fmt.Errorf("failed to register disallow-list wrapper with config manager: %w", err)
		}

		builder.SyncEngineParticipantsProviderFactory = func() module.IdentifierProvider {
			return id.NewIdentityFilterIdentifierProvider(
				filter.And(
					filter.HasRole[flow.Identity](flow.RoleConsensus),
					filter.Not(filter.HasNodeID[flow.Identity](node.Me.NodeID())),
					filter.NotEjectedFilter,
				),
				builder.IdentityProvider,
			)
		}
		return nil
	})
}

func (builder *FlowAccessNodeBuilder) Initialize() error {
	builder.InitIDProviders()

	builder.EnqueueResolver()

	// enqueue the regular network
	builder.EnqueueNetworkInit()

	builder.AdminCommand("get-transactions", func(conf *cmd.NodeConfig) commands.AdminCommand {
		return storageCommands.NewGetTransactionsCommand(conf.State, conf.Storage.Payloads, notNil(builder.collections))
	})

	// if this is an access node that supports public followers, enqueue the public network
	if builder.supportsObserver {
		builder.enqueuePublicNetworkInit()
		builder.enqueueRelayNetwork()
	}

	builder.EnqueuePingService()

	builder.EnqueueMetricsServerInit()

	if err := builder.RegisterBadgerMetrics(); err != nil {
		return err
	}

	builder.EnqueueTracer()
	builder.PreInit(cmd.DynamicStartPreInit)
	builder.ValidateRootSnapshot(badgerState.ValidRootSnapshotContainsEntityExpiryRange)

	return nil
}

func (builder *FlowAccessNodeBuilder) enqueueRelayNetwork() {
	builder.Component("relay network", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		relayNet := relaynet.NewRelayNetwork(
			node.EngineRegistry,
			builder.AccessNodeConfig.PublicNetworkConfig.Network,
			node.Logger,
			map[channels.Channel]channels.Channel{
				channels.ReceiveBlocks: channels.PublicReceiveBlocks,
			},
		)
		node.EngineRegistry = relayNet
		return relayNet, nil
	})
}

func (builder *FlowAccessNodeBuilder) Build() (cmd.Node, error) {
	var processedFinalizedBlockHeight storage.ConsumerProgressInitializer
	var processedTxErrorMessagesBlockHeight storage.ConsumerProgressInitializer

	if builder.executionDataSyncEnabled {
		builder.BuildExecutionSyncComponents()
	}

	ingestionDependable := module.NewProxiedReadyDoneAware()
	builder.IndexerDependencies.Add(ingestionDependable)
	versionControlDependable := module.NewProxiedReadyDoneAware()
	builder.IndexerDependencies.Add(versionControlDependable)
	stopControlDependable := module.NewProxiedReadyDoneAware()
	builder.IndexerDependencies.Add(stopControlDependable)
	var lastFullBlockHeight *counters.PersistentStrictMonotonicCounter

	builder.
		BuildConsensusFollower().
		Module("collection node client", func(node *cmd.NodeConfig) error {
			// collection node address is optional (if not specified, collection nodes will be chosen at random)
			if strings.TrimSpace(builder.rpcConf.CollectionAddr) == "" {
				node.Logger.Info().Msg("using a dynamic collection node address")
				return nil
			}

			node.Logger.Info().
				Str("collection_node", builder.rpcConf.CollectionAddr).
				Msg("using the static collection node address")

			collectionRPCConn, err := grpc.Dial(
				builder.rpcConf.CollectionAddr,
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(builder.rpcConf.MaxMsgSize))),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				rpcConnection.WithClientTimeoutOption(builder.rpcConf.BackendConfig.CollectionClientTimeout))
			if err != nil {
				return err
			}
			builder.CollectionRPC = access.NewAccessAPIClient(collectionRPCConn)
			return nil
		}).
		Module("historical access node clients", func(node *cmd.NodeConfig) error {
			addrs := strings.Split(builder.rpcConf.HistoricalAccessAddrs, ",")
			for _, addr := range addrs {
				if strings.TrimSpace(addr) == "" {
					continue
				}
				node.Logger.Info().Str("access_nodes", addr).Msg("historical access node addresses")

				historicalAccessRPCConn, err := grpc.Dial(
					addr,
					grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(builder.rpcConf.MaxMsgSize))),
					grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return err
				}
				builder.HistoricalAccessRPCs = append(builder.HistoricalAccessRPCs, access.NewAccessAPIClient(historicalAccessRPCConn))
			}
			return nil
		}).
		Module("transaction timing mempools", func(node *cmd.NodeConfig) error {
			var err error
			builder.TransactionTimings, err = stdmap.NewTransactionTimings(1500 * 300) // assume 1500 TPS * 300 seconds
			if err != nil {
				return err
			}

			builder.CollectionsToMarkFinalized, err = stdmap.NewTimes(50 * 300) // assume 50 collection nodes * 300 seconds
			if err != nil {
				return err
			}

			builder.CollectionsToMarkExecuted, err = stdmap.NewTimes(50 * 300) // assume 50 collection nodes * 300 seconds
			if err != nil {
				return err
			}

			builder.BlockTransactions, err = stdmap.NewIdentifierMap(10000)
			if err != nil {
				return err
			}

			builder.BlocksToMarkExecuted, err = stdmap.NewTimes(1 * 300) // assume 1 block per second * 300 seconds

			return err
		}).
		Module("transaction metrics", func(node *cmd.NodeConfig) error {
			builder.TransactionMetrics = metrics.NewTransactionCollector(
				node.Logger,
				builder.TransactionTimings,
				builder.logTxTimeToFinalized,
				builder.logTxTimeToExecuted,
				builder.logTxTimeToFinalizedExecuted,
				builder.logTxTimeToSealed,
			)
			return nil
		}).
		Module("transaction validation metrics", func(node *cmd.NodeConfig) error {
			builder.TransactionValidationMetrics = metrics.NewTransactionValidationCollector()
			return nil
		}).
		Module("rest metrics", func(node *cmd.NodeConfig) error {
			m, err := metrics.NewRestCollector(router.URLToRoute, node.MetricsRegisterer)
			if err != nil {
				return err
			}
			builder.RestMetrics = m
			return nil
		}).
		Module("access metrics", func(node *cmd.NodeConfig) error {
			builder.AccessMetrics = metrics.NewAccessCollector(
				metrics.WithTransactionMetrics(builder.TransactionMetrics),
				metrics.WithTransactionValidationMetrics(builder.TransactionValidationMetrics),
				metrics.WithBackendScriptsMetrics(builder.TransactionMetrics),
				metrics.WithRestMetrics(builder.RestMetrics),
			)
			return nil
		}).
		Module("collection metrics", func(node *cmd.NodeConfig) error {
			var err error
			builder.collectionExecutedMetric, err = indexer.NewCollectionExecutedMetricImpl(
				builder.Logger,
				builder.AccessMetrics,
				builder.CollectionsToMarkFinalized,
				builder.CollectionsToMarkExecuted,
				builder.BlocksToMarkExecuted,
				builder.collections,
				builder.Storage.Blocks,
				builder.BlockTransactions,
			)
			if err != nil {
				return err
			}

			return nil
		}).
		Module("ping metrics", func(node *cmd.NodeConfig) error {
			builder.PingMetrics = metrics.NewPingCollector()
			return nil
		}).
		Module("server certificate", func(node *cmd.NodeConfig) error {
			// generate the server certificate that will be served by the GRPC server
			x509Certificate, err := grpcutils.X509Certificate(node.NetworkKey)
			if err != nil {
				return err
			}
			tlsConfig := grpcutils.DefaultServerTLSConfig(x509Certificate)
			builder.rpcConf.TransportCredentials = credentials.NewTLS(tlsConfig)
			return nil
		}).
		Module("creating grpc servers", func(node *cmd.NodeConfig) error {
			builder.secureGrpcServer = grpcserver.NewGrpcServerBuilder(
				node.Logger,
				builder.rpcConf.SecureGRPCListenAddr,
				builder.rpcConf.MaxMsgSize,
				builder.rpcMetricsEnabled,
				builder.apiRatelimits,
				builder.apiBurstlimits,
				grpcserver.WithTransportCredentials(builder.rpcConf.TransportCredentials)).Build()

			builder.stateStreamGrpcServer = grpcserver.NewGrpcServerBuilder(
				node.Logger,
				builder.stateStreamConf.ListenAddr,
				builder.stateStreamConf.MaxExecutionDataMsgSize,
				builder.rpcMetricsEnabled,
				builder.apiRatelimits,
				builder.apiBurstlimits,
				grpcserver.WithStreamInterceptor()).Build()

			if builder.rpcConf.UnsecureGRPCListenAddr != builder.stateStreamConf.ListenAddr {
				builder.unsecureGrpcServer = grpcserver.NewGrpcServerBuilder(node.Logger,
					builder.rpcConf.UnsecureGRPCListenAddr,
					builder.rpcConf.MaxMsgSize,
					builder.rpcMetricsEnabled,
					builder.apiRatelimits,
					builder.apiBurstlimits).Build()
			} else {
				builder.unsecureGrpcServer = builder.stateStreamGrpcServer
			}

			return nil
		}).
		Module("backend script executor", func(node *cmd.NodeConfig) error {
			builder.ScriptExecutor = scripts.NewScriptExecutor(builder.Logger, builder.scriptExecMinBlock, builder.scriptExecMaxBlock)
			return nil
		}).
		Module("async register store", func(node *cmd.NodeConfig) error {
			builder.RegistersAsyncStore = execution.NewRegistersAsyncStore()
			return nil
		}).
		Module("events storage", func(node *cmd.NodeConfig) error {
			dbStore := cmd.GetStorageMultiDBStoreIfNeeded(node)
			builder.events = store.NewEvents(node.Metrics.Cache, dbStore)
			return nil
		}).
		Module("reporter", func(node *cmd.NodeConfig) error {
			builder.Reporter = index.NewReporter()
			return nil
		}).
		Module("events index", func(node *cmd.NodeConfig) error {
			builder.EventsIndex = index.NewEventsIndex(builder.Reporter, builder.events)
			return nil
		}).
		Module("transaction result index", func(node *cmd.NodeConfig) error {
			builder.TxResultsIndex = index.NewTransactionResultsIndex(builder.Reporter, builder.lightTransactionResults)
			return nil
		}).
		Module("processed finalized block height consumer progress", func(node *cmd.NodeConfig) error {
			processedFinalizedBlockHeight = store.NewConsumerProgress(builder.ProtocolDB, module.ConsumeProgressIngestionEngineBlockHeight)
			return nil
		}).
		Module("processed last full block height monotonic consumer progress", func(node *cmd.NodeConfig) error {
			rootBlockHeight := node.State.Params().FinalizedRoot().Height

			progress, err := store.NewConsumerProgress(builder.ProtocolDB, module.ConsumeProgressLastFullBlockHeight).Initialize(rootBlockHeight)
			if err != nil {
				return err
			}

			lastFullBlockHeight, err = counters.NewPersistentStrictMonotonicCounter(progress)
			if err != nil {
				return fmt.Errorf("failed to initialize monotonic consumer progress: %w", err)
			}

			return nil
		}).
		Module("transaction result error messages storage", func(node *cmd.NodeConfig) error {
			if builder.storeTxResultErrorMessages {
				dbStore := cmd.GetStorageMultiDBStoreIfNeeded(node)
				builder.transactionResultErrorMessages = store.NewTransactionResultErrorMessages(node.Metrics.Cache, dbStore, bstorage.DefaultCacheSize)
			}

			return nil
		}).
		Component("version control", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			if !builder.versionControlEnabled {
				noop := &module.NoopReadyDoneAware{}
				versionControlDependable.Init(noop)
				return noop, nil
			}

			nodeVersion, err := build.Semver()
			if err != nil {
				return nil, fmt.Errorf("could not load node version for version control. "+
					"version (%s) is not semver compliant: %w. Make sure a valid semantic version is provided in the VERSION environment variable", build.Version(), err)
			}

			versionControl, err := version.NewVersionControl(
				builder.Logger,
				node.Storage.VersionBeacons,
				nodeVersion,
				builder.SealedRootBlock.Header.Height,
				builder.LastFinalizedHeader.Height,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create version control: %w", err)
			}

			// VersionControl needs to consume BlockFinalized events.
			node.ProtocolEvents.AddConsumer(versionControl)

			builder.VersionControl = versionControl
			versionControlDependable.Init(builder.VersionControl)

			return versionControl, nil
		}).
		Component("stop control", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			if !builder.stopControlEnabled {
				noop := &module.NoopReadyDoneAware{}
				stopControlDependable.Init(noop)
				return noop, nil
			}

			stopControl := stop.NewStopControl(
				builder.Logger,
			)

			builder.VersionControl.AddVersionUpdatesConsumer(stopControl.OnVersionUpdate)

			builder.StopControl = stopControl
			stopControlDependable.Init(builder.StopControl)

			return stopControl, nil
		}).
		Component("RPC engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			config := builder.rpcConf
			backendConfig := config.BackendConfig
			accessMetrics := builder.AccessMetrics
			cacheSize := int(backendConfig.ConnectionPoolSize)

			var connBackendCache *rpcConnection.Cache
			var err error
			if cacheSize > 0 {
				connBackendCache, err = rpcConnection.NewCache(node.Logger, accessMetrics, cacheSize)
				if err != nil {
					return nil, fmt.Errorf("could not initialize connection cache: %w", err)
				}
			}

			connFactory := &rpcConnection.ConnectionFactoryImpl{
				CollectionGRPCPort:        builder.collectionGRPCPort,
				ExecutionGRPCPort:         builder.executionGRPCPort,
				CollectionNodeGRPCTimeout: backendConfig.CollectionClientTimeout,
				ExecutionNodeGRPCTimeout:  backendConfig.ExecutionClientTimeout,
				AccessMetrics:             accessMetrics,
				Log:                       node.Logger,
				Manager: rpcConnection.NewManager(
					node.Logger,
					accessMetrics,
					connBackendCache,
					config.MaxMsgSize,
					backendConfig.CircuitBreakerConfig,
					config.CompressorName,
				),
			}

			scriptExecMode, err := query_mode.ParseIndexQueryMode(config.BackendConfig.ScriptExecutionMode)
			if err != nil {
				return nil, fmt.Errorf("could not parse script execution mode: %w", err)
			}

			eventQueryMode, err := query_mode.ParseIndexQueryMode(config.BackendConfig.EventQueryMode)
			if err != nil {
				return nil, fmt.Errorf("could not parse event query mode: %w", err)
			}
			if eventQueryMode == query_mode.IndexQueryModeCompare {
				return nil, fmt.Errorf("event query mode 'compare' is not supported")
			}

			broadcaster := engine.NewBroadcaster()
			// create BlockTracker that will track for new blocks (finalized and sealed) and
			// handles block-related operations.
			blockTracker, err := subscriptiontracker.NewBlockTracker(
				node.State,
				builder.FinalizedRootBlock.Header.Height,
				node.Storage.Headers,
				broadcaster,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize block tracker: %w", err)
			}
			txResultQueryMode, err := query_mode.ParseIndexQueryMode(config.BackendConfig.TxResultQueryMode)
			if err != nil {
				return nil, fmt.Errorf("could not parse transaction result query mode: %w", err)
			}
			if txResultQueryMode == query_mode.IndexQueryModeCompare {
				return nil, fmt.Errorf("transaction result query mode 'compare' is not supported")
			}

			// If execution data syncing and indexing is disabled, pass nil indexReporter
			var indexReporter state_synchronization.IndexReporter
			if builder.executionDataSyncEnabled && builder.executionDataIndexingEnabled {
				indexReporter = builder.Reporter
			}

			checkPayerBalanceMode, err := txvalidator.ParsePayerBalanceMode(builder.checkPayerBalanceMode)
			if err != nil {
				return nil, fmt.Errorf("could not parse payer balance mode: %w", err)

			}

			preferredENIdentifiers, err := flow.IdentifierListFromHex(backendConfig.PreferredExecutionNodeIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to convert node id string to Flow Identifier for preferred EN map: %w", err)
			}

			fixedENIdentifiers, err := flow.IdentifierListFromHex(backendConfig.FixedExecutionNodeIDs)
			if err != nil {
				return nil, fmt.Errorf("failed to convert node id string to Flow Identifier for fixed EN map: %w", err)
			}

			builder.ExecNodeIdentitiesProvider = commonrpc.NewExecutionNodeIdentitiesProvider(
				node.Logger,
				node.State,
				node.Storage.Receipts,
				preferredENIdentifiers,
				fixedENIdentifiers,
			)

			nodeCommunicator := node_communicator.NewNodeCommunicator(backendConfig.CircuitBreakerConfig.Enabled)
			builder.txResultErrorMessageProvider = error_message_provider.NewTxErrorMessageProvider(
				node.Logger,
				builder.transactionResultErrorMessages, // might be nil
				notNil(builder.TxResultsIndex),
				connFactory,
				nodeCommunicator,
				notNil(builder.ExecNodeIdentitiesProvider),
			)

			builder.nodeBackend, err = backend.New(backend.Params{
				State:                 node.State,
				CollectionRPC:         builder.CollectionRPC, // might be nil
				HistoricalAccessNodes: notNil(builder.HistoricalAccessRPCs),
				Blocks:                node.Storage.Blocks,
				Headers:               node.Storage.Headers,
				Collections:           notNil(builder.collections),
				Transactions:          notNil(builder.transactions),
				ExecutionReceipts:     node.Storage.Receipts,
				ExecutionResults:      node.Storage.Results,
				TxResultErrorMessages: builder.transactionResultErrorMessages, // might be nil
				ChainID:               node.RootChainID,
				AccessMetrics:         notNil(builder.AccessMetrics),
				ConnFactory:           connFactory,
				RetryEnabled:          builder.retryEnabled,
				MaxHeightRange:        backendConfig.MaxHeightRange,
				Log:                   node.Logger,
				SnapshotHistoryLimit:  backend.DefaultSnapshotHistoryLimit,
				Communicator:          nodeCommunicator,
				TxResultCacheSize:     builder.TxResultCacheSize,
				ScriptExecutor:        notNil(builder.ScriptExecutor),
				ScriptExecutionMode:   scriptExecMode,
				CheckPayerBalanceMode: checkPayerBalanceMode,
				EventQueryMode:        eventQueryMode,
				BlockTracker:          blockTracker,
				SubscriptionHandler: subscription.NewSubscriptionHandler(
					builder.Logger,
					broadcaster,
					builder.stateStreamConf.ClientSendTimeout,
					builder.stateStreamConf.ResponseLimit,
					builder.stateStreamConf.ClientSendBufferSize,
				),
				EventsIndex:                notNil(builder.EventsIndex),
				TxResultQueryMode:          txResultQueryMode,
				TxResultsIndex:             notNil(builder.TxResultsIndex),
				LastFullBlockHeight:        lastFullBlockHeight,
				IndexReporter:              indexReporter,
				VersionControl:             notNil(builder.VersionControl),
				ExecNodeIdentitiesProvider: notNil(builder.ExecNodeIdentitiesProvider),
				TxErrorMessageProvider:     notNil(builder.txResultErrorMessageProvider),
			})
			if err != nil {
				return nil, fmt.Errorf("could not initialize backend: %w", err)
			}

			engineBuilder, err := rpc.NewBuilder(
				node.Logger,
				node.State,
				config,
				node.RootChainID,
				notNil(builder.AccessMetrics),
				builder.rpcMetricsEnabled,
				notNil(builder.Me),
				notNil(builder.nodeBackend),
				notNil(builder.nodeBackend),
				notNil(builder.secureGrpcServer),
				notNil(builder.unsecureGrpcServer),
				notNil(builder.stateStreamBackend),
				builder.stateStreamConf,
				indexReporter,
			)
			if err != nil {
				return nil, err
			}

			builder.RpcEng, err = engineBuilder.
				WithLegacy().
				WithBlockSignerDecoder(signature.NewBlockSignerDecoder(builder.Committee)).
				Build()
			if err != nil {
				return nil, err
			}
			builder.FollowerDistributor.AddOnBlockFinalizedConsumer(builder.RpcEng.OnFinalizedBlock)

			return builder.RpcEng, nil
		}).
		Component("ingestion engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			var err error

			builder.RequestEng, err = requester.New(
				node.Logger.With().Str("entity", "collection").Logger(),
				node.Metrics.Engine,
				node.EngineRegistry,
				node.Me,
				node.State,
				channels.RequestCollections,
				filter.HasRole[flow.Identity](flow.RoleCollection),
				func() flow.Entity { return &flow.Collection{} },
			)
			if err != nil {
				return nil, fmt.Errorf("could not create requester engine: %w", err)
			}

			if builder.storeTxResultErrorMessages {
				builder.TxResultErrorMessagesCore = tx_error_messages.NewTxErrorMessagesCore(
					node.Logger,
					notNil(builder.txResultErrorMessageProvider),
					builder.transactionResultErrorMessages,
					notNil(builder.ExecNodeIdentitiesProvider),
				)
			}

			collectionSyncer := ingestion.NewCollectionSyncer(
				node.Logger,
				notNil(builder.collectionExecutedMetric),
				builder.RequestEng,
				node.State,
				node.Storage.Blocks,
				notNil(builder.collections),
				notNil(builder.transactions),
				lastFullBlockHeight,
			)
			builder.RequestEng.WithHandle(collectionSyncer.OnCollectionDownloaded)

			builder.IngestEng, err = ingestion.New(
				node.Logger,
				node.EngineRegistry,
				node.State,
				node.Me,
				node.Storage.Blocks,
				node.Storage.Results,
				node.Storage.Receipts,
				processedFinalizedBlockHeight,
				notNil(collectionSyncer),
				notNil(builder.collectionExecutedMetric),
				notNil(builder.TxResultErrorMessagesCore),
			)
			if err != nil {
				return nil, err
			}
			ingestionDependable.Init(builder.IngestEng)
			builder.FollowerDistributor.AddOnBlockFinalizedConsumer(builder.IngestEng.OnFinalizedBlock)

			return builder.IngestEng, nil
		}).
		Component("requester engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			// We initialize the requester engine inside the ingestion engine due to the mutual dependency. However, in
			// order for it to properly start and shut down, we should still return it as its own engine here, so it can
			// be handled by the scaffold.
			return builder.RequestEng, nil
		}).
		AdminCommand("backfill-tx-error-messages", func(config *cmd.NodeConfig) commands.AdminCommand {
			return storageCommands.NewBackfillTxErrorMessagesCommand(
				builder.State,
				builder.TxResultErrorMessagesCore,
			)
		})

	if builder.storeTxResultErrorMessages {
		builder.Module("processed error messages block height consumer progress", func(node *cmd.NodeConfig) error {
			processedTxErrorMessagesBlockHeight = store.NewConsumerProgress(
				builder.ProtocolDB,
				module.ConsumeProgressEngineTxErrorMessagesBlockHeight,
			)
			return nil
		})
		builder.Component("transaction result error messages engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			engine, err := tx_error_messages.New(
				node.Logger,
				node.State,
				node.Storage.Headers,
				processedTxErrorMessagesBlockHeight,
				builder.TxResultErrorMessagesCore,
			)
			if err != nil {
				return nil, err
			}
			builder.FollowerDistributor.AddOnBlockFinalizedConsumer(engine.OnFinalizedBlock)

			return engine, nil
		})
	}

	if builder.supportsObserver {
		builder.Component("public sync request handler", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			syncRequestHandler, err := synceng.NewRequestHandlerEngine(
				node.Logger.With().Bool("public", true).Logger(),
				unstaked.NewUnstakedEngineCollector(node.Metrics.Engine),
				builder.AccessNodeConfig.PublicNetworkConfig.Network,
				node.Me,
				node.State,
				node.Storage.Blocks,
				builder.SyncCore,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create public sync request handler: %w", err)
			}
			builder.FollowerDistributor.AddFinalizationConsumer(syncRequestHandler)

			return syncRequestHandler, nil
		})
	}

	builder.Component("secure grpc server", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		return builder.secureGrpcServer, nil
	})

	builder.Component("state stream unsecure grpc server", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
		return builder.stateStreamGrpcServer, nil
	})

	if builder.rpcConf.UnsecureGRPCListenAddr != builder.stateStreamConf.ListenAddr {
		builder.Component("unsecure grpc server", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			return builder.unsecureGrpcServer, nil
		})
	}

	if builder.pingEnabled {
		builder.Component("ping engine", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			ping, err := pingeng.New(
				node.Logger,
				node.IdentityProvider,
				node.IDTranslator,
				node.Me,
				builder.PingMetrics,
				builder.nodeInfoFile,
				node.PingService,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create ping engine: %w", err)
			}

			return ping, nil
		})
	}

	return builder.FlowNodeBuilder.Build()
}

// enqueuePublicNetworkInit enqueues the public network component initialized for the staked node
func (builder *FlowAccessNodeBuilder) enqueuePublicNetworkInit() {
	var publicLibp2pNode p2p.LibP2PNode
	builder.
		Module("public network metrics", func(node *cmd.NodeConfig) error {
			builder.PublicNetworkConfig.Metrics = metrics.NewNetworkCollector(builder.Logger, metrics.WithNetworkPrefix("public"))
			return nil
		}).
		Component("public libp2p node", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			var err error
			publicLibp2pNode, err = builder.initPublicLibp2pNode(
				builder.NodeConfig.NetworkKey,
				builder.PublicNetworkConfig.BindAddress,
				builder.PublicNetworkConfig.Metrics)
			if err != nil {
				return nil, fmt.Errorf("could not create public libp2p node: %w", err)
			}

			return publicLibp2pNode, nil
		}).
		Component("public network", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			msgValidators := publicNetworkMsgValidators(node.Logger.With().Bool("public", true).Logger(), node.IdentityProvider, builder.NodeID)
			receiveCache := netcache.NewHeroReceiveCache(builder.FlowConfig.NetworkConfig.NetworkReceivedMessageCacheSize,
				builder.Logger,
				metrics.NetworkReceiveCacheMetricsFactory(builder.HeroCacheMetricsFactory(), network.PublicNetwork))

			err := node.Metrics.Mempool.Register(metrics.PrependPublicPrefix(metrics.ResourceNetworkingReceiveCache), receiveCache.Size)
			if err != nil {
				return nil, fmt.Errorf("could not register networking receive cache metric: %w", err)
			}

			net, err := underlay.NewNetwork(&underlay.NetworkConfig{
				Logger:                builder.Logger.With().Str("module", "public-network").Logger(),
				Libp2pNode:            publicLibp2pNode,
				Codec:                 cborcodec.NewCodec(),
				Me:                    builder.Me,
				Topology:              topology.EmptyTopology{}, // topology returns empty list since peers are not known upfront
				Metrics:               builder.PublicNetworkConfig.Metrics,
				BitSwapMetrics:        builder.Metrics.Bitswap,
				IdentityProvider:      builder.IdentityProvider,
				ReceiveCache:          receiveCache,
				ConduitFactory:        conduit.NewDefaultConduitFactory(),
				SporkId:               builder.SporkID,
				UnicastMessageTimeout: underlay.DefaultUnicastTimeout,
				IdentityTranslator:    builder.IDTranslator,
				AlspCfg: &alspmgr.MisbehaviorReportManagerConfig{
					Logger:                  builder.Logger,
					SpamRecordCacheSize:     builder.FlowConfig.NetworkConfig.AlspConfig.SpamRecordCacheSize,
					SpamReportQueueSize:     builder.FlowConfig.NetworkConfig.AlspConfig.SpamReportQueueSize,
					DisablePenalty:          builder.FlowConfig.NetworkConfig.AlspConfig.DisablePenalty,
					HeartBeatInterval:       builder.FlowConfig.NetworkConfig.AlspConfig.HearBeatInterval,
					AlspMetrics:             builder.Metrics.Network,
					NetworkType:             network.PublicNetwork,
					HeroCacheMetricsFactory: builder.HeroCacheMetricsFactory(),
				},
				SlashingViolationConsumerFactory: func(adapter network.ConduitAdapter) network.ViolationsConsumer {
					return slashing.NewSlashingViolationsConsumer(builder.Logger, builder.Metrics.Network, adapter)
				},
			}, underlay.WithMessageValidators(msgValidators...))
			if err != nil {
				return nil, fmt.Errorf("could not initialize network: %w", err)
			}

			builder.NetworkUnderlay = net
			builder.AccessNodeConfig.PublicNetworkConfig.Network = net

			node.Logger.Info().Msgf("network will run on address: %s", builder.PublicNetworkConfig.BindAddress)
			return net, nil
		}).
		Component("public peer manager", func(node *cmd.NodeConfig) (module.ReadyDoneAware, error) {
			return publicLibp2pNode.PeerManagerComponent(), nil
		})
}

// initPublicLibp2pNode initializes the public libp2p node for the public (unstaked) network.
// The LibP2P host is created with the following options:
//   - DHT as server
//   - The address from the node config or the specified bind address as the listen address
//   - The passed in private key as the libp2p key
//   - No connection gater
//   - Default Flow libp2p pubsub options
//
// Args:
//   - networkKey: The private key to use for the libp2p node
//
// - bindAddress: The address to bind the libp2p node to.
// - networkMetrics: The metrics collector for the network
// Returns:
// - The libp2p node instance for the public network.
// - Any error encountered during initialization. Any error should be considered fatal.
func (builder *FlowAccessNodeBuilder) initPublicLibp2pNode(networkKey crypto.PrivateKey, bindAddress string, networkMetrics module.LibP2PMetrics) (p2p.LibP2PNode,
	error,
) {
	connManager, err := connection.NewConnManager(builder.Logger, networkMetrics, &builder.FlowConfig.NetworkConfig.ConnectionManager)
	if err != nil {
		return nil, fmt.Errorf("could not create connection manager: %w", err)
	}
	libp2pNode, err := p2pbuilder.NewNodeBuilder(builder.Logger, &builder.FlowConfig.NetworkConfig.GossipSub, &p2pbuilderconfig.MetricsConfig{
		HeroCacheFactory: builder.HeroCacheMetricsFactory(),
		Metrics:          networkMetrics,
	},
		network.PublicNetwork,
		bindAddress,
		networkKey,
		builder.SporkID,
		builder.IdentityProvider,
		&builder.FlowConfig.NetworkConfig.ResourceManager,
		&p2pbuilderconfig.PeerManagerConfig{
			// TODO: eventually, we need pruning enabled even on public network. However, it needs a modified version of
			// the peer manager that also operate on the public identities.
			ConnectionPruning: connection.PruningDisabled,
			UpdateInterval:    builder.FlowConfig.NetworkConfig.PeerUpdateInterval,
			ConnectorFactory:  connection.DefaultLibp2pBackoffConnectorFactory(),
		},
		&p2p.DisallowListCacheConfig{
			MaxSize: builder.FlowConfig.NetworkConfig.DisallowListNotificationCacheSize,
			Metrics: metrics.DisallowListCacheMetricsFactory(builder.HeroCacheMetricsFactory(), network.PublicNetwork),
		},
		&p2pbuilderconfig.UnicastConfig{
			Unicast: builder.FlowConfig.NetworkConfig.Unicast,
		}).
		SetProtocolPeerCacheList(protocols.FlowProtocolID(builder.SporkID)).
		SetBasicResolver(builder.Resolver).
		SetSubscriptionFilter(networkingsubscription.NewRoleBasedFilter(flow.RoleAccess, builder.IdentityProvider)).
		SetConnectionManager(connManager).
		SetRoutingSystem(func(ctx context.Context, h host.Host) (routing.Routing, error) {
			return dht.NewDHT(ctx, h, protocols.FlowPublicDHTProtocolID(builder.SporkID), builder.Logger, networkMetrics, dht.AsServer())
		}).
		Build()
	if err != nil {
		return nil, fmt.Errorf("could not build libp2p node for staked access node: %w", err)
	}

	return libp2pNode, nil
}

// notNil ensures that the input is not nil and returns it
// the usage is to ensure the dependencies are initialized before initializing a module.
// for instance, the IngestionEngine depends on storage.Collections, which is initialized in a
// different function, so we need to ensure that the storage.Collections is initialized before
// creating the IngestionEngine.
func notNil[T any](dep T) T {
	if any(dep) == nil {
		panic("dependency is nil")
	}
	return dep
}
