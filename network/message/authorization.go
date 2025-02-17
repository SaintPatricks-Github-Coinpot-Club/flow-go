package message

import (
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/model/libp2p/message"
	"github.com/onflow/flow-go/model/messages"
	"github.com/onflow/flow-go/network/channels"
)

var authorizationConfigs map[string]MsgAuthConfig

// MsgAuthConfig contains authorization information for a specific flow message. The authorization
// is represented as a map from network channel -> list of all roles allowed to send the message on
// the channel.
type MsgAuthConfig struct {
	// Name is the string representation of the message type.
	Name string
	// Type is a func that returns a new instance of message type.
	Type func() interface{}
	// Config is the mapping of network channel to list of authorized flow roles.
	Config map[channels.Channel]flow.RoleList
}

// EnsureAuthorized checks if the specified role is authorized to send the message on the provided channel and
// asserts that the message is authorized to be sent on the channel.
// Expected error returns during normal operations:
//   - ErrUnauthorizedMessageOnChannel: the channel is not included in the message's list of authorized channels
//   - ErrUnauthorizedRole: the role is not included in the message's list of authorized roles for the provided channel
func (m MsgAuthConfig) EnsureAuthorized(role flow.Role, channel channels.Channel) error {
	authorizedRoles, ok := m.Config[channel]
	if !ok {
		return ErrUnauthorizedMessageOnChannel
	}

	if !authorizedRoles.Contains(role) {
		return ErrUnauthorizedRole
	}

	return nil
}

func initializeMessageAuthConfigsMap() {
	authorizationConfigs = make(map[string]MsgAuthConfig)

	// consensus
	authorizationConfigs[BlockProposal] = MsgAuthConfig{
		Name: BlockProposal,
		Type: func() interface{} {
			return new(messages.BlockProposal)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.ConsensusCommittee: {flow.RoleConsensus},
			channels.PushBlocks:         {flow.RoleConsensus}, // channel alias ReceiveBlocks = PushBlocks
		},
	}
	authorizationConfigs[BlockVote] = MsgAuthConfig{
		Name: BlockVote,
		Type: func() interface{} {
			return new(messages.BlockVote)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.ConsensusCommittee: {flow.RoleConsensus},
		},
	}

	// protocol state sync
	authorizationConfigs[SyncRequest] = MsgAuthConfig{
		Name: SyncRequest,
		Type: func() interface{} {
			return new(messages.SyncRequest)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.SyncCommittee:     flow.Roles(),
			channels.SyncClusterPrefix: {flow.RoleCollection},
		},
	}
	authorizationConfigs[SyncResponse] = MsgAuthConfig{
		Name: SyncResponse,
		Type: func() interface{} {
			return new(messages.SyncResponse)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.SyncCommittee:     {flow.RoleConsensus},
			channels.SyncClusterPrefix: {flow.RoleCollection},
		},
	}
	authorizationConfigs[RangeRequest] = MsgAuthConfig{
		Name: RangeRequest,
		Type: func() interface{} {
			return new(messages.RangeRequest)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.SyncCommittee:     flow.Roles(),
			channels.SyncClusterPrefix: {flow.RoleCollection},
		},
	}
	authorizationConfigs[BatchRequest] = MsgAuthConfig{
		Name: BatchRequest,
		Type: func() interface{} {
			return new(messages.BatchRequest)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.SyncCommittee:     flow.Roles(),
			channels.SyncClusterPrefix: {flow.RoleCollection},
		},
	}
	authorizationConfigs[BlockResponse] = MsgAuthConfig{
		Name: BlockResponse,
		Type: func() interface{} {
			return new(messages.BlockResponse)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.SyncCommittee: {flow.RoleConsensus},
		},
	}

	// cluster consensus
	authorizationConfigs[ClusterBlockProposal] = MsgAuthConfig{
		Name: ClusterBlockProposal,
		Type: func() interface{} {
			return new(messages.ClusterBlockProposal)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.ConsensusClusterPrefix: {flow.RoleCollection},
		},
	}
	authorizationConfigs[ClusterBlockVote] = MsgAuthConfig{
		Name: ClusterBlockVote,
		Type: func() interface{} {
			return new(messages.ClusterBlockVote)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.ConsensusClusterPrefix: {flow.RoleCollection},
		},
	}
	authorizationConfigs[ClusterBlockResponse] = MsgAuthConfig{
		Name: ClusterBlockResponse,
		Type: func() interface{} {
			return new(messages.ClusterBlockResponse)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.SyncClusterPrefix: {flow.RoleCollection},
		},
	}

	// collections, guarantees & transactions
	authorizationConfigs[CollectionGuarantee] = MsgAuthConfig{
		Name: CollectionGuarantee,
		Type: func() interface{} {
			return new(flow.CollectionGuarantee)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.PushGuarantees: {flow.RoleCollection}, // channel alias ReceiveGuarantees = PushGuarantees
		},
	}
	authorizationConfigs[TransactionBody] = MsgAuthConfig{
		Name: TransactionBody,
		Type: func() interface{} {
			return new(flow.TransactionBody)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.PushTransactions: {flow.RoleCollection}, // channel alias ReceiveTransactions = PushTransactions
		},
	}

	// core messages for execution & verification
	authorizationConfigs[ExecutionReceipt] = MsgAuthConfig{
		Name: ExecutionReceipt,
		Type: func() interface{} {
			return new(flow.ExecutionReceipt)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.PushReceipts: {flow.RoleExecution}, // channel alias ReceiveReceipts = PushReceipts
		},
	}
	authorizationConfigs[ResultApproval] = MsgAuthConfig{
		Name: ResultApproval,
		Type: func() interface{} {
			return new(flow.ResultApproval)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.PushApprovals: {flow.RoleVerification}, // channel alias ReceiveApprovals = PushApprovals
		},
	}

	// data exchange for execution of blocks
	authorizationConfigs[ChunkDataRequest] = MsgAuthConfig{
		Name: ChunkDataRequest,
		Type: func() interface{} {
			return new(messages.ChunkDataRequest)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.RequestChunks: {flow.RoleVerification}, // channel alias RequestChunks = ProvideChunks
		},
	}
	authorizationConfigs[ChunkDataResponse] = MsgAuthConfig{
		Name: ChunkDataResponse,
		Type: func() interface{} {
			return new(messages.ChunkDataResponse)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.ProvideChunks: {flow.RoleExecution}, // channel alias RequestChunks = ProvideChunks
		},
	}

	// result approvals
	authorizationConfigs[ApprovalRequest] = MsgAuthConfig{
		Name: ApprovalRequest,
		Type: func() interface{} {
			return new(messages.ApprovalRequest)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.RequestApprovalsByChunk: {flow.RoleConsensus}, // channel alias ProvideApprovalsByChunk  = RequestApprovalsByChunk
		},
	}
	authorizationConfigs[ApprovalResponse] = MsgAuthConfig{
		Name: ApprovalResponse,
		Type: func() interface{} {
			return new(messages.ApprovalResponse)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.ProvideApprovalsByChunk: {flow.RoleVerification}, // channel alias ProvideApprovalsByChunk  = RequestApprovalsByChunk

		},
	}

	// generic entity exchange engines
	authorizationConfigs[EntityRequest] = MsgAuthConfig{
		Name: EntityRequest,
		Type: func() interface{} {
			return new(messages.EntityRequest)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.RequestReceiptsByBlockID: {flow.RoleConsensus},
			channels.RequestCollections:       {flow.RoleAccess, flow.RoleExecution},
		},
	}
	authorizationConfigs[EntityResponse] = MsgAuthConfig{
		Name: EntityResponse,
		Type: func() interface{} {
			return new(messages.EntityResponse)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.ProvideReceiptsByBlockID: {flow.RoleExecution},
			channels.ProvideCollections:       {flow.RoleCollection},
		},
	}

	// testing
	authorizationConfigs[TestMessage] = MsgAuthConfig{
		Name: TestMessage,
		Type: func() interface{} {
			return new(message.TestMessage)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.TestNetworkChannel: flow.Roles(),
			channels.TestMetricsChannel: flow.Roles(),
		},
	}

	// DKG
	authorizationConfigs[DKGMessage] = MsgAuthConfig{
		Name: DKGMessage,
		Type: func() interface{} {
			return new(messages.DKGMessage)
		},
		Config: map[channels.Channel]flow.RoleList{
			channels.DKGCommittee: {flow.RoleConsensus},
		},
	}
}

