package cohort1

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"

	accessproto "github.com/onflow/flow/protobuf/go/flow/access"
	"github.com/onflow/flow/protobuf/go/flow/entities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/onflow/flow-go/engine/access/rpc/backend/query_mode"
	"github.com/onflow/flow-go/integration/tests/mvp"
	"github.com/onflow/flow-go/utils/dsl"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/onflow/cadence"

	sdk "github.com/onflow/flow-go-sdk"
	client "github.com/onflow/flow-go-sdk/access/grpc"

	"github.com/onflow/flow-go/integration/testnet"
	"github.com/onflow/flow-go/integration/tests/lib"
	"github.com/onflow/flow-go/integration/utils"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/utils/unittest"
)

// This is a collection of tests that validate various Access API endpoints work as expected.

var (
	simpleScript       = `access(all) fun main(): Int { return 42; }`
	simpleScriptResult = cadence.NewInt(42)

	OriginalContract = dsl.Contract{
		Name: "TestingContract",
		Members: []dsl.CadenceCode{
			dsl.Code(`
				access(all) fun message(): String {
					return "Initial Contract"
				}`,
			),
		},
	}

	UpdatedContract = dsl.Contract{
		Name: "TestingContract",
		Members: []dsl.CadenceCode{
			dsl.Code(`
				access(all) fun message(): String {
					return "Updated Contract"
				}`,
			),
		},
	}
)

const (
	GetMessageScript = `
import TestingContract from 0x%s

access(all)
fun main(): String {
	return TestingContract.message()
}`
)

func TestAccessAPI(t *testing.T) {
	suite.Run(t, new(AccessAPISuite))
}

type AccessAPISuite struct {
	suite.Suite

	log zerolog.Logger

	// root context for the current test
	ctx    context.Context
	cancel context.CancelFunc

	net *testnet.FlowNetwork

	accessNode2   *testnet.Container
	an1Client     *client.Client
	an2Client     *client.Client
	serviceClient *testnet.Client
}

func (s *AccessAPISuite) TearDownTest() {
	s.log.Info().Msg("================> Start TearDownTest")
	s.net.Remove()
	s.cancel()
	s.log.Info().Msg("================> Finish TearDownTest")
}

