package unittest

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/encoding/ccf"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/crypto"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	"github.com/onflow/flow-go/model/flow"
)

// This file contains service event fixtures for testing purposes.

func EpochSetupRandomSourceFixture() []byte {
	source := make([]byte, flow.EpochSetupRandomSourceLength)
	_, err := rand.Read(source)
	if err != nil {
		panic(err)
	}
	return source
}

// EpochSetupFixtureByChainID returns an EpochSetup service event as a Cadence event
// representation and as a protocol model representation.
func EpochSetupFixtureByChainID(chain flow.ChainID) (flow.Event, *flow.EpochSetup) {
	events := systemcontracts.ServiceEventsForChain(chain)

	event := EventFixture(events.EpochSetup.EventType(), 1, 1, IdentifierFixture(), 0)
	randomSource := EpochSetupRandomSourceFixture()
	event.Payload = EpochSetupFixtureCCF(randomSource)

	expected := &flow.EpochSetup{
		Counter:            1,
		FirstView:          100,
		FinalView:          200,
		DKGPhase1FinalView: 150,
		DKGPhase2FinalView: 160,
		DKGPhase3FinalView: 170,
		RandomSource:       randomSource,
		TargetDuration:     200,
		TargetEndTime:      2000000000,
		Assignments: flow.AssignmentList{
			{
				flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000001"),
				flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000002"),
			},
			{
				flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000003"),
				flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000004"),
			},
		},
		Participants: flow.IdentitySkeletonList{
			{
				Role:          flow.RoleCollection,
				NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000001"),
				Address:       "1.flow.com",
				NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
				StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
				InitialWeight: 100,
			},
			{
				Role:          flow.RoleCollection,
				NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000002"),
				Address:       "2.flow.com",
				NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
				StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
				InitialWeight: 100,
			},
			{
				Role:          flow.RoleCollection,
				NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000003"),
				Address:       "3.flow.com",
				NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
				StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
				InitialWeight: 100,
			},
			{
				Role:          flow.RoleCollection,
				NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000004"),
				Address:       "4.flow.com",
				NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
				StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
				InitialWeight: 100,
			},
			{
				Role:          flow.RoleConsensus,
				NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000011"),
				Address:       "11.flow.com",
				NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "cfdfe8e4362c8f79d11772cb7277ab16e5033a63e8dd5d34caf1b041b77e5b2d63c2072260949ccf8907486e4cfc733c8c42ca0e4e208f30470b0d950856cd47"),
				StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "8207559cd7136af378bba53a8f0196dee3849a3ab02897c1995c3e3f6ca0c4a776c3ae869d1ddbb473090054be2400ad06d7910aa2c5d1780220fdf3765a3c1764bce10c6fe66a5a2be51a422e878518bd750424bb56b8a0ecf0f8ad2057e83f"),
				InitialWeight: 100,
			},
			{
				Role:          flow.RoleExecution,
				NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000021"),
				Address:       "21.flow.com",
				NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "d64318ba0dbf68f3788fc81c41d507c5822bf53154530673127c66f50fe4469ccf1a054a868a9f88506a8999f2386d86fcd2b901779718cba4fb53c2da258f9e"),
				StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "880b162b7ec138b36af401d07868cb08d25746d905395edbb4625bdf105d4bb2b2f4b0f4ae273a296a6efefa7ce9ccb914e39947ce0e83745125cab05d62516076ff0173ed472d3791ccef937597c9ea12381d76f547a092a4981d77ff3fba83"),
				InitialWeight: 100,
			},
			{
				Role:          flow.RoleVerification,
				NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000031"),
				Address:       "31.flow.com",
				NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "697241208dcc9142b6f53064adc8ff1c95760c68beb2ba083c1d005d40181fd7a1b113274e0163c053a3addd47cd528ec6a1f190cf465aac87c415feaae011ae"),
				StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "b1f97d0a06020eca97352e1adde72270ee713c7daf58da7e74bf72235321048b4841bdfc28227964bf18e371e266e32107d238358848bcc5d0977a0db4bda0b4c33d3874ff991e595e0f537c7b87b4ddce92038ebc7b295c9ea20a1492302aa7"),
				InitialWeight: 100,
			},
		},
	}

	return event, expected
}

// EpochCommitFixtureByChainID returns an EpochCommit service event as a Cadence event
// representation and as a protocol model representation.
func EpochCommitFixtureByChainID(chain flow.ChainID) (flow.Event, *flow.EpochCommit) {

	events := systemcontracts.ServiceEventsForChain(chain)

	event := EventFixture(events.EpochCommit.EventType(), 1, 1, IdentifierFixture(), 0)
	event.Payload = EpochCommitFixtureCCF

	expected := &flow.EpochCommit{
		Counter: 1,
		ClusterQCs: []flow.ClusterQCVoteData{
			{
				VoterIDs: []flow.Identifier{
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000001"),
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000002"),
				},
				SigData: MustDecodeSignatureHex("b072ed22ed305acd44818a6c836e09b4e844eebde6a4fdbf5cec983e2872b86c8b0f6c34c0777bf52e385ab7c45dc55d"),
			},
			{
				VoterIDs: []flow.Identifier{
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000003"),
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000004"),
				},
				SigData: MustDecodeSignatureHex("899e266a543e1b3a564f68b22f7be571f2e944ec30fadc4b39e2d5f526ba044c0f3cb2648f8334fc216fa3360a0418b2"),
			},
		},
		DKGGroupKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "8c588266db5f5cda629e83f8aa04ae9413593fac19e4865d06d291c9d14fbdd9bdb86a7a12f9ef8590c79cb635e3163315d193087e9336092987150d0cd2b14ac6365f7dc93eec573752108b8c12368abb65f0652d9f644e5aed611c37926950"),
		DKGParticipantKeys: []crypto.PublicKey{
			MustDecodePublicKeyHex(crypto.BLSBLS12381, "87a339e4e5c74f089da20a33f515d8c8f4464ab53ede5a74aa2432cd1ae66d522da0c122249ee176cd747ddc83ca81090498389384201614caf51eac392c1c0a916dfdcfbbdf7363f9552b6468434add3d3f6dc91a92bbe3ee368b59b7828488"),
		},
		DKGIndexMap: flow.DKGIndexMap{
			flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000011"): 0,
		},
	}

	return event, expected
}

