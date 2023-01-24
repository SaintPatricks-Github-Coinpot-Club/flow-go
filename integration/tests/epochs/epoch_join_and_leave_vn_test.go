package epochs

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/onflow/flow-go/model/flow"
)

func TestEpochJoinAndLeaveVN(t *testing.T) {
	suite.Run(t, new(EpochJoinAndLeaveVNSuite))
}

type EpochJoinAndLeaveVNSuite struct {
	DynamicEpochTransitionSuite
}

func (s *EpochJoinAndLeaveVNSuite) SetupTest() {
	// require approvals for seals to verify that the joining VN is producing valid seals in the second epoch
	s.RequiredSealApprovals = 1
	// increase epoch length to account for greater sealing lag due to above
	s.StakingAuctionLen = 100
	s.DKGPhaseLen = 100
	s.EpochLen = 450
	s.EpochCommitSafetyThreshold = 20
	s.DynamicEpochTransitionSuite.SetupTest()
}

// TestEpochJoinAndLeaveVN should update verification nodes and assert healthy network conditions
// after the epoch transition completes. See health check function for details.
func (s *EpochJoinAndLeaveVNSuite) TestEpochJoinAndLeaveVN() {
	s.runTestEpochJoinAndLeave(flow.RoleVerification, s.assertNetworkHealthyAfterVNChange)
}