func (s *AccessAPISuite) SetupTest() {
	s.log = unittest.LoggerForTest(s.Suite.T(), zerolog.InfoLevel)
	s.log.Info().Msg("================> SetupTest")
	defer func() {
		s.log.Info().Msg("================> Finish SetupTest")
	}()

	// access node
	defaultAccessConfig := testnet.NewNodeConfig(
		flow.RoleAccess,
		testnet.WithLogLevel(zerolog.FatalLevel),
		// make sure test continues to test as expected if the default config changes
		testnet.WithAdditionalFlagf("--script-execution-mode=%s", query_mode.IndexQueryModeExecutionNodesOnly),
		testnet.WithAdditionalFlagf("--tx-result-query-mode=%s", query_mode.IndexQueryModeExecutionNodesOnly),
	)

	indexingAccessConfig := testnet.NewNodeConfig(
		flow.RoleAccess,
		testnet.WithLogLevel(zerolog.InfoLevel),
		testnet.WithAdditionalFlag("--execution-data-sync-enabled=true"),
		testnet.WithAdditionalFlagf("--execution-data-dir=%s", testnet.DefaultExecutionDataServiceDir),
		testnet.WithAdditionalFlag("--execution-data-retry-delay=1s"),
		testnet.WithAdditionalFlag("--execution-data-indexing-enabled=true"),
		testnet.WithAdditionalFlagf("--execution-state-dir=%s", testnet.DefaultExecutionStateDir),
		testnet.WithAdditionalFlagf("--script-execution-mode=%s", query_mode.IndexQueryModeLocalOnly),
	)

	consensusConfigs := []func(config *testnet.NodeConfig){
		// `cruise-ctl-fallback-proposal-duration` is set to 250ms instead to of 100ms
		// to purposely slow down the block rate. This is needed since the crypto module
		// update providing faster BLS operations.
		// TODO: fix the access integration test logic to function without slowing down
		// the block rate
		testnet.WithAdditionalFlag("--cruise-ctl-fallback-proposal-duration=250ms"),
		testnet.WithAdditionalFlagf("--required-verification-seal-approvals=%d", 1),
		testnet.WithAdditionalFlagf("--required-construction-seal-approvals=%d", 1),
		testnet.WithLogLevel(zerolog.FatalLevel),
	}

	nodeConfigs := []testnet.NodeConfig{
		testnet.NewNodeConfig(flow.RoleCollection, testnet.WithLogLevel(zerolog.FatalLevel)),
		testnet.NewNodeConfig(flow.RoleCollection, testnet.WithLogLevel(zerolog.FatalLevel)),
		testnet.NewNodeConfig(flow.RoleExecution, testnet.WithLogLevel(zerolog.FatalLevel)),
		testnet.NewNodeConfig(flow.RoleExecution, testnet.WithLogLevel(zerolog.FatalLevel)),
		testnet.NewNodeConfig(flow.RoleConsensus, consensusConfigs...),
		testnet.NewNodeConfig(flow.RoleConsensus, consensusConfigs...),
		testnet.NewNodeConfig(flow.RoleConsensus, consensusConfigs...),
		testnet.NewNodeConfig(flow.RoleVerification, testnet.WithLogLevel(zerolog.FatalLevel)),

		// AN1 should be the a vanilla node to allow other nodes to bootstrap successfully
		defaultAccessConfig,

		// Tests will focus on AN2
		indexingAccessConfig,
	}

	conf := testnet.NewNetworkConfig("access_api_test", nodeConfigs)
	s.net = testnet.PrepareFlowNetwork(s.T(), conf, flow.Localnet)

	// start the network
	s.T().Logf("starting flow network with docker containers")
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.net.Start(s.ctx)

	var err error
	s.accessNode2 = s.net.ContainerByName("access_2")

	s.an2Client, err = s.accessNode2.SDKClient()
	s.Require().NoError(err)

	s.an1Client, err = s.net.ContainerByName(testnet.PrimaryAN).SDKClient()
	s.Require().NoError(err)

	// pause until the network is progressing
	var header *sdk.BlockHeader
	s.Require().Eventually(func() bool {
		header, err = s.an2Client.GetLatestBlockHeader(s.ctx, true)
		s.Require().NoError(err)

		return header.Height > 0
	}, 30*time.Second, 1*time.Second)

	// the service client uses GetAccount and requires the first block to be indexed
	s.Require().Eventually(func() bool {
		s.serviceClient, err = s.accessNode2.TestnetClient()
		return err == nil
	}, 30*time.Second, 1*time.Second)
}

// TestScriptExecutionAndGetAccountsAN1 test the Access API endpoints for executing scripts and getting
// accounts using execution nodes.
//
// Note: not combining AN1, AN2 tests together because that causes a drastic increase in test run times. test cases are read-only
// and should not interfere with each other.
func (s *AccessAPISuite) TestScriptExecutionAndGetAccountsAN1() {
	// deploy the test contract
	_ = s.deployContract(lib.CounterContract, false)
	txResult := s.deployCounter()
	targetHeight := txResult.BlockHeight + 1
	s.waitUntilIndexed(targetHeight)

	// Run tests against Access 1, which uses the execution node
	s.testGetAccount(s.an1Client)
	s.testExecuteScriptWithSimpleScript(s.an1Client)
	s.testExecuteScriptWithSimpleContract(s.an1Client, targetHeight)
}

