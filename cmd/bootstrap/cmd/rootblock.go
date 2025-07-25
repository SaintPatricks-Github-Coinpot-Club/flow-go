package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-go/cmd"
	"github.com/onflow/flow-go/cmd/bootstrap/run"
	"github.com/onflow/flow-go/cmd/util/cmd/common"
	model "github.com/onflow/flow-go/model/bootstrap"
	"github.com/onflow/flow-go/model/dkg"
	"github.com/onflow/flow-go/model/encodable"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/epochs"
	"github.com/onflow/flow-go/state/protocol"
	"github.com/onflow/flow-go/state/protocol/inmem"
	"github.com/onflow/flow-go/state/protocol/protocol_state/kvstore"
)

var (
	flagRootChain     string
	flagRootParent    string
	flagRootHeight    uint64
	flagRootTimestamp string
	// Deprecated: Replaced by ProtocolStateVersion
	// Historically, this flag set a spork-scoped version number, by convention equal to the major software version.
	// Now that we have HCUs which change the major software version mid-spork, this is no longer useful.
	deprecatedFlagProtocolVersion   uint
	flagFinalizationSafetyThreshold uint64
	flagEpochExtensionViewCount     uint64
	flagCollectionClusters          uint
	flagEpochCounter                uint64
	flagNumViewsInEpoch             uint64
	flagNumViewsInStakingAuction    uint64
	flagNumViewsInDKGPhase          uint64
	// Epoch target end time config
	flagUseDefaultEpochTargetEndTime bool
	flagEpochTimingRefCounter        uint64
	flagEpochTimingRefTimestamp      uint64
	flagEpochTimingDuration          uint64
)

// rootBlockCmd represents the rootBlock command
var rootBlockCmd = &cobra.Command{
	Use:   "rootblock",
	Short: "Generate root block data",
	Long:  `Run Beacon KeyGen, generate root block and votes for root block needed for constructing QC. Serialize all info into file`,
	Run:   rootBlock,
}

func init() {
	rootCmd.AddCommand(rootBlockCmd)
	addRootBlockCmdFlags()
}

