package scripts

import (
	"context"
	"crypto/md5" //nolint:gosec
	"fmt"
	"testing"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	execproto "github.com/onflow/flow/protobuf/go/flow/execution"

	access "github.com/onflow/flow-go/engine/access/mock"
	"github.com/onflow/flow-go/engine/access/rpc/backend/common"
	"github.com/onflow/flow-go/engine/access/rpc/backend/node_communicator"
	"github.com/onflow/flow-go/engine/access/rpc/backend/query_mode"
	connectionmock "github.com/onflow/flow-go/engine/access/rpc/connection/mock"
	commonrpc "github.com/onflow/flow-go/engine/common/rpc"
	fvmerrors "github.com/onflow/flow-go/fvm/errors"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/execution"
	execmock "github.com/onflow/flow-go/module/execution/mock"
	"github.com/onflow/flow-go/module/irrecoverable"
	"github.com/onflow/flow-go/module/metrics"
	protocol "github.com/onflow/flow-go/state/protocol/mock"
	"github.com/onflow/flow-go/storage"
	storagemock "github.com/onflow/flow-go/storage/mock"
	"github.com/onflow/flow-go/utils/unittest"
	"github.com/onflow/flow-go/utils/unittest/mocks"
)

var (
	expectedResponse = []byte("response_data")

	cadenceErr    = fvmerrors.NewCodedError(fvmerrors.ErrCodeCadenceRunTimeError, "cadence error")
	fvmFailureErr = fvmerrors.NewCodedFailure(fvmerrors.FailureCodeBlockFinderFailure, "fvm error")
	ctxCancelErr  = fvmerrors.NewCodedError(fvmerrors.ErrCodeScriptExecutionCancelledError, "context canceled error")
	timeoutErr    = fvmerrors.NewCodedError(fvmerrors.ErrCodeScriptExecutionTimedOutError, "timeout error")
	compLimitErr  = fvmerrors.NewCodedError(fvmerrors.ErrCodeComputationLimitExceededError, "computation limit exceeded error")
	memLimitErr   = fvmerrors.NewCodedError(fvmerrors.ErrCodeMemoryLimitExceededError, "memory limit exceeded error")
)

// Create a suite similar to GetAccount that covers each of the modes
type BackendScriptsSuite struct {
	suite.Suite

	log        zerolog.Logger
	state      *protocol.State
	snapshot   *protocol.Snapshot
	params     *protocol.Params
	rootHeader *flow.Header

	headers           *storagemock.Headers
	receipts          *storagemock.ExecutionReceipts
	connectionFactory *connectionmock.ConnectionFactory
	chainID           flow.ChainID

	executionNodes flow.IdentityList
	execClient     *access.ExecutionAPIClient

	block *flow.Block

	script        []byte
	arguments     [][]byte
	failingScript []byte
}

func TestBackendScriptsSuite(t *testing.T) {
	suite.Run(t, new(BackendScriptsSuite))
}

func (s *BackendScriptsSuite) SetupTest() {
	s.log = unittest.Logger()
	s.state = protocol.NewState(s.T())
	s.snapshot = protocol.NewSnapshot(s.T())
	s.rootHeader = unittest.BlockHeaderFixture()
	s.params = protocol.NewParams(s.T())
	s.headers = storagemock.NewHeaders(s.T())
	s.receipts = storagemock.NewExecutionReceipts(s.T())
	s.connectionFactory = connectionmock.NewConnectionFactory(s.T())
	s.chainID = flow.Testnet

	s.execClient = access.NewExecutionAPIClient(s.T())
	s.executionNodes = unittest.IdentityListFixture(2, unittest.WithRole(flow.RoleExecution))

	block := unittest.BlockFixture()
	s.block = &block

	s.script = []byte("access(all) fun main() { return 1 }")
	s.arguments = [][]byte{[]byte("arg1"), []byte("arg2")}
	s.failingScript = []byte("access(all) fun main() { panic(\"!!\") }")
}