// TestScriptExecutionAndGetAccountsAN2 test the Access API endpoints for executing scripts and getting
// accounts using local storage.
//
// Note: not combining AN1, AN2 tests together because that causes a drastic increase in test run times. test cases are read-only
// and should not interfere with each other.
func (s *AccessAPISuite) TestScriptExecutionAndGetAccountsAN2() {
	// deploy the test contract
	_ = s.deployContract(lib.CounterContract, false)
	txResult := s.deployCounter()
	targetHeight := txResult.BlockHeight + 1
	s.waitUntilIndexed(targetHeight)

	// Run tests against Access 2, which uses local storage
	s.testGetAccount(s.an2Client)
	s.testExecuteScriptWithSimpleScript(s.an2Client)
	s.testExecuteScriptWithSimpleContract(s.an2Client, targetHeight)
}

func (s *AccessAPISuite) TestMVPScriptExecutionLocalStorage() {
	// this is a specialized test that creates accounts, deposits funds, deploys contracts, etc, and
	// uses the provided access node to handle the Access API calls. there is an existing test that
	// covers the default config, so we only need to test with local storage.
	mvp.RunMVPTest(s.T(), s.ctx, s.net, s.accessNode2)
}

// TestSendAndSubscribeTransactionStatuses tests the functionality of sending and subscribing to transaction statuses.
//
// This test verifies that a transaction can be created, signed, sent to the access API, and then the status of the transaction
// can be subscribed to. It performs the following steps:
// 1. Establishes a connection to the access API.
// 2. Creates a new account key and prepares a transaction for account creation.
// 3. Signs the transaction.
// 4. Sends and subscribes to the transaction status using the access API.
// 5. Verifies the received transaction statuses, ensuring they are received in order and the final status is "SEALED".
func (s *AccessAPISuite) TestSendAndSubscribeTransactionStatuses() {
	accessNodeContainer := s.net.ContainerByName(testnet.PrimaryAN)

	// Establish a gRPC connection to the access API
	conn, err := grpc.Dial(accessNodeContainer.Addr(testnet.GRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Require().NotNil(conn)

	// Create a client for the access API
	accessClient := accessproto.NewAccessAPIClient(conn)
	serviceClient, err := accessNodeContainer.TestnetClient()
	s.Require().NoError(err)
	s.Require().NotNil(serviceClient)

	// Get the latest block ID
	latestBlockID, err := serviceClient.GetLatestBlockID(s.ctx)
	s.Require().NoError(err)

	// Generate a new account transaction
	accountKey := test.AccountKeyGenerator().New()
	payer := serviceClient.SDKServiceAddress()

	tx, err := templates.CreateAccount([]*sdk.AccountKey{accountKey}, nil, payer)
	s.Require().NoError(err)
	tx.SetComputeLimit(1000).
		SetReferenceBlockID(sdk.HexToID(latestBlockID.String())).
		SetProposalKey(payer, 0, serviceClient.GetAndIncrementSeqNumber()).
		SetPayer(payer)

	tx, err = serviceClient.SignTransaction(tx)
	s.Require().NoError(err)

	// Convert the transaction to a message format expected by the access API
	authorizers := make([][]byte, len(tx.Authorizers))
	for i, auth := range tx.Authorizers {
		authorizers[i] = auth.Bytes()
	}

	convertToMessageSig := func(sigs []sdk.TransactionSignature) []*entities.Transaction_Signature {
		msgSigs := make([]*entities.Transaction_Signature, len(sigs))
		for i, sig := range sigs {
			msgSigs[i] = &entities.Transaction_Signature{
				Address:   sig.Address.Bytes(),
				KeyId:     uint32(sig.KeyIndex),
				Signature: sig.Signature,
			}
		}

		return msgSigs
	}

	transactionMsg := &entities.Transaction{
		Script:           tx.Script,
		Arguments:        tx.Arguments,
		ReferenceBlockId: tx.ReferenceBlockID.Bytes(),
		GasLimit:         tx.GasLimit,
		ProposalKey: &entities.Transaction_ProposalKey{
			Address:        tx.ProposalKey.Address.Bytes(),
			KeyId:          uint32(tx.ProposalKey.KeyIndex),
			SequenceNumber: tx.ProposalKey.SequenceNumber,
		},
		Payer:              tx.Payer.Bytes(),
		Authorizers:        authorizers,
		PayloadSignatures:  convertToMessageSig(tx.PayloadSignatures),
		EnvelopeSignatures: convertToMessageSig(tx.EnvelopeSignatures),
	}

	// Send and subscribe to the transaction status using the access API
	subClient, err := accessClient.SendAndSubscribeTransactionStatuses(s.ctx, &accessproto.SendAndSubscribeTransactionStatusesRequest{
		Transaction:          transactionMsg,
		EventEncodingVersion: entities.EventEncodingVersion_CCF_V0,
	})
	s.Require().NoError(err)

	expectedCounter := uint64(0)
	lastReportedTxStatus := entities.TransactionStatus_UNKNOWN
	var txID sdk.Identifier

	for {
		resp, err := subClient.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}

			s.Require().NoError(err)
		}

		if txID == sdk.EmptyID {
			txID = sdk.Identifier(resp.TransactionResults.TransactionId)
		}

		s.Assert().Equal(expectedCounter, resp.GetMessageIndex())
		s.Assert().Equal(txID, sdk.Identifier(resp.TransactionResults.TransactionId))
		// Check if all statuses received one by one. The subscription should send responses for each of the statuses,
		// and the message should be sent in the order of transaction statuses.
		// Expected order: pending(1) -> finalized(2) -> executed(3) -> sealed(4)
		s.Assert().Equal(lastReportedTxStatus, resp.TransactionResults.Status-1)

		expectedCounter++
		lastReportedTxStatus = resp.TransactionResults.Status
	}

	// Check, if the final transaction status is sealed.
	s.Assert().Equal(entities.TransactionStatus_SEALED, lastReportedTxStatus)
}