func addRootBlockCmdFlags() {
	// required parameters for network configuration and generation of root node identities
	rootBlockCmd.Flags().StringVar(&flagConfig, "config", "",
		"path to a JSON file containing multiple node configurations (fields Role, Address, Weight)")
	rootBlockCmd.Flags().StringVar(&flagInternalNodePrivInfoDir, "internal-priv-dir", "", "path to directory "+
		"containing the output from the `keygen` command for internal nodes")
	rootBlockCmd.Flags().StringVar(&flagPartnerNodeInfoDir, "partner-dir", "", "path to directory "+
		"containing one JSON file starting with node-info.pub.<NODE_ID>.json for every partner node (fields "+
		" in the JSON file: Role, Address, NodeID, NetworkPubKey, StakingPubKey)")
	rootBlockCmd.Flags().StringVar(&deprecatedFlagPartnerStakes, "partner-stakes", "", "deprecated: use --partner-weights")
	rootBlockCmd.Flags().StringVar(&flagPartnerWeights, "partner-weights", "", "path to a JSON file containing "+
		"a map from partner node's NodeID to their stake")

	cmd.MarkFlagRequired(rootBlockCmd, "config")
	cmd.MarkFlagRequired(rootBlockCmd, "internal-priv-dir")
	cmd.MarkFlagRequired(rootBlockCmd, "partner-dir")
	cmd.MarkFlagRequired(rootBlockCmd, "partner-weights")

	// required parameters for generation of epoch setup and commit events
	rootBlockCmd.Flags().Uint64Var(&flagEpochCounter, "epoch-counter", 0, "epoch counter for the epoch beginning with the root block")
	rootBlockCmd.Flags().Uint64Var(&flagNumViewsInEpoch, "epoch-length", 4000, "length of each epoch measured in views")
	rootBlockCmd.Flags().Uint64Var(&flagNumViewsInStakingAuction, "epoch-staking-phase-length", 100, "length of the epoch staking phase measured in views")
	rootBlockCmd.Flags().Uint64Var(&flagNumViewsInDKGPhase, "epoch-dkg-phase-length", 1000, "length of each DKG phase measured in views")

	// optional parameters to influence various aspects of identity generation
	rootBlockCmd.Flags().UintVar(&flagCollectionClusters, "collection-clusters", 2, "number of collection clusters")

	cmd.MarkFlagRequired(rootBlockCmd, "epoch-counter")
	cmd.MarkFlagRequired(rootBlockCmd, "epoch-length")
	cmd.MarkFlagRequired(rootBlockCmd, "epoch-staking-phase-length")
	cmd.MarkFlagRequired(rootBlockCmd, "epoch-dkg-phase-length")

	// required parameters for generation of root block, root execution result and root block seal
	rootBlockCmd.Flags().StringVar(&flagRootChain, "root-chain", "local", "chain ID for the root block (can be 'main', 'test', 'sandbox', 'preview', 'bench', or 'local'")
	rootBlockCmd.Flags().StringVar(&flagRootParent, "root-parent", "0000000000000000000000000000000000000000000000000000000000000000", "ID for the parent of the root block")
	rootBlockCmd.Flags().Uint64Var(&flagRootHeight, "root-height", 0, "height of the root block")
	rootBlockCmd.Flags().StringVar(&flagRootTimestamp, "root-timestamp", time.Now().UTC().Format(time.RFC3339), "timestamp of the root block (RFC3339)")
	rootBlockCmd.Flags().UintVar(&deprecatedFlagProtocolVersion, "protocol-version", 0, "deprecated: this flag will be ignored and remove in a future release")
	rootBlockCmd.Flags().Uint64Var(&flagFinalizationSafetyThreshold, "finalization-safety-threshold", 500, "defines finalization safety threshold")
	rootBlockCmd.Flags().Uint64Var(&flagEpochExtensionViewCount, "epoch-extension-view-count", 100_000, "length of epoch extension in views, default is 100_000 which is approximately 1 day")

	cmd.MarkFlagRequired(rootBlockCmd, "root-chain")
	cmd.MarkFlagRequired(rootBlockCmd, "root-parent")
	cmd.MarkFlagRequired(rootBlockCmd, "root-height")
	cmd.MarkFlagRequired(rootBlockCmd, "finalization-safety-threshold")
	cmd.MarkFlagRequired(rootBlockCmd, "epoch-extension-view-count")

	// Epoch timing config - these values must be set identically to `EpochTimingConfig` in the FlowEpoch smart contract.
	// See https://github.com/onflow/flow-core-contracts/blob/240579784e9bb8d97d91d0e3213614e25562c078/contracts/epochs/FlowEpoch.cdc#L259-L266
	// Must specify either:
	//   1. --use-default-epoch-timing and no other `--epoch-timing*` flags
	//   2. All `--epoch-timing*` flags except --use-default-epoch-timing
	//
	// Use Option 1 for Benchnet, Localnet, etc.
	// Use Option 2 for Mainnet, Testnet, Canary.
	rootBlockCmd.Flags().BoolVar(&flagUseDefaultEpochTargetEndTime, "use-default-epoch-timing", false, "whether to use the default target end time")
	rootBlockCmd.Flags().Uint64Var(&flagEpochTimingRefCounter, "epoch-timing-ref-counter", 0, "the reference epoch for computing the root epoch's target end time")
	rootBlockCmd.Flags().Uint64Var(&flagEpochTimingRefTimestamp, "epoch-timing-ref-timestamp", 0, "the end time of the reference epoch, specified in second-precision Unix time, to use to compute the root epoch's target end time")
	rootBlockCmd.Flags().Uint64Var(&flagEpochTimingDuration, "epoch-timing-duration", 0, "the duration of each epoch in seconds, used to compute the root epoch's target end time")

	rootBlockCmd.MarkFlagsOneRequired("use-default-epoch-timing", "epoch-timing-ref-counter", "epoch-timing-ref-timestamp", "epoch-timing-duration")
	rootBlockCmd.MarkFlagsRequiredTogether("epoch-timing-ref-counter", "epoch-timing-ref-timestamp", "epoch-timing-duration")
	for _, flag := range []string{"epoch-timing-ref-counter", "epoch-timing-ref-timestamp", "epoch-timing-duration"} {
		rootBlockCmd.MarkFlagsMutuallyExclusive("use-default-epoch-timing", flag)
	}
}