func (s *BackendScriptsSuite) defaultBackend(executor execution.ScriptExecutor, mode query_mode.IndexQueryMode) *Scripts {
	loggedScripts, err := lru.New[[md5.Size]byte, time.Time](common.DefaultLoggedScriptsCacheSize)
	s.Require().NoError(err)

	scripts, err := NewScriptsBackend(
		s.log,
		metrics.NewNoopCollector(),
		s.headers,
		s.state,
		s.connectionFactory,
		node_communicator.NewNodeCommunicator(false),
		executor,
		mode,
		commonrpc.NewExecutionNodeIdentitiesProvider(
			s.log,
			s.state,
			s.receipts,
			flow.IdentifierList{},
			flow.IdentifierList{},
		),
		loggedScripts,
	)
	require.NoError(s.T(), err)

	return scripts
}

// setupExecutionNodes sets up the mocks required to test against an EN backend
func (s *BackendScriptsSuite) setupExecutionNodes(block *flow.Block) {
	s.params.On("FinalizedRoot").Return(s.rootHeader, nil)
	s.state.On("Params").Return(s.params)
	s.state.On("Final").Return(s.snapshot)
	s.snapshot.On("Identities", mock.Anything).Return(s.executionNodes, nil)

	// this line causes a S1021 lint error because receipts is explicitly declared. this is required
	// to ensure the mock library handles the response type correctly
	var receipts flow.ExecutionReceiptList //nolint:gosimple
	receipts = unittest.ReceiptsForBlockFixture(block, s.executionNodes.NodeIDs())
	s.receipts.On("ByBlockID", block.ID()).Return(receipts, nil)

	s.connectionFactory.On("GetExecutionAPIClient", mock.Anything).
		Return(s.execClient, &mocks.MockCloser{}, nil)
}

// setupENSuccessResponse configures the execution client mock to return a successful response
func (s *BackendScriptsSuite) setupENSuccessResponse(blockID flow.Identifier) {
	expectedExecRequest := &execproto.ExecuteScriptAtBlockIDRequest{
		BlockId:   blockID[:],
		Script:    s.script,
		Arguments: s.arguments,
	}

	s.execClient.On("ExecuteScriptAtBlockID", mock.Anything, expectedExecRequest).
		Return(&execproto.ExecuteScriptAtBlockIDResponse{
			Value: expectedResponse,
		}, nil)
}

// setupENFailingResponse configures the execution client mock to return a failing response
func (s *BackendScriptsSuite) setupENFailingResponse(blockID flow.Identifier, err error) {
	expectedExecRequest := &execproto.ExecuteScriptAtBlockIDRequest{
		BlockId:   blockID[:],
		Script:    s.failingScript,
		Arguments: s.arguments,
	}

	s.execClient.On("ExecuteScriptAtBlockID", mock.Anything, expectedExecRequest).
		Return(nil, err)
}

// TestExecuteScriptOnExecutionNode_HappyPath tests that the backend successfully executes scripts
// on execution nodes
func (s *BackendScriptsSuite) TestExecuteScriptOnExecutionNode_HappyPath() {
	ctx := context.Background()

	s.setupExecutionNodes(s.block)
	s.setupENSuccessResponse(s.block.ID())

	scripts := s.defaultBackend(execmock.NewScriptExecutor(s.T()), query_mode.IndexQueryModeExecutionNodesOnly)

	s.Run("GetAccount", func() {
		s.testExecuteScriptAtLatestBlock(ctx, scripts, codes.OK)
	})

	s.Run("ExecuteScriptAtBlockID", func() {
		s.testExecuteScriptAtBlockID(ctx, scripts, codes.OK)
	})

	s.Run("ExecuteScriptAtBlockHeight", func() {
		s.testExecuteScriptAtBlockHeight(ctx, scripts, codes.OK)
	})
}

