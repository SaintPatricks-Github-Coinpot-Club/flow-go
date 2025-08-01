package verification

import (
	"fmt"

	"github.com/onflow/crypto/hash"

	"github.com/onflow/flow-go/consensus/hotstuff"
	"github.com/onflow/flow-go/consensus/hotstuff/model"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	msig "github.com/onflow/flow-go/module/signature"
)

// StakingSigner creates votes for the collector clusters consensus.
// When a participant votes for a block, it _always_ provide the staking signature
// as part of their vote. StakingSigner is responsible for creating correctly
// signed proposals and votes.
type StakingSigner struct {
	me                  module.Local
	stakingHasher       hash.Hasher
	timeoutObjectHasher hash.Hasher
	signerID            flow.Identifier
}

var _ hotstuff.Signer = (*StakingSigner)(nil)

// NewStakingSigner instantiates a StakingSigner, which signs votes and
// proposals with the staking key.  The generated signatures are aggregatable.
func NewStakingSigner(
	me module.Local,
) *StakingSigner {

	sc := &StakingSigner{
		me:                  me,
		stakingHasher:       msig.NewBLSHasher(msig.CollectorVoteTag),
		timeoutObjectHasher: msig.NewBLSHasher(msig.CollectorTimeoutTag),
		signerID:            me.NodeID(),
	}
	return sc
}

// CreateVote will create a vote with a staking signature for the given block.
func (c *StakingSigner) CreateVote(block *model.Block) (*model.Vote, error) {

	// create the signature data
	sigData, err := c.genSigData(block)
	if err != nil {
		return nil, fmt.Errorf("could not create signature: %w", err)
	}

	vote, err := model.NewVote(model.UntrustedVote{
		View:     block.View,
		BlockID:  block.BlockID,
		SignerID: c.signerID,
		SigData:  sigData,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create vote: %w", err)
	}

	return vote, nil
}

// CreateTimeout will create a signed timeout object for the given view.
func (c *StakingSigner) CreateTimeout(curView uint64, newestQC *flow.QuorumCertificate, lastViewTC *flow.TimeoutCertificate) (*model.TimeoutObject, error) {
	// create timeout object specific message
	msg := MakeTimeoutMessage(curView, newestQC.View)
	sigData, err := c.me.Sign(msg, c.timeoutObjectHasher)
	if err != nil {
		return nil, fmt.Errorf("could not generate signature for timeout object at view %d: %w", curView, err)
	}

	timeout, err := model.NewTimeoutObject(
		model.UntrustedTimeoutObject{
			View:        curView,
			NewestQC:    newestQC,
			LastViewTC:  lastViewTC,
			SignerID:    c.signerID,
			SigData:     sigData,
			TimeoutTick: 0,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not construct timeout object: %w", err)
	}

	return timeout, nil
}

// genSigData generates the signature data for our local node for the given block.
// It returns:
//   - (stakingSig, nil) signature signed with staking key.  The sig is 48 bytes long
//   - (nil, error) if there is any exception
func (c *StakingSigner) genSigData(block *model.Block) ([]byte, error) {
	// create the message to be signed and generate signatures
	msg := MakeVoteMessage(block.View, block.BlockID)

	stakingSig, err := c.me.Sign(msg, c.stakingHasher)
	if err != nil {
		return nil, fmt.Errorf("could not generate staking signature for block (%v) at view %v: %w", block.BlockID, block.View, err)
	}

	return stakingSig, nil
}