// TestContractUpdate tests that the Access API can index contract updates, and that the program cache
// is invalidated when a contract is updated.
func (s *AccessAPISuite) TestContractUpdate() {
	txResult := s.deployContract(OriginalContract, false)
	targetHeight := txResult.BlockHeight + 1
	s.waitUntilIndexed(targetHeight)

	script := fmt.Sprintf(GetMessageScript, s.serviceClient.SDKServiceAddress().Hex())

	// execute script and verify we get the original message
	result, err := s.an2Client.ExecuteScriptAtBlockHeight(s.ctx, targetHeight, []byte(script), nil)
	s.Require().NoError(err)
	s.Require().Equal("Initial Contract", string(result.(cadence.String)))

	txResult = s.deployContract(UpdatedContract, true)
	targetHeight = txResult.BlockHeight + 1
	s.waitUntilIndexed(targetHeight)

	// execute script and verify we get the updated message
	result, err = s.an2Client.ExecuteScriptAtBlockHeight(s.ctx, targetHeight, []byte(script), nil)
	s.Require().NoError(err)
	s.Require().Equal("Updated Contract", string(result.(cadence.String)))
}

func (s *AccessAPISuite) testGetAccount(client *client.Client) {
	header, err := client.GetLatestBlockHeader(s.ctx, true)
	s.Require().NoError(err)

	serviceAddress := s.serviceClient.SDKServiceAddress()

	s.Run("get account at latest block", func() {
		account, err := s.waitAccountsUntilIndexed(func() (*sdk.Account, error) {
			return client.GetAccount(s.ctx, serviceAddress)
		})
		s.Require().NoError(err)
		s.Assert().Equal(serviceAddress, account.Address)
		s.Assert().NotZero(account.Balance)
	})

	s.Run("get account block ID", func() {
		account, err := s.waitAccountsUntilIndexed(func() (*sdk.Account, error) {
			return client.GetAccountAtLatestBlock(s.ctx, serviceAddress)
		})
		s.Require().NoError(err)
		s.Assert().Equal(serviceAddress, account.Address)
		s.Assert().NotZero(account.Balance)
	})

	s.Run("get account block height", func() {
		account, err := s.waitAccountsUntilIndexed(func() (*sdk.Account, error) {
			return client.GetAccountAtBlockHeight(s.ctx, serviceAddress, header.Height)
		})
		s.Require().NoError(err)
		s.Assert().Equal(serviceAddress, account.Address)
		s.Assert().NotZero(account.Balance)
	})

	s.Run("get newly created account", func() {
		addr, err := utils.CreateFlowAccount(s.ctx, s.serviceClient)
		s.Require().NoError(err)
		acc, err := client.GetAccount(s.ctx, addr)
		s.Require().NoError(err)
		s.Assert().Equal(addr, acc.Address)
	})
}