// TestExecuteScriptOnExecutionNode_Fails tests that the backend returns an error when the execution
// node returns an error
func (s *BackendScriptsSuite) TestExecuteScriptOnExecutionNode_Fails() {
	ctx := context.Background()

	// use a status code that's not used in the API to make sure it's passed through
	statusCode := codes.FailedPrecondition
	errToReturn := status.Error(statusCode, "random error")

	s.setupExecutionNodes(s.block)
	s.setupENFailingResponse(s.block.ID(), errToReturn)

	scripts := s.defaultBackend(execmock.NewScriptExecutor(s.T()), query_mode.IndexQueryModeExecutionNodesOnly)

	s.Run("GetAccount", func() {
		s.testExecuteScriptAtLatestBlock(ctx, scripts, statusCode)
	})

	s.Run("ExecuteScriptAtBlockID", func() {
		s.testExecuteScriptAtBlockID(ctx, scripts, statusCode)
	})

	s.Run("ExecuteScriptAtBlockHeight", func() {
		s.testExecuteScriptAtBlockHeight(ctx, scripts, statusCode)
	})
}

// TestExecuteScriptFromStorage_HappyPath tests that the backend successfully executes scripts using
// the local storage
func (s *BackendScriptsSuite) TestExecuteScriptFromStorage_HappyPath() {
	ctx := context.Background()

	scriptExecutor := execmock.NewScriptExecutor(s.T())
	scriptExecutor.On("ExecuteAtBlockHeight", mock.Anything, s.script, s.arguments, s.block.Header.Height).
		Return(expectedResponse, nil)

	scripts := s.defaultBackend(scriptExecutor, query_mode.IndexQueryModeLocalOnly)

	s.Run("GetAccount - happy path", func() {
		s.testExecuteScriptAtLatestBlock(ctx, scripts, codes.OK)
	})

	s.Run("GetAccountAtLatestBlock - happy path", func() {
		s.testExecuteScriptAtBlockID(ctx, scripts, codes.OK)
	})

	s.Run("GetAccountAtBlockHeight - happy path", func() {
		s.testExecuteScriptAtBlockHeight(ctx, scripts, codes.OK)
	})
}

// TestExecuteScriptFromStorage_Fails tests that errors received from local storage are handled
// and converted to the appropriate status code
func (s *BackendScriptsSuite) TestExecuteScriptFromStorage_Fails() {
	ctx := context.Background()

	scriptExecutor := execmock.NewScriptExecutor(s.T())
	scripts := s.defaultBackend(scriptExecutor, query_mode.IndexQueryModeLocalOnly)

	testCases := []struct {
		err        error
		statusCode codes.Code
	}{
		{
			err:        storage.ErrHeightNotIndexed,
			statusCode: codes.OutOfRange,
		},
		{
			err:        storage.ErrNotFound,
			statusCode: codes.NotFound,
		},
		{
			err:        fmt.Errorf("system error"),
			statusCode: codes.Internal,
		},
		{
			err:        cadenceErr,
			statusCode: codes.InvalidArgument,
		},
		{
			err:        fvmFailureErr,
			statusCode: codes.Internal,
		},
	}

	for _, tt := range testCases {
		scriptExecutor.On("ExecuteAtBlockHeight", mock.Anything, s.failingScript, s.arguments, s.block.Header.Height).
			Return(nil, tt.err).Times(3)

		s.Run(fmt.Sprintf("GetAccount - fails with %v", tt.err), func() {
			s.testExecuteScriptAtLatestBlock(ctx, scripts, tt.statusCode)
		})

		s.Run(fmt.Sprintf("GetAccountAtLatestBlock - fails with %v", tt.err), func() {
			s.testExecuteScriptAtBlockID(ctx, scripts, tt.statusCode)
		})

		s.Run(fmt.Sprintf("GetAccountAtBlockHeight - fails with %v", tt.err), func() {
			s.testExecuteScriptAtBlockHeight(ctx, scripts, tt.statusCode)
		})
	}
}