func rootBlock(cmd *cobra.Command, args []string) {

	// maintain backward compatibility with old flag name
	if deprecatedFlagPartnerStakes != "" {
		log.Warn().Msg("using deprecated flag --partner-stakes (use --partner-weights instead)")
		if flagPartnerWeights == "" {
			flagPartnerWeights = deprecatedFlagPartnerStakes
		} else {
			log.Fatal().Msg("cannot use both --partner-stakes and --partner-weights flags (use only --partner-weights)")
		}
	}
	if deprecatedFlagProtocolVersion != 0 {
		log.Warn().Msg("using deprecated flag --protocol-version; please remove this flag from your workflow, it is ignored and will be removed in a future release")
	}

	// validate epoch configs
	err := validateEpochConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid or unsafe config for finalization safety threshold")
	}
	err = validateOrPopulateEpochTimingConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid epoch timing config")
	}

	// Read partner node's information and internal node's information.
	// With "internal nodes" we reference nodes, whose private keys we have. In comparison,
	// for "partner nodes" we generally do not have their keys. However, we allow some overlap,
	// in that we tolerate a configuration where information about an "internal node" is also
	// duplicated in the list of "partner nodes".
	log.Info().Msg("collecting partner network and staking keys")
	rawPartnerNodes, err := common.ReadFullPartnerNodeInfos(log, flagPartnerWeights, flagPartnerNodeInfoDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read full partner node infos")
	}
	log.Info().Msg("")

	log.Info().Msg("generating internal private networking and staking keys")
	internalNodes, err := common.ReadFullInternalNodeInfos(log, flagInternalNodePrivInfoDir, flagConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read full internal node infos")
	}

	log.Info().Msg("")

	// we now convert to the strict meaning of: "internal nodes" vs "partner nodes"
	//  • "internal nodes" we have they private keys for
	//  • "partner nodes" we don't have the keys for
	//  • both sets are disjoint (no common nodes)
	log.Info().Msg("remove internal partner nodes")
	partnerNodes := common.FilterInternalPartners(rawPartnerNodes, internalNodes)
	log.Info().Msgf("removed %d internal partner nodes", len(rawPartnerNodes)-len(partnerNodes))

	log.Info().Msg("checking constraints on consensus nodes")
	checkConstraints(partnerNodes, internalNodes)
	log.Info().Msg("")

	log.Info().Msg("assembling network and staking keys")
	stakingNodes, err := mergeNodeInfos(internalNodes, partnerNodes)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to merge node infos")
	}
	publicInfo, err := model.ToPublicNodeInfoList(stakingNodes)
	if err != nil {
		log.Fatal().Msg("failed to read public node info")
	}
	err = common.WriteJSON(model.PathNodeInfosPub, flagOutdir, publicInfo)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write json")
	}
	log.Info().Msgf("wrote file %s/%s", flagOutdir, model.PathNodeInfosPub)
	log.Info().Msg("")

	log.Info().Msg("running DKG for consensus nodes")
	randomBeaconData, dkgIndexMap := runBeaconKG(model.FilterByRole(stakingNodes, flow.RoleConsensus))
	log.Info().Msg("")

	// create flow.IdentityList representation of the participant set
	participants := model.ToIdentityList(stakingNodes).Sort(flow.Canonical[flow.Identity])

	log.Info().Msg("computing collection node clusters")
	assignments, clusters, err := common.ConstructClusterAssignment(log, model.ToIdentityList(partnerNodes), model.ToIdentityList(internalNodes), int(flagCollectionClusters))
	if err != nil {
		log.Fatal().Err(err).Msg("unable to generate cluster assignment")
	}
	log.Info().Msg("")

	log.Info().Msg("constructing root blocks for collection node clusters")
	clusterBlocks := run.GenerateRootClusterBlocks(flagEpochCounter, clusters)
	log.Info().Msg("")

	log.Info().Msg("constructing root QCs for collection node clusters")
	clusterQCs := run.ConstructRootQCsForClusters(log, clusters, internalNodes, clusterBlocks)
	log.Info().Msg("")

	log.Info().Msg("constructing root header")
	header := constructRootHeader(flagRootChain, flagRootParent, flagRootHeight, flagRootTimestamp)
	log.Info().Msg("")

	log.Info().Msg("constructing intermediary bootstrapping data")
	epochSetup, epochCommit, err := constructRootEpochEvents(header.View, participants, assignments, clusterQCs, randomBeaconData, dkgIndexMap)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to construct root epoch events")
	}
	epochConfig := generateExecutionStateEpochConfig(epochSetup, clusterQCs, randomBeaconData)
	intermediaryEpochData := IntermediaryEpochData{
		RootEpochSetup:       epochSetup,
		RootEpochCommit:      epochCommit,
		ExecutionStateConfig: epochConfig,
	}
	intermediaryParamsData := IntermediaryParamsData{
		FinalizationSafetyThreshold: flagFinalizationSafetyThreshold,
		EpochExtensionViewCount:     flagEpochExtensionViewCount,
	}
	intermediaryData := IntermediaryBootstrappingData{
		IntermediaryEpochData:  intermediaryEpochData,
		IntermediaryParamsData: intermediaryParamsData,
	}
	err = common.WriteJSON(model.PathIntermediaryBootstrappingData, flagOutdir, intermediaryData)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write json")
	}
	log.Info().Msgf("wrote file %s/%s", flagOutdir, model.PathIntermediaryBootstrappingData)
	log.Info().Msg("")

	log.Info().Msg("constructing root block")
	minEpochStateEntry, err := inmem.EpochProtocolStateFromServiceEvents(epochSetup, epochCommit)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to construct epoch protocol state")
	}

	rootProtocolState, err := kvstore.NewDefaultKVStore(
		flagFinalizationSafetyThreshold,
		flagEpochExtensionViewCount,
		minEpochStateEntry.ID(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to construct root kvstore")
	}
	block := constructRootBlock(header, rootProtocolState.ID())
	err = common.WriteJSON(model.PathRootBlockData, flagOutdir, block)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write json")
	}
	log.Info().Msgf("wrote file %s/%s", flagOutdir, model.PathRootBlockData)
	log.Info().Msg("")

	log.Info().Msg("constructing and writing votes")
	constructRootVotes(
		block,
		model.FilterByRole(stakingNodes, flow.RoleConsensus),
		model.FilterByRole(internalNodes, flow.RoleConsensus),
		randomBeaconData,
	)
	log.Info().Msg("")
}