// EpochRecoverFixtureByChainID returns an EpochRecover service event as a Cadence event
// representation and as a protocol model representation.
func EpochRecoverFixtureByChainID(chain flow.ChainID) (flow.Event, *flow.EpochRecover) {

	events := systemcontracts.ServiceEventsForChain(chain)

	randomSource := EpochSetupRandomSourceFixture()
	event := EventFixture(events.EpochRecover.EventType(), 1, 1, IdentifierFixture(), 0)
	event.Payload = EpochRecoverFixtureCCF(randomSource)

	expected := &flow.EpochRecover{
		EpochSetup: flow.EpochSetup{
			Counter:            1,
			FirstView:          100,
			FinalView:          200,
			DKGPhase1FinalView: 150,
			DKGPhase2FinalView: 160,
			DKGPhase3FinalView: 170,
			RandomSource:       randomSource,
			TargetDuration:     200,
			TargetEndTime:      2000000000,
			Assignments: flow.AssignmentList{
				{
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000001"),
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000002"),
				},
				{
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000003"),
					flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000004"),
				},
			},
			Participants: flow.IdentitySkeletonList{
				{
					Role:          flow.RoleCollection,
					NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000001"),
					Address:       "1.flow.com",
					NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
					StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
					InitialWeight: 100,
				},
				{
					Role:          flow.RoleCollection,
					NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000002"),
					Address:       "2.flow.com",
					NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
					StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
					InitialWeight: 100,
				},
				{
					Role:          flow.RoleCollection,
					NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000003"),
					Address:       "3.flow.com",
					NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
					StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
					InitialWeight: 100,
				},
				{
					Role:          flow.RoleCollection,
					NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000004"),
					Address:       "4.flow.com",
					NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),
					StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),
					InitialWeight: 100,
				},
				{
					Role:          flow.RoleConsensus,
					NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000011"),
					Address:       "11.flow.com",
					NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "cfdfe8e4362c8f79d11772cb7277ab16e5033a63e8dd5d34caf1b041b77e5b2d63c2072260949ccf8907486e4cfc733c8c42ca0e4e208f30470b0d950856cd47"),
					StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "8207559cd7136af378bba53a8f0196dee3849a3ab02897c1995c3e3f6ca0c4a776c3ae869d1ddbb473090054be2400ad06d7910aa2c5d1780220fdf3765a3c1764bce10c6fe66a5a2be51a422e878518bd750424bb56b8a0ecf0f8ad2057e83f"),
					InitialWeight: 100,
				},
				{
					Role:          flow.RoleExecution,
					NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000021"),
					Address:       "21.flow.com",
					NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "d64318ba0dbf68f3788fc81c41d507c5822bf53154530673127c66f50fe4469ccf1a054a868a9f88506a8999f2386d86fcd2b901779718cba4fb53c2da258f9e"),
					StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "880b162b7ec138b36af401d07868cb08d25746d905395edbb4625bdf105d4bb2b2f4b0f4ae273a296a6efefa7ce9ccb914e39947ce0e83745125cab05d62516076ff0173ed472d3791ccef937597c9ea12381d76f547a092a4981d77ff3fba83"),
					InitialWeight: 100,
				},
				{
					Role:          flow.RoleVerification,
					NodeID:        flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000031"),
					Address:       "31.flow.com",
					NetworkPubKey: MustDecodePublicKeyHex(crypto.ECDSAP256, "697241208dcc9142b6f53064adc8ff1c95760c68beb2ba083c1d005d40181fd7a1b113274e0163c053a3addd47cd528ec6a1f190cf465aac87c415feaae011ae"),
					StakingPubKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "b1f97d0a06020eca97352e1adde72270ee713c7daf58da7e74bf72235321048b4841bdfc28227964bf18e371e266e32107d238358848bcc5d0977a0db4bda0b4c33d3874ff991e595e0f537c7b87b4ddce92038ebc7b295c9ea20a1492302aa7"),
					InitialWeight: 100,
				},
			},
		},
		EpochCommit: flow.EpochCommit{
			Counter: 1,
			ClusterQCs: []flow.ClusterQCVoteData{
				{
					VoterIDs: []flow.Identifier{
						flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000001"),
						flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000002"),
					},
					SigData: MustDecodeSignatureHex("b072ed22ed305acd44818a6c836e09b4e844eebde6a4fdbf5cec983e2872b86c8b0f6c34c0777bf52e385ab7c45dc55d"),
				},
				{
					VoterIDs: []flow.Identifier{
						flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000003"),
						flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000004"),
					},
					SigData: MustDecodeSignatureHex("899e266a543e1b3a564f68b22f7be571f2e944ec30fadc4b39e2d5f526ba044c0f3cb2648f8334fc216fa3360a0418b2"),
				},
			},
			DKGGroupKey: MustDecodePublicKeyHex(crypto.BLSBLS12381, "8c588266db5f5cda629e83f8aa04ae9413593fac19e4865d06d291c9d14fbdd9bdb86a7a12f9ef8590c79cb635e3163315d193087e9336092987150d0cd2b14ac6365f7dc93eec573752108b8c12368abb65f0652d9f644e5aed611c37926950"),
			DKGParticipantKeys: []crypto.PublicKey{
				MustDecodePublicKeyHex(crypto.BLSBLS12381, "87a339e4e5c74f089da20a33f515d8c8f4464ab53ede5a74aa2432cd1ae66d522da0c122249ee176cd747ddc83ca81090498389384201614caf51eac392c1c0a916dfdcfbbdf7363f9552b6468434add3d3f6dc91a92bbe3ee368b59b7828488"),
			},
			DKGIndexMap: flow.DKGIndexMap{
				flow.MustHexStringToIdentifier("0000000000000000000000000000000000000000000000000000000000000011"): 0,
			},
		},
	}

	return event, expected
}