func (s *AccessAPISuite) testExecuteScriptWithSimpleScript(client *client.Client) {
	header, err := client.GetLatestBlockHeader(s.ctx, true)
	s.Require().NoError(err)

	s.Run("execute at latest block", func() {
		result, err := s.waitScriptExecutionUntilIndexed(func() (cadence.Value, error) {
			return client.ExecuteScriptAtLatestBlock(s.ctx, []byte(simpleScript), nil)
		})
		s.Require().NoError(err)
		s.Assert().Equal(simpleScriptResult, result)
	})

	s.Run("execute at block height", func() {
		result, err := s.waitScriptExecutionUntilIndexed(func() (cadence.Value, error) {
			return client.ExecuteScriptAtBlockHeight(s.ctx, header.Height, []byte(simpleScript), nil)
		})
		s.Require().NoError(err)
		s.Assert().Equal(simpleScriptResult, result)
	})

	s.Run("execute at block ID", func() {
		result, err := s.waitScriptExecutionUntilIndexed(func() (cadence.Value, error) {
			return client.ExecuteScriptAtBlockID(s.ctx, header.ID, []byte(simpleScript), nil)
		})
		s.Require().NoError(err)
		s.Assert().Equal(simpleScriptResult, result)
	})
}

func (s *AccessAPISuite) testExecuteScriptWithSimpleContract(client *client.Client, targetHeight uint64) {
	header, err := client.GetBlockHeaderByHeight(s.ctx, targetHeight)
	s.Require().NoError(err)

	// Check that the initialized value is set
	serviceAccount := s.serviceClient.Account()
	script := lib.ReadCounterScript(serviceAccount.Address, serviceAccount.Address).ToCadence()

	s.Run("execute at latest block", func() {
		result, err := s.waitScriptExecutionUntilIndexed(func() (cadence.Value, error) {
			return client.ExecuteScriptAtLatestBlock(s.ctx, []byte(script), nil)
		})
		s.Require().NoError(err)
		s.Assert().Equal(lib.CounterInitializedValue, result.(cadence.Int).Int())
	})

	s.Run("execute at block height", func() {
		result, err := s.waitScriptExecutionUntilIndexed(func() (cadence.Value, error) {
			return client.ExecuteScriptAtBlockHeight(s.ctx, header.Height, []byte(script), nil)
		})
		s.Require().NoError(err)
		s.Assert().Equal(lib.CounterInitializedValue, result.(cadence.Int).Int())
	})

	s.Run("execute at block ID", func() {
		result, err := s.waitScriptExecutionUntilIndexed(func() (cadence.Value, error) {
			return client.ExecuteScriptAtBlockID(s.ctx, header.ID, []byte(script), nil)
		})
		s.Require().NoError(err)
		s.Assert().Equal(lib.CounterInitializedValue, result.(cadence.Int).Int())
	})

	s.Run("execute at past block height", func() {
		// targetHeight is when the counter was deployed, use a height before that to check that
		// the contract was deployed, but the value was not yet set
		pastHeight := targetHeight - 2

		result, err := client.ExecuteScriptAtBlockHeight(s.ctx, pastHeight, []byte(script), nil)
		s.Require().NoError(err)

		s.Assert().Equal(lib.CounterDefaultValue, result.(cadence.Int).Int())
	})
}