// validateEpochConfig validates configuration of the epoch commitment deadline.
func validateEpochConfig() error {
	chainID := parseChainID(flagRootChain)
	dkgFinalView := flagNumViewsInStakingAuction + flagNumViewsInDKGPhase*3 // 3 DKG phases
	epochCommitDeadline := flagNumViewsInEpoch - flagFinalizationSafetyThreshold

	defaultEpochSafetyParams, err := protocol.DefaultEpochSafetyParams(chainID)
	if err != nil {
		return fmt.Errorf("could not get default epoch commit safety threshold: %w", err)
	}

	// sanity check: the safety threshold is >= the default for the chain
	if flagFinalizationSafetyThreshold < defaultEpochSafetyParams.FinalizationSafetyThreshold {
		return fmt.Errorf("potentially unsafe epoch config: epoch commit safety threshold smaller than expected (%d < %d)", flagFinalizationSafetyThreshold, defaultEpochSafetyParams.FinalizationSafetyThreshold)
	}
	if flagEpochExtensionViewCount < defaultEpochSafetyParams.EpochExtensionViewCount {
		return fmt.Errorf("potentially unsafe epoch config: epoch extension view count smaller than expected (%d < %d)", flagEpochExtensionViewCount, defaultEpochSafetyParams.EpochExtensionViewCount)
	}
	// sanity check: epoch commitment deadline cannot be before the DKG end
	if epochCommitDeadline <= dkgFinalView {
		return fmt.Errorf("invalid epoch config: the epoch commitment deadline (%d) is before the DKG final view (%d)", epochCommitDeadline, dkgFinalView)
	}
	// sanity check: the difference between DKG end and safety threshold is also >= the default safety threshold
	if epochCommitDeadline-dkgFinalView < defaultEpochSafetyParams.FinalizationSafetyThreshold {
		return fmt.Errorf("potentially unsafe epoch config: time between DKG end and epoch commitment deadline is smaller than expected (%d-%d < %d)",
			epochCommitDeadline, dkgFinalView, defaultEpochSafetyParams.FinalizationSafetyThreshold)
	}
	return nil
}

// generateExecutionStateEpochConfig generates epoch-related configuration used
// to generate an empty root execution state. This config is generated in the
// `rootblock` alongside the root epoch and root protocol state ID for consistency.
func generateExecutionStateEpochConfig(
	epochSetup *flow.EpochSetup,
	clusterQCs []*flow.QuorumCertificate,
	dkgData dkg.ThresholdKeySet,
) epochs.EpochConfig {

	randomSource := make([]byte, flow.EpochSetupRandomSourceLength)
	if _, err := rand.Read(randomSource); err != nil {
		log.Fatal().Err(err).Msg("failed to generate a random source")
	}
	cdcRandomSource, err := cadence.NewString(hex.EncodeToString(randomSource))
	if err != nil {
		log.Fatal().Err(err).Msg("invalid random source")
	}

	epochConfig := epochs.EpochConfig{
		EpochTokenPayout:             cadence.UFix64(0),
		RewardCut:                    cadence.UFix64(0),
		FLOWsupplyIncreasePercentage: cadence.UFix64(0),
		CurrentEpochCounter:          cadence.UInt64(epochSetup.Counter),
		NumViewsInEpoch:              cadence.UInt64(flagNumViewsInEpoch),
		NumViewsInStakingAuction:     cadence.UInt64(flagNumViewsInStakingAuction),
		NumViewsInDKGPhase:           cadence.UInt64(flagNumViewsInDKGPhase),
		NumCollectorClusters:         cadence.UInt16(flagCollectionClusters),
		RandomSource:                 cdcRandomSource,
		CollectorClusters:            epochSetup.Assignments,
		ClusterQCs:                   clusterQCs,
		DKGPubKeys:                   encodable.WrapRandomBeaconPubKeys(dkgData.PubKeyShares),
	}
	return epochConfig
}