// TestExecuteScriptWithFailover_HappyPath tests that when an error is returned executing a script
// from local storage, the backend will attempt to run it on an execution node
func (s *BackendScriptsSuite) TestExecuteScriptWithFailover_HappyPath() {
	ctx := context.Background()

	errors := []error{
		storage.ErrHeightNotIndexed,
		storage.ErrNotFound,
		fmt.Errorf("system error"),
		fvmFailureErr,
		compLimitErr,
		memLimitErr,
	}

	s.setupExecutionNodes(s.block)
	s.setupENSuccessResponse(s.block.ID())

	scriptExecutor := execmock.NewScriptExecutor(s.T())
	scripts := s.defaultBackend(scriptExecutor, query_mode.IndexQueryModeFailover)

	for _, errToReturn := range errors {
		// configure local script executor to fail
		scriptExecutor.On("ExecuteAtBlockHeight", mock.Anything, s.script, s.arguments, s.block.Header.Height).
			Return(nil, errToReturn).Times(3)

		s.Run(fmt.Sprintf("ExecuteScriptAtLatestBlock - recovers %v", errToReturn), func() {
			s.testExecuteScriptAtLatestBlock(ctx, scripts, codes.OK)
		})

		s.Run(fmt.Sprintf("ExecuteScriptAtBlockID - recovers %v", errToReturn), func() {
			s.testExecuteScriptAtBlockID(ctx, scripts, codes.OK)
		})

		s.Run(fmt.Sprintf("ExecuteScriptAtBlockHeight - recovers %v", errToReturn), func() {
			s.testExecuteScriptAtBlockHeight(ctx, scripts, codes.OK)
		})
	}
}

// TestExecuteScriptWithFailover_SkippedForCorrectCodes tests that failover is skipped for
// FVM errors that result in InvalidArgument or Canceled errors
func (s *BackendScriptsSuite) TestExecuteScriptWithFailover_SkippedForCorrectCodes() {
	ctx := context.Background()

	// configure local script executor to fail
	scriptExecutor := execmock.NewScriptExecutor(s.T())
	scripts := s.defaultBackend(scriptExecutor, query_mode.IndexQueryModeFailover)

	testCases := []struct {
		err        error
		statusCode codes.Code
	}{
		{
			err:        cadenceErr,
			statusCode: codes.InvalidArgument,
		},
		{
			err:        ctxCancelErr,
			statusCode: codes.Canceled,
		},
	}

	for _, tt := range testCases {
		scriptExecutor.On("ExecuteAtBlockHeight", mock.Anything, s.failingScript, s.arguments, s.block.Header.Height).
			Return(nil, tt.err).
			Times(3)

		s.Run(fmt.Sprintf("ExecuteScriptAtLatestBlock - %s", tt.statusCode), func() {
			s.testExecuteScriptAtLatestBlock(ctx, scripts, tt.statusCode)
		})

		s.Run(fmt.Sprintf("ExecuteScriptAtBlockID - %s", tt.statusCode), func() {
			s.testExecuteScriptAtBlockID(ctx, scripts, tt.statusCode)
		})

		s.Run(fmt.Sprintf("ExecuteScriptAtBlockHeight - %s", tt.statusCode), func() {
			s.testExecuteScriptAtBlockHeight(ctx, scripts, tt.statusCode)
		})
	}
}

// TestExecuteScriptWithFailover_ReturnsENErrors tests that when an error is returned from the execution
// node during a failover, it is returned to the caller.
func (s *BackendScriptsSuite) TestExecuteScriptWithFailover_ReturnsENErrors() {
	ctx := context.Background()

	// use a status code that's not used in the API to make sure it's passed through
	statusCode := codes.FailedPrecondition
	errToReturn := status.Error(statusCode, "random error")

	// setup the execution client mocks
	s.setupExecutionNodes(s.block)
	s.setupENFailingResponse(s.block.ID(), errToReturn)

	// configure local script executor to fail
	scriptExecutor := execmock.NewScriptExecutor(s.T())
	scriptExecutor.On("ExecuteAtBlockHeight", mock.Anything, mock.Anything, mock.Anything, s.block.Header.Height).
		Return(nil, storage.ErrHeightNotIndexed)

	scripts := s.defaultBackend(scriptExecutor, query_mode.IndexQueryModeFailover)

	s.Run("ExecuteScriptAtLatestBlock", func() {
		s.testExecuteScriptAtLatestBlock(ctx, scripts, statusCode)
	})

	s.Run("ExecuteScriptAtBlockID", func() {
		s.testExecuteScriptAtBlockID(ctx, scripts, statusCode)
	})

	s.Run("ExecuteScriptAtBlockHeight", func() {
		s.testExecuteScriptAtBlockHeight(ctx, scripts, statusCode)
	})
}