// VersionBeaconFixtureByChainID returns a VersionTable service event as a Cadence event
// representation and as a protocol model representation.
func VersionBeaconFixtureByChainID(chain flow.ChainID) (flow.Event, *flow.VersionBeacon) {

	events := systemcontracts.ServiceEventsForChain(chain)

	event := EventFixture(events.VersionBeacon.EventType(), 1, 1, IdentifierFixture(), 0)
	event.Payload = VersionBeaconFixtureCCF

	expected := &flow.VersionBeacon{
		VersionBoundaries: []flow.VersionBoundary{
			{
				BlockHeight: 44,
				Version:     "2.13.7-test",
			},
		},
		Sequence: 5,
	}

	return event, expected
}

func ProtocolStateVersionUpgradeFixtureByChainID(chain flow.ChainID) (flow.Event, *flow.ProtocolStateVersionUpgrade) {
	events := systemcontracts.ServiceEventsForChain(chain)

	event := EventFixture(events.ProtocolStateVersionUpgrade.EventType(), 1, 1, IdentifierFixture(), 0)
	event.Payload = ProtocolStateVersionUpgradeFixtureCCF

	expected := &flow.ProtocolStateVersionUpgrade{
		NewProtocolStateVersion: 1,
		ActiveView:              1000,
	}

	return event, expected
}

func createEpochSetupEvent(randomSourceHex string) cadence.Event {

	return cadence.NewEvent([]cadence.Value{
		// counter
		cadence.NewUInt64(1),

		// nodeInfo
		createEpochNodes(),

		// firstView
		cadence.NewUInt64(100),

		// finalView
		cadence.NewUInt64(200),

		// collectorClusters
		createEpochCollectors(),

		// randomSource
		cadence.String(randomSourceHex),

		// DKGPhase1FinalView
		cadence.UInt64(150),

		// DKGPhase2FinalView
		cadence.UInt64(160),

		// DKGPhase3FinalView
		cadence.UInt64(170),

		// targetDuration
		cadence.UInt64(200),

		// targetEndTime
		cadence.UInt64(2000000000),
	}).WithType(newFlowEpochEpochSetupEventType())
}