// GetMessageAuthConfig checks the underlying type and returns the correct
// message auth Config.
// Expected error returns during normal operations:
//   - ErrUnknownMsgType : if underlying type of v does  not match any of the known message types
func GetMessageAuthConfig(v interface{}) (MsgAuthConfig, error) {
	switch v.(type) {
	// consensus
	case *messages.BlockProposal:
		return authorizationConfigs[BlockProposal], nil
	case *messages.BlockVote:
		return authorizationConfigs[BlockVote], nil

	// protocol state sync
	case *messages.SyncRequest:
		return authorizationConfigs[SyncRequest], nil
	case *messages.SyncResponse:
		return authorizationConfigs[SyncResponse], nil
	case *messages.RangeRequest:
		return authorizationConfigs[RangeRequest], nil
	case *messages.BatchRequest:
		return authorizationConfigs[BatchRequest], nil
	case *messages.BlockResponse:
		return authorizationConfigs[BlockResponse], nil

	// cluster consensus
	case *messages.ClusterBlockProposal:
		return authorizationConfigs[ClusterBlockProposal], nil
	case *messages.ClusterBlockVote:
		return authorizationConfigs[ClusterBlockVote], nil
	case *messages.ClusterBlockResponse:
		return authorizationConfigs[ClusterBlockResponse], nil

	// collections, guarantees & transactions
	case *flow.CollectionGuarantee:
		return authorizationConfigs[CollectionGuarantee], nil
	case *flow.TransactionBody:
		return authorizationConfigs[TransactionBody], nil

	// core messages for execution & verification
	case *flow.ExecutionReceipt:
		return authorizationConfigs[ExecutionReceipt], nil
	case *flow.ResultApproval:
		return authorizationConfigs[ResultApproval], nil

	// data exchange for execution of blocks
	case *messages.ChunkDataRequest:
		return authorizationConfigs[ChunkDataRequest], nil
	case *messages.ChunkDataResponse:
		return authorizationConfigs[ChunkDataResponse], nil

	// result approvals
	case *messages.ApprovalRequest:
		return authorizationConfigs[ApprovalRequest], nil
	case *messages.ApprovalResponse:
		return authorizationConfigs[ApprovalResponse], nil

	// generic entity exchange engines
	case *messages.EntityRequest:
		return authorizationConfigs[EntityRequest], nil
	case *messages.EntityResponse:
		return authorizationConfigs[EntityResponse], nil

	// testing
	case *message.TestMessage:
		return authorizationConfigs[TestMessage], nil

	// dkg
	case *messages.DKGMessage:
		return authorizationConfigs[DKGMessage], nil

	default:
		return MsgAuthConfig{}, NewUnknownMsgTypeErr(v)
	}
}

// GetAllMessageAuthConfigs returns all the configured message auth configurations.
func GetAllMessageAuthConfigs() []MsgAuthConfig {
	configs := make([]MsgAuthConfig, len(authorizationConfigs))

	i := 0
	for _, config := range authorizationConfigs {
		configs[i] = config
		i++
	}

	return configs
}