// TestExecuteScriptAtLatestBlockFromStorage_InconsistentState tests that signaler context received error when node state is
// inconsistent
func (s *BackendScriptsSuite) TestExecuteScriptAtLatestBlockFromStorage_InconsistentState() {
	scriptExecutor := execmock.NewScriptExecutor(s.T())
	scripts := s.defaultBackend(scriptExecutor, query_mode.IndexQueryModeLocalOnly)

	s.Run(fmt.Sprintf("ExecuteScriptAtLatestBlock - fails with %v", "inconsistent node's state"), func() {
		s.state.On("Sealed").Return(s.snapshot, nil)

		err := fmt.Errorf("inconsistent node's state")
		s.snapshot.On("Head").Return(nil, err)

		signCtxErr := irrecoverable.NewExceptionf("failed to lookup sealed header: %w", err)
		signalerCtx := irrecoverable.WithSignalerContext(context.Background(),
			irrecoverable.NewMockSignalerContextExpectError(s.T(), context.Background(), signCtxErr))

		actual, err := scripts.ExecuteScriptAtLatestBlock(signalerCtx, s.script, s.arguments)
		s.Require().Error(err)
		s.Require().Nil(actual)
	})
}

func (s *BackendScriptsSuite) testExecuteScriptAtLatestBlock(ctx context.Context, scripts *Scripts, statusCode codes.Code) {
	s.state.On("Sealed").Return(s.snapshot, nil).Once()
	s.snapshot.On("Head").Return(s.block.Header, nil).Once()

	if statusCode == codes.OK {
		actual, err := scripts.ExecuteScriptAtLatestBlock(ctx, s.script, s.arguments)
		s.Require().NoError(err)
		s.Require().Equal(expectedResponse, actual)
	} else {
		actual, err := scripts.ExecuteScriptAtLatestBlock(ctx, s.failingScript, s.arguments)
		s.Require().Error(err)
		s.Require().Equal(statusCode, status.Code(err), "error code mismatch: expected %d, got %d: %s", statusCode, status.Code(err), err)
		s.Require().Nil(actual)
	}
}

func (s *BackendScriptsSuite) testExecuteScriptAtBlockID(ctx context.Context, scripts *Scripts, statusCode codes.Code) {
	blockID := s.block.ID()
	s.headers.On("ByBlockID", blockID).Return(s.block.Header, nil).Once()

	if statusCode == codes.OK {
		actual, err := scripts.ExecuteScriptAtBlockID(ctx, blockID, s.script, s.arguments)
		s.Require().NoError(err)
		s.Require().Equal(expectedResponse, actual)
	} else {
		actual, err := scripts.ExecuteScriptAtBlockID(ctx, blockID, s.failingScript, s.arguments)
		s.Require().Error(err)
		s.Require().Equal(statusCode, status.Code(err), "error code mismatch: expected %d, got %d: %s", statusCode, status.Code(err), err)
		s.Require().Nil(actual)
	}
}

func (s *BackendScriptsSuite) testExecuteScriptAtBlockHeight(ctx context.Context, scripts *Scripts, statusCode codes.Code) {
	height := s.block.Header.Height
	s.headers.On("ByHeight", height).Return(s.block.Header, nil).Once()

	if statusCode == codes.OK {
		actual, err := scripts.ExecuteScriptAtBlockHeight(ctx, height, s.script, s.arguments)
		s.Require().NoError(err)
		s.Require().Equal(expectedResponse, actual)
	} else {
		actual, err := scripts.ExecuteScriptAtBlockHeight(ctx, height, s.failingScript, s.arguments)
		s.Require().Error(err)
		s.Require().Equalf(statusCode, status.Code(err), "error code mismatch: expected %d, got %d: %s", statusCode, status.Code(err), err)
		s.Require().Nil(actual)
	}
}