func (s *AccessAPISuite) deployContract(contract dsl.Contract, isUpdate bool) *sdk.TransactionResult {
	header, err := s.serviceClient.GetLatestSealedBlockHeader(s.ctx)
	s.Require().NoError(err)

	// Deploy the contract
	var tx *sdk.Transaction
	if isUpdate {
		tx, err = s.serviceClient.UpdateContract(s.ctx, header.ID, contract)
	} else {
		tx, err = s.serviceClient.DeployContract(s.ctx, header.ID, contract)
	}
	s.Require().NoError(err)

	result, err := s.serviceClient.WaitForExecuted(s.ctx, tx.ID())
	s.Require().NoError(err)
	s.Require().Empty(result.Error, "deploy tx should be accepted but got: %s", result.Error)

	return result
}

func (s *AccessAPISuite) deployCounter() *sdk.TransactionResult {
	header, err := s.serviceClient.GetLatestSealedBlockHeader(s.ctx)
	s.Require().NoError(err)

	// Add counter to service account
	serviceAddress := s.serviceClient.SDKServiceAddress()
	tx := sdk.NewTransaction().
		SetScript([]byte(lib.CreateCounterTx(serviceAddress).ToCadence())).
		SetReferenceBlockID(sdk.Identifier(header.ID)).
		SetProposalKey(serviceAddress, 0, s.serviceClient.GetAndIncrementSeqNumber()).
		SetPayer(serviceAddress).
		AddAuthorizer(serviceAddress).
		SetComputeLimit(9999)

	err = s.serviceClient.SignAndSendTransaction(s.ctx, tx)
	s.Require().NoError(err)

	result, err := s.serviceClient.WaitForSealed(s.ctx, tx.ID())
	s.Require().NoError(err)
	s.Require().Empty(result.Error, "create counter tx should be accepted but got: %s", result.Error)

	return result
}

type getAccount func() (*sdk.Account, error)
type executeScript func() (cadence.Value, error)

var indexDelay = 10 * time.Second
var indexRetry = 100 * time.Millisecond

// wait for sealed block to get indexed, as there is a delay in syncing blocks between nodes
func (s *AccessAPISuite) waitAccountsUntilIndexed(get getAccount) (*sdk.Account, error) {
	var account *sdk.Account
	var err error
	s.Require().Eventually(func() bool {
		account, err = get()
		return notOutOfRangeError(err)
	}, indexDelay, indexRetry)

	return account, err
}

func (s *AccessAPISuite) waitScriptExecutionUntilIndexed(execute executeScript) (cadence.Value, error) {
	var val cadence.Value
	var err error
	s.Require().Eventually(func() bool {
		val, err = execute()
		return notOutOfRangeError(err)
	}, indexDelay, indexRetry)

	return val, err
}

func (s *AccessAPISuite) waitUntilIndexed(height uint64) {
	// wait until the block is indexed
	// This relying on the fact that the API is configured to only use the local db, and will return
	// an error if the height is not indexed yet.
	//
	// TODO: once the indexed height is include in the Access API's metadata response, we can get
	// ride of this
	s.Require().Eventually(func() bool {
		_, err := s.an2Client.ExecuteScriptAtBlockHeight(s.ctx, height, []byte(simpleScript), nil)
		return err == nil
	}, 30*time.Second, 1*time.Second)
}

// make sure we either don't have an error or the error is not out of range error, since in that case we have to wait a bit longer for index to get synced
func notOutOfRangeError(err error) bool {
	statusErr, ok := status.FromError(err)
	if !ok || err == nil {
		return true
	}
	return statusErr.Code() != codes.OutOfRange
}