func createEpochNodes() cadence.Array {

	nodeInfoType := newFlowIDTableStakingNodeInfoStructType()

	nodeInfo1 := cadence.NewStruct([]cadence.Value{
		// id
		cadence.String("0000000000000000000000000000000000000000000000000000000000000001"),

		// role
		cadence.UInt8(1),

		// networkingAddress
		cadence.String("1.flow.com"),

		// networkingKey
		cadence.String("378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),

		// stakingKey
		cadence.String("af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),

		// tokensStaked
		ufix64FromString("0.00000000"),

		// tokensCommitted
		ufix64FromString("1350000.00000000"),

		// tokensUnstaking
		ufix64FromString("0.00000000"),

		// tokensUnstaked
		ufix64FromString("0.00000000"),

		// tokensRewarded
		ufix64FromString("0.00000000"),

		// delegators
		cadence.NewArray([]cadence.Value{}).WithType(cadence.NewVariableSizedArrayType(cadence.UInt32Type)),

		// delegatorIDCounter
		cadence.UInt32(0),

		// tokensRequestedToUnstake
		ufix64FromString("0.00000000"),

		// initialWeight
		cadence.UInt64(100),
	}).WithType(nodeInfoType)

	nodeInfo2 := cadence.NewStruct([]cadence.Value{
		// id
		cadence.String("0000000000000000000000000000000000000000000000000000000000000002"),

		// role
		cadence.UInt8(1),

		// networkingAddress
		cadence.String("2.flow.com"),

		// networkingKey
		cadence.String("378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),

		// stakingKey
		cadence.String("af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),

		// tokensStaked
		ufix64FromString("0.00000000"),

		// tokensCommitted
		ufix64FromString("1350000.00000000"),

		// tokensUnstaking
		ufix64FromString("0.00000000"),

		// tokensUnstaked
		ufix64FromString("0.00000000"),

		// tokensRewarded
		ufix64FromString("0.00000000"),

		// delegators
		cadence.NewArray([]cadence.Value{}).WithType(cadence.NewVariableSizedArrayType(cadence.UInt32Type)),

		// delegatorIDCounter
		cadence.UInt32(0),

		// tokensRequestedToUnstake
		ufix64FromString("0.00000000"),

		// initialWeight
		cadence.UInt64(100),
	}).WithType(nodeInfoType)

	nodeInfo3 := cadence.NewStruct([]cadence.Value{
		// id
		cadence.String("0000000000000000000000000000000000000000000000000000000000000003"),

		// role
		cadence.UInt8(1),

		// networkingAddress
		cadence.String("3.flow.com"),

		// networkingKey
		cadence.String("378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),

		// stakingKey
		cadence.String("af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),

		// tokensStaked
		ufix64FromString("0.00000000"),

		// tokensCommitted
		ufix64FromString("1350000.00000000"),

		// tokensUnstaking
		ufix64FromString("0.00000000"),

		// tokensUnstaked
		ufix64FromString("0.00000000"),

		// tokensRewarded
		ufix64FromString("0.00000000"),

		// delegators
		cadence.NewArray([]cadence.Value{}).WithType(cadence.NewVariableSizedArrayType(cadence.UInt32Type)),

		// delegatorIDCounter
		cadence.UInt32(0),

		// tokensRequestedToUnstake
		ufix64FromString("0.00000000"),

		// initialWeight
		cadence.UInt64(100),
	}).WithType(nodeInfoType)

	nodeInfo4 := cadence.NewStruct([]cadence.Value{
		// id
		cadence.String("0000000000000000000000000000000000000000000000000000000000000004"),

		// role
		cadence.UInt8(1),

		// networkingAddress
		cadence.String("4.flow.com"),

		// networkingKey
		cadence.String("378dbf45d85c614feb10d8bd4f78f4b6ef8eec7d987b937e123255444657fb3da031f232a507e323df3a6f6b8f50339c51d188e80c0e7a92420945cc6ca893fc"),

		// stakingKey
		cadence.String("af4aade26d76bb2ab15dcc89adcef82a51f6f04b3cb5f4555214b40ec89813c7a5f95776ea4fe449de48166d0bbc59b919b7eabebaac9614cf6f9461fac257765415f4d8ef1376a2365ec9960121888ea5383d88a140c24c29962b0a14e4e4e7"),

		// tokensStaked
		ufix64FromString("0.00000000"),

		// tokensCommitted
		ufix64FromString("1350000.00000000"),

		// tokensUnstaking
		ufix64FromString("0.00000000"),

		// tokensUnstaked
		ufix64FromString("0.00000000"),

		// tokensRewarded
		ufix64FromString("0.00000000"),

		// delegators
		cadence.NewArray([]cadence.Value{}).WithType(cadence.NewVariableSizedArrayType(cadence.UInt32Type)),

		// delegatorIDCounter
		cadence.UInt32(0),

		// tokensRequestedToUnstake
		ufix64FromString("0.00000000"),

		// initialWeight
		cadence.UInt64(100),
	}).WithType(nodeInfoType)

	nodeInfo5 := cadence.NewStruct([]cadence.Value{
		// id
		cadence.String("0000000000000000000000000000000000000000000000000000000000000011"),

		// role
		cadence.UInt8(2),

		// networkingAddress
		cadence.String("11.flow.com"),

		// networkingKey
		cadence.String("cfdfe8e4362c8f79d11772cb7277ab16e5033a63e8dd5d34caf1b041b77e5b2d63c2072260949ccf8907486e4cfc733c8c42ca0e4e208f30470b0d950856cd47"),

		// stakingKey
		cadence.String("8207559cd7136af378bba53a8f0196dee3849a3ab02897c1995c3e3f6ca0c4a776c3ae869d1ddbb473090054be2400ad06d7910aa2c5d1780220fdf3765a3c1764bce10c6fe66a5a2be51a422e878518bd750424bb56b8a0ecf0f8ad2057e83f"),

		// tokensStaked
		ufix64FromString("0.00000000"),

		// tokensCommitted
		ufix64FromString("1350000.00000000"),

		// tokensUnstaking
		ufix64FromString("0.00000000"),

		// tokensUnstaked
		ufix64FromString("0.00000000"),

		// tokensRewarded
		ufix64FromString("0.00000000"),

		// delegators
		cadence.NewArray([]cadence.Value{}).WithType(cadence.NewVariableSizedArrayType(cadence.UInt32Type)),

		// delegatorIDCounter
		cadence.UInt32(0),

		// tokensRequestedToUnstake
		ufix64FromString("0.00000000"),

		// initialWeight
		cadence.UInt64(100),
	}).WithType(nodeInfoType)

	nodeInfo6 := cadence.NewStruct([]cadence.Value{
		// id
		cadence.String("0000000000000000000000000000000000000000000000000000000000000021"),

		// role
		cadence.UInt8(3),

		// networkingAddress
		cadence.String("21.flow.com"),

		// networkingKey
		cadence.String("d64318ba0dbf68f3788fc81c41d507c5822bf53154530673127c66f50fe4469ccf1a054a868a9f88506a8999f2386d86fcd2b901779718cba4fb53c2da258f9e"),

		// stakingKey
		cadence.String("880b162b7ec138b36af401d07868cb08d25746d905395edbb4625bdf105d4bb2b2f4b0f4ae273a296a6efefa7ce9ccb914e39947ce0e83745125cab05d62516076ff0173ed472d3791ccef937597c9ea12381d76f547a092a4981d77ff3fba83"),

		// tokensStaked
		ufix64FromString("0.00000000"),

		// tokensCommitted
		ufix64FromString("1350000.00000000"),

		// tokensUnstaking
		ufix64FromString("0.00000000"),

		// tokensUnstaked
		ufix64FromString("0.00000000"),

		// tokensRewarded
		ufix64FromString("0.00000000"),

		// delegators
		cadence.NewArray([]cadence.Value{}).WithType(cadence.NewVariableSizedArrayType(cadence.UInt32Type)),

		// delegatorIDCounter
		cadence.UInt32(0),

		// tokensRequestedToUnstake
		ufix64FromString("0.00000000"),

		// initialWeight
		cadence.UInt64(100),
	}).WithType(nodeInfoType)

	nodeInfo7 := cadence.NewStruct([]cadence.Value{
		// id
		cadence.String("0000000000000000000000000000000000000000000000000000000000000031"),

		// role
		cadence.UInt8(4),

		// networkingAddress
		cadence.String("31.flow.com"),

		// networkingKey
		cadence.String("697241208dcc9142b6f53064adc8ff1c95760c68beb2ba083c1d005d40181fd7a1b113274e0163c053a3addd47cd528ec6a1f190cf465aac87c415feaae011ae"),

		// stakingKey
		cadence.String("b1f97d0a06020eca97352e1adde72270ee713c7daf58da7e74bf72235321048b4841bdfc28227964bf18e371e266e32107d238358848bcc5d0977a0db4bda0b4c33d3874ff991e595e0f537c7b87b4ddce92038ebc7b295c9ea20a1492302aa7"),

		// tokensStaked
		ufix64FromString("0.00000000"),

		// tokensCommitted
		ufix64FromString("1350000.00000000"),

		// tokensUnstaking
		ufix64FromString("0.00000000"),

		// tokensUnstaked
		ufix64FromString("0.00000000"),

		// tokensRewarded
		ufix64FromString("0.00000000"),

		// delegators
		cadence.NewArray([]cadence.Value{}).WithType(cadence.NewVariableSizedArrayType(cadence.UInt32Type)),

		// delegatorIDCounter
		cadence.UInt32(0),

		// tokensRequestedToUnstake
		ufix64FromString("0.00000000"),

		// initialWeight
		cadence.UInt64(100),
	}).WithType(nodeInfoType)

	return cadence.NewArray([]cadence.Value{
		nodeInfo1,
		nodeInfo2,
		nodeInfo3,
		nodeInfo4,
		nodeInfo5,
		nodeInfo6,
		nodeInfo7,
	}).WithType(cadence.NewVariableSizedArrayType(nodeInfoType))
}

func createEpochCollectors() cadence.Array {

	clusterType := NewFlowClusterQCClusterStructType()

	voteType := newFlowClusterQCVoteStructType()

	cluster1 := cadence.NewStruct([]cadence.Value{
		// index
		cadence.NewUInt16(0),

		// nodeWeights
		cadence.NewDictionary([]cadence.KeyValuePair{
			{
				Key:   cadence.String("0000000000000000000000000000000000000000000000000000000000000001"),
				Value: cadence.UInt64(100),
			},
			{
				Key:   cadence.String("0000000000000000000000000000000000000000000000000000000000000002"),
				Value: cadence.UInt64(100),
			},
		}).WithType(cadence.NewMeteredDictionaryType(nil, cadence.StringType, cadence.UInt64Type)),

		// totalWeight
		cadence.NewUInt64(100),

		// generatedVotes
		cadence.NewDictionary(nil).WithType(cadence.NewDictionaryType(cadence.StringType, voteType)),

		// uniqueVoteMessageTotalWeights
		cadence.NewDictionary(nil).WithType(cadence.NewDictionaryType(cadence.StringType, cadence.UInt64Type)),
	}).WithType(clusterType)

	cluster2 := cadence.NewStruct([]cadence.Value{
		// index
		cadence.NewUInt16(1),

		// nodeWeights
		cadence.NewDictionary([]cadence.KeyValuePair{
			{
				Key:   cadence.String("0000000000000000000000000000000000000000000000000000000000000003"),
				Value: cadence.UInt64(100),
			},
			{
				Key:   cadence.String("0000000000000000000000000000000000000000000000000000000000000004"),
				Value: cadence.UInt64(100),
			},
		}).WithType(cadence.NewMeteredDictionaryType(nil, cadence.StringType, cadence.UInt64Type)),

		// totalWeight
		cadence.NewUInt64(0),

		// generatedVotes
		cadence.NewDictionary(nil).WithType(cadence.NewDictionaryType(cadence.StringType, voteType)),

		// uniqueVoteMessageTotalWeights
		cadence.NewDictionary(nil).WithType(cadence.NewDictionaryType(cadence.StringType, cadence.UInt64Type)),
	}).WithType(clusterType)

	return cadence.NewArray([]cadence.Value{
		cluster1,
		cluster2,
	}).WithType(cadence.NewVariableSizedArrayType(clusterType))
}

func createEpochCommitEvent() cadence.Event {

	clusterQCType := newFlowClusterQCClusterQCStructType()

	cluster1 := cadence.NewStruct([]cadence.Value{
		// index
		cadence.UInt16(0),

		// voteSignatures
		cadence.NewArray([]cadence.Value{
			cadence.String("a39cd1e1bf7e2fb0609b7388ce5215a6a4c01eef2aee86e1a007faa28a6b2a3dc876e11bb97cdb26c3846231d2d01e4d"),
			cadence.String("91673ad9c717d396c9a0953617733c128049ac1a639653d4002ab245b121df1939430e313bcbfd06948f6a281f6bf853"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),

		// voteMessage
		cadence.String("irrelevant_for_these_purposes"),

		// voterIDs
		cadence.NewArray([]cadence.Value{
			cadence.String("0000000000000000000000000000000000000000000000000000000000000001"),
			cadence.String("0000000000000000000000000000000000000000000000000000000000000002"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),
	}).WithType(clusterQCType)

	cluster2 := cadence.NewStruct([]cadence.Value{
		// index
		cadence.UInt16(1),

		// voteSignatures
		cadence.NewArray([]cadence.Value{
			cadence.String("b2bff159971852ed63e72c37991e62c94822e52d4fdcd7bf29aaf9fb178b1c5b4ce20dd9594e029f3574cb29533b857a"),
			cadence.String("9931562f0248c9195758da3de4fb92f24fa734cbc20c0cb80280163560e0e0348f843ac89ecbd3732e335940c1e8dccb"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),

		// voteMessage
		cadence.String("irrelevant_for_these_purposes"),

		// voterIDs
		cadence.NewArray([]cadence.Value{
			cadence.String("0000000000000000000000000000000000000000000000000000000000000003"),
			cadence.String("0000000000000000000000000000000000000000000000000000000000000004"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),
	}).WithType(clusterQCType)

	return cadence.NewEvent([]cadence.Value{
		// counter
		cadence.NewUInt64(1),

		// clusterQCs
		cadence.NewArray([]cadence.Value{
			cluster1,
			cluster2,
		}).WithType(cadence.NewVariableSizedArrayType(clusterQCType)),

		// dkgGroupKey
		cadence.String("8c588266db5f5cda629e83f8aa04ae9413593fac19e4865d06d291c9d14fbdd9bdb86a7a12f9ef8590c79cb635e3163315d193087e9336092987150d0cd2b14ac6365f7dc93eec573752108b8c12368abb65f0652d9f644e5aed611c37926950"),

		// dkgPubKeys
		cadence.NewArray([]cadence.Value{
			cadence.String("87a339e4e5c74f089da20a33f515d8c8f4464ab53ede5a74aa2432cd1ae66d522da0c122249ee176cd747ddc83ca81090498389384201614caf51eac392c1c0a916dfdcfbbdf7363f9552b6468434add3d3f6dc91a92bbe3ee368b59b7828488"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),

		// dkgIdMapping
		cadence.NewDictionary([]cadence.KeyValuePair{
			{
				Key:   cadence.String("0000000000000000000000000000000000000000000000000000000000000011"),
				Value: cadence.NewInt(0),
			},
		}).WithType(cadence.NewDictionaryType(cadence.StringType, cadence.IntType)),
	}).WithType(newFlowEpochEpochCommitEventType())
}

func createEpochRecoverEvent(randomSourceHex string) cadence.Event {

	clusterQCVoteDataType := newFlowClusterQCClusterQCVoteDataStructType()

	cluster1 := cadence.NewStruct([]cadence.Value{
		// aggregatedSignature
		cadence.String("b072ed22ed305acd44818a6c836e09b4e844eebde6a4fdbf5cec983e2872b86c8b0f6c34c0777bf52e385ab7c45dc55d"),
		// Node IDs of signers
		cadence.NewArray([]cadence.Value{
			cadence.String("0000000000000000000000000000000000000000000000000000000000000001"),
			cadence.String("0000000000000000000000000000000000000000000000000000000000000002"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),
	}).WithType(clusterQCVoteDataType)

	cluster2 := cadence.NewStruct([]cadence.Value{
		// aggregatedSignature
		cadence.String("899e266a543e1b3a564f68b22f7be571f2e944ec30fadc4b39e2d5f526ba044c0f3cb2648f8334fc216fa3360a0418b2"),
		// Node IDs of signers
		cadence.NewArray([]cadence.Value{
			cadence.String("0000000000000000000000000000000000000000000000000000000000000003"),
			cadence.String("0000000000000000000000000000000000000000000000000000000000000004"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),
	}).WithType(clusterQCVoteDataType)

	return cadence.NewEvent([]cadence.Value{
		// counter
		cadence.NewUInt64(1),

		// nodeInfo
		createEpochNodes(),

		// firstView
		cadence.NewUInt64(100),

		// finalView
		cadence.NewUInt64(200),

		// collectorClusters
		cadence.NewArray([]cadence.Value{
			// cluster 1
			cadence.NewArray([]cadence.Value{
				cadence.String("0000000000000000000000000000000000000000000000000000000000000001"),
				cadence.String("0000000000000000000000000000000000000000000000000000000000000002"),
			}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),
			// cluster 2
			cadence.NewArray([]cadence.Value{
				cadence.String("0000000000000000000000000000000000000000000000000000000000000003"),
				cadence.String("0000000000000000000000000000000000000000000000000000000000000004"),
			}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.NewVariableSizedArrayType(cadence.StringType))),

		// randomSource
		cadence.String(randomSourceHex),

		// DKGPhase1FinalView
		cadence.UInt64(150),

		// DKGPhase2FinalView
		cadence.UInt64(160),

		// DKGPhase3FinalView
		cadence.UInt64(170),

		// targetDuration
		cadence.UInt64(200),

		// targetEndTime
		cadence.UInt64(2000000000),

		// clusterQCs
		cadence.NewArray([]cadence.Value{
			// cluster 1
			cluster1,
			// cluster 2
			cluster2,
		}).WithType(cadence.NewVariableSizedArrayType(clusterQCVoteDataType)),

		// dkgGroupKey
		cadence.String("8c588266db5f5cda629e83f8aa04ae9413593fac19e4865d06d291c9d14fbdd9bdb86a7a12f9ef8590c79cb635e3163315d193087e9336092987150d0cd2b14ac6365f7dc93eec573752108b8c12368abb65f0652d9f644e5aed611c37926950"),

		// dkgPubKeys
		cadence.NewArray([]cadence.Value{
			cadence.String("87a339e4e5c74f089da20a33f515d8c8f4464ab53ede5a74aa2432cd1ae66d522da0c122249ee176cd747ddc83ca81090498389384201614caf51eac392c1c0a916dfdcfbbdf7363f9552b6468434add3d3f6dc91a92bbe3ee368b59b7828488"),
		}).WithType(cadence.NewVariableSizedArrayType(cadence.StringType)),

		// dkgIdMapping
		cadence.NewDictionary([]cadence.KeyValuePair{
			{
				Key:   cadence.String("0000000000000000000000000000000000000000000000000000000000000011"),
				Value: cadence.NewInt(0),
			},
		}).WithType(cadence.NewDictionaryType(cadence.StringType, cadence.IntType)),
	}).WithType(newFlowEpochEpochRecoverEventType())
}

func createVersionBeaconEvent() cadence.Event {
	versionBoundaryType := NewNodeVersionBeaconVersionBoundaryStructType()

	semverType := NewNodeVersionBeaconSemverStructType()

	semver := cadence.NewStruct([]cadence.Value{
		// major
		cadence.UInt8(2),

		// minor
		cadence.UInt8(13),

		// patch
		cadence.UInt8(7),

		// preRelease
		cadence.NewOptional(cadence.String("test")),
	}).WithType(semverType)

	versionBoundary := cadence.NewStruct([]cadence.Value{
		// blockHeight
		cadence.UInt64(44),

		// version
		semver,
	}).WithType(versionBoundaryType)

	return cadence.NewEvent([]cadence.Value{
		// versionBoundaries
		cadence.NewArray([]cadence.Value{
			versionBoundary,
		}).WithType(cadence.NewVariableSizedArrayType(versionBoundaryType)),

		// sequence
		cadence.UInt64(5),
	}).WithType(NewNodeVersionBeaconVersionBeaconEventType())
}

func createProtocolStateVersionUpgradeEvent() cadence.Event {
	newVersion := cadence.NewUInt64(1)
	activeView := cadence.NewUInt64(1000)

	return cadence.NewEvent([]cadence.Value{
		newVersion,
		activeView,
	}).WithType(NewProtocolStateVersionUpgradeEventType())
}

func newFlowClusterQCVoteStructType() cadence.Type {

	// A.01cf0e2f2f715450.FlowClusterQC.Vote

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "FlowClusterQC")

	return cadence.NewStructType(
		location,
		"FlowClusterQC.Vote",
		[]cadence.Field{
			{
				Identifier: "nodeID",
				Type:       cadence.StringType,
			},
			{
				Identifier: "signature",
				Type:       cadence.NewOptionalType(cadence.StringType),
			},
			{
				Identifier: "message",
				Type:       cadence.NewOptionalType(cadence.StringType),
			},
			{
				Identifier: "clusterIndex",
				Type:       cadence.UInt16Type,
			},
			{
				Identifier: "weight",
				Type:       cadence.UInt64Type,
			},
		},
		nil,
	)
}

func newFlowIDTableStakingNodeInfoStructType() *cadence.StructType {

	// A.01cf0e2f2f715450.FlowIDTableStaking.NodeInfo

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "FlowIDTableStaking")

	return cadence.NewStructType(
		location,
		"FlowIDTableStaking.NodeInfo",
		[]cadence.Field{
			{
				Identifier: "id",
				Type:       cadence.StringType,
			},
			{
				Identifier: "role",
				Type:       cadence.UInt8Type,
			},
			{
				Identifier: "networkingAddress",
				Type:       cadence.StringType,
			},
			{
				Identifier: "networkingKey",
				Type:       cadence.StringType,
			},
			{
				Identifier: "stakingKey",
				Type:       cadence.StringType,
			},
			{
				Identifier: "tokensStaked",
				Type:       cadence.UFix64Type,
			},
			{
				Identifier: "tokensCommitted",
				Type:       cadence.UFix64Type,
			},
			{
				Identifier: "tokensUnstaking",
				Type:       cadence.UFix64Type,
			},
			{
				Identifier: "tokensUnstaked",
				Type:       cadence.UFix64Type,
			},
			{
				Identifier: "tokensRewarded",
				Type:       cadence.UFix64Type,
			},
			{
				Identifier: "delegators",
				Type:       cadence.NewVariableSizedArrayType(cadence.UInt32Type),
			},
			{
				Identifier: "delegatorIDCounter",
				Type:       cadence.UInt32Type,
			},
			{
				Identifier: "tokensRequestedToUnstake",
				Type:       cadence.UFix64Type,
			},
			{
				Identifier: "initialWeight",
				Type:       cadence.UInt64Type,
			},
		},
		nil,
	)
}

func newFlowEpochEpochSetupEventType() *cadence.EventType {

	// A.01cf0e2f2f715450.FlowEpoch.EpochSetup

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "FlowEpoch")

	return cadence.NewEventType(
		location,
		"FlowEpoch.EpochSetup",
		[]cadence.Field{
			{
				Identifier: "counter",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "nodeInfo",
				Type:       cadence.NewVariableSizedArrayType(newFlowIDTableStakingNodeInfoStructType()),
			},
			{
				Identifier: "firstView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "finalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "collectorClusters",
				Type:       cadence.NewVariableSizedArrayType(NewFlowClusterQCClusterStructType()),
			},
			{
				Identifier: "randomSource",
				Type:       cadence.StringType,
			},
			{
				Identifier: "DKGPhase1FinalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "DKGPhase2FinalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "DKGPhase3FinalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "targetDuration",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "targetEndTime",
				Type:       cadence.UInt64Type,
			},
		},
		nil,
	)
}

func newFlowEpochEpochCommitEventType() *cadence.EventType {

	// A.01cf0e2f2f715450.FlowEpoch.EpochCommit

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "FlowEpoch")

	return cadence.NewEventType(
		location,
		"FlowEpoch.EpochCommit",
		[]cadence.Field{
			{
				Identifier: "counter",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "clusterQCs",
				Type:       cadence.NewVariableSizedArrayType(newFlowClusterQCClusterQCStructType()),
			},
			{
				Identifier: "dkgGroupKey",
				Type:       cadence.StringType,
			},
			{
				Identifier: "dkgPubKeys",
				Type:       cadence.NewVariableSizedArrayType(cadence.StringType),
			},
			{
				Identifier: "dkgIdMapping",
				Type:       cadence.NewDictionaryType(cadence.StringType, cadence.IntType),
			},
		},
		nil,
	)
}

func newFlowEpochEpochRecoverEventType() *cadence.EventType {

	// A.01cf0e2f2f715450.FlowEpoch.EpochRecover

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "FlowEpoch")

	return cadence.NewEventType(
		location,
		"FlowEpoch.EpochRecover",
		[]cadence.Field{
			{
				Identifier: "counter",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "nodeInfo",
				Type:       cadence.NewVariableSizedArrayType(newFlowIDTableStakingNodeInfoStructType()),
			},
			{
				Identifier: "firstView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "finalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "clusterAssignments",
				Type:       cadence.NewVariableSizedArrayType(cadence.NewVariableSizedArrayType(cadence.StringType)),
			},
			{
				Identifier: "randomSource",
				Type:       cadence.StringType,
			},
			{
				Identifier: "DKGPhase1FinalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "DKGPhase2FinalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "DKGPhase3FinalView",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "targetDuration",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "targetEndTime",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "clusterQCVoteData",
				Type:       cadence.NewVariableSizedArrayType(newFlowClusterQCClusterQCVoteDataStructType()),
			},
			{
				Identifier: "dkgGroupKey",
				Type:       cadence.StringType,
			},
			{
				Identifier: "dkgPubKeys",
				Type:       cadence.NewVariableSizedArrayType(cadence.StringType),
			},
			{
				Identifier: "dkgIdMapping",
				Type:       cadence.NewDictionaryType(cadence.StringType, cadence.IntType),
			},
		},
		nil,
	)
}

func newFlowClusterQCClusterQCStructType() *cadence.StructType {

	// A.01cf0e2f2f715450.FlowClusterQC.ClusterQC"

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "FlowClusterQC")

	return cadence.NewStructType(
		location,
		"FlowClusterQC.ClusterQC",
		[]cadence.Field{
			{
				Identifier: "index",
				Type:       cadence.UInt16Type,
			},
			{
				Identifier: "voteSignatures",
				Type:       cadence.NewVariableSizedArrayType(cadence.StringType),
			},
			{
				Identifier: "voteMessage",
				Type:       cadence.StringType,
			},
			{
				Identifier: "voterIDs",
				Type:       cadence.NewVariableSizedArrayType(cadence.StringType),
			},
		},
		nil,
	)
}

func newFlowClusterQCClusterQCVoteDataStructType() *cadence.StructType {

	// A.01cf0e2f2f715450.FlowClusterQC.ClusterQCVoteData"

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "ClusterQCVoteData")

	return cadence.NewStructType(
		location,
		"FlowClusterQC.ClusterQCVoteData",
		[]cadence.Field{
			{
				Identifier: "aggregatedSignature",
				Type:       cadence.StringType,
			},
			{
				Identifier: "voterIDs",
				Type:       cadence.NewVariableSizedArrayType(cadence.StringType),
			},
		},
		nil,
	)
}

func NewNodeVersionBeaconVersionBeaconEventType() *cadence.EventType {

	// A.01cf0e2f2f715450.NodeVersionBeacon.VersionBeacon

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "NodeVersionBeacon")

	return cadence.NewEventType(
		location,
		"NodeVersionBeacon.VersionBeacon",
		[]cadence.Field{
			{
				Identifier: "versionBoundaries",
				Type:       cadence.NewVariableSizedArrayType(NewNodeVersionBeaconVersionBoundaryStructType()),
			},
			{
				Identifier: "sequence",
				Type:       cadence.UInt64Type,
			},
		},
		nil,
	)
}

func NewNodeVersionBeaconVersionBoundaryStructType() *cadence.StructType {

	// A.01cf0e2f2f715450.NodeVersionBeacon.VersionBoundary

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "NodeVersionBeacon")

	return cadence.NewStructType(
		location,
		"NodeVersionBeacon.VersionBoundary",
		[]cadence.Field{
			{
				Identifier: "blockHeight",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "version",
				Type:       NewNodeVersionBeaconSemverStructType(),
			},
		},
		nil,
	)
}

func NewNodeVersionBeaconSemverStructType() *cadence.StructType {

	// A.01cf0e2f2f715450.NodeVersionBeacon.Semver

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "NodeVersionBeacon")

	return cadence.NewStructType(
		location,
		"NodeVersionBeacon.Semver",
		[]cadence.Field{
			{
				Identifier: "major",
				Type:       cadence.UInt8Type,
			},
			{
				Identifier: "minor",
				Type:       cadence.UInt8Type,
			},
			{
				Identifier: "patch",
				Type:       cadence.UInt8Type,
			},
			{
				Identifier: "preRelease",
				Type:       cadence.NewOptionalType(cadence.StringType),
			},
		},
		nil,
	)
}

func NewProtocolStateVersionUpgradeEventType() *cadence.EventType {

	// A.01cf0e2f2f715450.NodeVersionBeacon.ProtocolStateVersionUpgrade

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "ProtocolStateVersionUpgrade")

	return cadence.NewEventType(
		location,
		"NodeVersionBeacon.ProtocolStateVersionUpgrade",
		[]cadence.Field{
			{
				Identifier: "newProtocolVersion",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "activeView",
				Type:       cadence.UInt64Type,
			},
		},
		nil,
	)
}

func ufix64FromString(s string) cadence.UFix64 {
	f, err := cadence.NewUFix64(s)
	if err != nil {
		panic(err)
	}
	return f
}

func EpochSetupFixtureCCF(randomSource []byte) []byte {
	randomSourceHex := hex.EncodeToString(randomSource)
	b, err := ccf.Encode(createEpochSetupEvent(randomSourceHex))
	if err != nil {
		panic(err)
	}
	_, err = ccf.Decode(nil, b)
	if err != nil {
		panic(err)
	}
	return b
}

func EpochSetupCCFWithNonHexRandomSource() []byte {
	// randomSource of correct length but made of non hex characters
	randomSource := "ZZ"
	for len(randomSource) != 2*flow.EpochSetupRandomSourceLength {
		randomSource = randomSource + "aa"
	}

	event := createEpochSetupEvent(randomSource)

	b, err := ccf.Encode(event)
	if err != nil {
		panic(err)
	}
	_, err = ccf.Decode(nil, b)
	if err != nil {
		panic(err)
	}
	return b
}

var EpochCommitFixtureCCF = func() []byte {
	b, err := ccf.Encode(createEpochCommitEvent())
	if err != nil {
		panic(err)
	}
	_, err = ccf.Decode(nil, b)
	if err != nil {
		panic(err)
	}
	return b
}()

func EpochRecoverFixtureCCF(randomSource []byte) []byte {
	randomSourceHex := hex.EncodeToString(randomSource)
	b, err := ccf.Encode(createEpochRecoverEvent(randomSourceHex))
	if err != nil {
		panic(err)
	}
	_, err = ccf.Decode(nil, b)
	if err != nil {
		panic(err)
	}
	return b
}

var VersionBeaconFixtureCCF = func() []byte {
	b, err := ccf.Encode(createVersionBeaconEvent())
	if err != nil {
		panic(err)
	}
	_, err = ccf.Decode(nil, b)
	if err != nil {
		panic(err)
	}
	return b
}()

var ProtocolStateVersionUpgradeFixtureCCF = func() []byte {
	b, err := ccf.Encode(createProtocolStateVersionUpgradeEvent())
	if err != nil {
		panic(err)
	}
	_, err = ccf.Decode(nil, b)
	if err != nil {
		panic(err)
	}
	return b
}()

func VerifyCdcArguments(t *testing.T, expected []cadence.Value, actual []interface{}) {

	for index, arg := range actual {

		// marshal to bytes
		bz, err := json.Marshal(arg)
		require.NoError(t, err)

		// parse cadence value
		decoded, err := jsoncdc.Decode(nil, bz)
		require.NoError(t, err)

		assert.Equal(t, expected[index], decoded)
	}
}

// InterfafceToCdcValues decodes jsoncdc encoded values from interface -> cadence value.
func InterfafceToCdcValues(t *testing.T, vals []interface{}) []cadence.Value {
	decoded := make([]cadence.Value, len(vals))
	for index, val := range vals {

		// marshal to bytes
		bz, err := json.Marshal(val)
		require.NoError(t, err)

		// parse cadence value
		cdcVal, err := jsoncdc.Decode(nil, bz)
		require.NoError(t, err)

		decoded[index] = cdcVal
	}

	return decoded
}

func NewFlowClusterQCClusterStructType() *cadence.StructType {

	// A.01cf0e2f2f715450.FlowClusterQC.Cluster

	address, _ := common.HexToAddress("01cf0e2f2f715450")
	location := common.NewAddressLocation(nil, address, "FlowClusterQC")

	return cadence.NewStructType(
		location,
		"FlowClusterQC.Cluster",
		[]cadence.Field{
			{
				Identifier: "index",
				Type:       cadence.UInt16Type,
			},
			{
				Identifier: "nodeWeights",
				Type:       cadence.NewDictionaryType(cadence.StringType, cadence.UInt64Type),
			},
			{
				Identifier: "totalWeight",
				Type:       cadence.UInt64Type,
			},
			{
				Identifier: "generatedVotes",
				Type:       cadence.NewDictionaryType(cadence.StringType, newFlowClusterQCVoteStructType()),
			},
			{
				Identifier: "uniqueVoteMessageTotalWeights",
				Type:       cadence.NewDictionaryType(cadence.StringType, cadence.UInt64Type),
			},
		},
		nil,
	)
}
