package convert

import (
	"encoding/hex"
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/encoding/ccf"
	"github.com/onflow/crypto"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	"github.com/onflow/flow-go/model/flow"
)

// ServiceEvent converts a service event encoded as the generic flow.Event
// type to a flow.ServiceEvent type for use within protocol software and protocol
// state. This acts as the conversion from the Cadence type to the flow-go type.
// CAUTION: This function must only be used for input events computed locally, by an
// Execution or Verification Node; it is not resilient to malicious inputs.
// No errors are expected during normal operation.
func ServiceEvent(chainID flow.ChainID, event flow.Event) (*flow.ServiceEvent, error) {
	events := systemcontracts.ServiceEventsForChain(chainID)

	// depending on type of service event construct Go type
	switch event.Type {
	case events.EpochSetup.EventType():
		return convertServiceEventEpochSetup(event)
	case events.EpochCommit.EventType():
		return convertServiceEventEpochCommit(event)
	case events.EpochRecover.EventType():
		return convertServiceEventEpochRecover(event)
	case events.VersionBeacon.EventType():
		return convertServiceEventVersionBeacon(event)
	case events.ProtocolStateVersionUpgrade.EventType():
		return convertServiceEventProtocolStateVersionUpgrade(event)
	default:
		return nil, fmt.Errorf("invalid event type: %s", event.Type)
	}
}

func getField[T cadence.Value](fields map[string]cadence.Value, fieldName string) (T, error) {
	field, ok := fields[fieldName]
	if !ok || field == nil {
		var zero T
		return zero, fmt.Errorf(
			"required field not found: %s",
			fieldName,
		)
	}

	value, ok := field.(T)
	if !ok {
		var zero T
		return zero, invalidCadenceTypeError(fieldName, field, zero)
	}

	return value, nil
}

// convertServiceEventEpochSetup converts a service event encoded as the generic
// flow.Event type to a ServiceEvent type for an EpochSetup event
// CONVENTION: in the returned `EpochSetup` event,
//   - Node identities listed in `EpochSetup.Participants` are in CANONICAL ORDER
//   - for each cluster assignment (i.e. element in `EpochSetup.Assignments`), the nodeIDs are listed in CANONICAL ORDER
//
// CAUTION: This function must only be used for input events computed locally, by an
// Execution or Verification Node; it is not resilient to malicious inputs.
// No errors are expected during normal operation.
func convertServiceEventEpochSetup(event flow.Event) (*flow.ServiceEvent, error) {
	// decode bytes using ccf
	payload, err := ccf.Decode(nil, event.Payload)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event payload: %w", err)
	}

	// NOTE: variable names prefixed with cdc represent cadence types
	cdcEvent, ok := payload.(cadence.Event)
	if !ok {
		return nil, invalidCadenceTypeError("payload", payload, cadence.Event{})
	}

	if cdcEvent.Type() == nil {
		return nil, fmt.Errorf("EpochSetup event doesn't have type")
	}

	fields := cadence.FieldsMappedByName(cdcEvent)

	const expectedFieldCount = 11
	if len(fields) < expectedFieldCount {
		return nil, fmt.Errorf(
			"insufficient fields in EpochSetup event (%d < %d)",
			len(fields),
			expectedFieldCount,
		)
	}

	// parse EpochSetup event

	counter, err := getField[cadence.UInt64](fields, "counter")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	cdcParticipants, err := getField[cadence.Array](fields, "nodeInfo")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	firstView, err := getField[cadence.UInt64](fields, "firstView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	finalView, err := getField[cadence.UInt64](fields, "finalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	cdcClusters, err := getField[cadence.Array](fields, "collectorClusters")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	randomSrcHex, err := getField[cadence.String](fields, "randomSource")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	targetDuration, err := getField[cadence.UInt64](fields, "targetDuration") // Epoch duration [seconds]
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	targetEndTimeUnix, err := getField[cadence.UInt64](fields, "targetEndTime") // Unix time [seconds]
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	dkgPhase1FinalView, err := getField[cadence.UInt64](fields, "DKGPhase1FinalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	dkgPhase2FinalView, err := getField[cadence.UInt64](fields, "DKGPhase2FinalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	dkgPhase3FinalView, err := getField[cadence.UInt64](fields, "DKGPhase3FinalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochSetup event: %w", err)
	}

	// random source from the event must be a hex string
	// containing exactly 128 bits (equivalent to 16 bytes or 32 hex characters)
	randomSource, err := hex.DecodeString(string(randomSrcHex))
	if err != nil {
		return nil, fmt.Errorf(
			"could not decode random source hex (%v): %w",
			randomSrcHex,
			err,
		)
	}

	// parse cluster assignments; returned assignments are in canonical order
	assignments, err := convertClusterAssignments(cdcClusters.Values)
	if err != nil {
		return nil, fmt.Errorf("could not convert cluster assignments: %w", err)
	}

	// parse epoch participants; returned node identities are in canonical order
	participants, err := convertParticipants(cdcParticipants.Values)
	if err != nil {
		return nil, fmt.Errorf("could not convert participants: %w", err)
	}
	setup, err := flow.NewEpochSetup(
		flow.UntrustedEpochSetup{
			Counter:            uint64(counter),
			FirstView:          uint64(firstView),
			DKGPhase1FinalView: uint64(dkgPhase1FinalView),
			DKGPhase2FinalView: uint64(dkgPhase2FinalView),
			DKGPhase3FinalView: uint64(dkgPhase3FinalView),
			FinalView:          uint64(finalView),
			Participants:       participants,
			Assignments:        assignments,
			RandomSource:       randomSource,
			TargetDuration:     uint64(targetDuration),
			TargetEndTime:      uint64(targetEndTimeUnix),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not construct epoch setup: %w", err)
	}

	// construct the service event
	serviceEvent := &flow.ServiceEvent{
		Type:  flow.ServiceEventSetup,
		Event: setup,
	}

	return serviceEvent, nil
}

// convertServiceEventEpochCommit converts a service event encoded as the generic
// flow.Event type to a ServiceEvent type for an EpochCommit event.
// CAUTION: This function must only be used for input events computed locally, by an
// Execution or Verification Node; it is not resilient to malicious inputs.
// No errors are expected during normal operation.
func convertServiceEventEpochCommit(event flow.Event) (*flow.ServiceEvent, error) {
	// decode bytes using ccf
	payload, err := ccf.Decode(nil, event.Payload)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event payload: %w", err)
	}

	cdcEvent, ok := payload.(cadence.Event)
	if !ok {
		return nil, invalidCadenceTypeError("payload", payload, cadence.Event{})
	}

	if cdcEvent.Type() == nil {
		return nil, fmt.Errorf("EpochCommit event doesn't have type")
	}

	fields := cadence.FieldsMappedByName(cdcEvent)

	const expectedFieldCount = 5
	if len(fields) < expectedFieldCount {
		return nil, fmt.Errorf(
			"insufficient fields in EpochCommit event (%d < %d)",
			len(fields),
			expectedFieldCount,
		)
	}

	// Extract EpochCommit event fields

	counter, err := getField[cadence.UInt64](fields, "counter")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochCommit event: %w", err)
	}

	cdcClusterQCVotes, err := getField[cadence.Array](fields, "clusterQCs")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochCommit event: %w", err)
	}

	cdcDKGKeys, err := getField[cadence.Array](fields, "dkgPubKeys")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochCommit event: %w", err)
	}

	cdcDKGGroupKey, err := getField[cadence.String](fields, "dkgGroupKey")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochCommit event: %w", err)
	}

	cdcDKGIndexMap, err := getField[cadence.Dictionary](fields, "dkgIdMapping")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochCommit event: %w", err)
	}

	// parse cluster qc votes
	clusterQCs, err := convertClusterQCVotes(cdcClusterQCVotes.Values)
	if err != nil {
		return nil, fmt.Errorf("could not convert cluster qc votes: %w", err)
	}

	// parse DKG participants
	dKGParticipantKeys, err := convertDKGKeys(cdcDKGKeys.Values)
	if err != nil {
		return nil, fmt.Errorf("could not convert Random Beacon keys: %w", err)
	}

	// parse DKG group key
	dKGGroupKey, err := convertDKGKey(cdcDKGGroupKey)
	if err != nil {
		return nil, fmt.Errorf("could not convert Random Beacon group key: %w", err)
	}

	// parse DKG Index Map
	//
	// CAUTION: When the Execution or Verification Node serializes the EpochCommit to compute its ID, the DKGIndexMap
	// is converted from a map to a slice. This is necessary because maps don't have a deterministic order. For *valid*
	// EpochCommit events, the following convention holds (see DKGIndexMap type declaration for details):
	//   - For the DKG committee 𝒟, its size is n = |𝒟| = len(DKGIndexMap).
	//   - The values in DKGIndexMap must form the set {0, 1, …, n-1}, as required by the low level cryptography
	//     module (convention simplifying the implementation).
	// Therefore, a valid `DKGIndexMap` can always be represented as an `IdentifierList` slice `s` such that
	// nodeID := s[i] for i ∈ {0, …, n-1} corresponds to a key value pair (nodeID, i) in DKGIndexMap. The
	// `EpochCommit.EncodeRLP` method performs this conversion (and panics when the convention is violated).
	//    Generally, execution should be permissive and forward all system events to the Protocol State, which then
	// performs comprehensive validity checks and decides whether events are accepted or rejected. However, we can
	// only compute the ID of an EpochCommit whose DKGIndexMap satisfies the convention above. Furthermore, we do
	// not want to depend on the System Smart Contracts to _always_ emit valid DKGIndexMap - especially for Epoch
	// Recovery, where humans provide some of the parameters in the EpochCommit event.
	//    Therefore, we check here that DKGIndexMap satisfies the convention required by `EncodeRLP` and error
	// otherwise. When we error here, the corresponding event will just be omitted from `ExecutionResult.ServiceEvents`.
	// (In contrast, erroring during the ID computation will result in an irrecoverable execution halt, because the
	// ExecutionResult has already been fully constructed, but can't be broadcast).
	//    We will only drop service events whose DKGIndexMap is invalid. As the Protocol State will anyway discard
	// such events, it is fine to not relay them in the first place.
	dKGIndexMap := make(flow.DKGIndexMap, len(cdcDKGIndexMap.Pairs))
	for _, pair := range cdcDKGIndexMap.Pairs {
		nodeID, err := flow.HexStringToIdentifier(string(pair.Key.(cadence.String)))
		if err != nil {
			return nil, fmt.Errorf("failed to decode flow.Identifer in DKGIndexMap entry from EpochRecover event: %w", err)
		}
		index := pair.Value.(cadence.Int).Int()
		dKGIndexMap[nodeID] = index
	}

	commit, err := flow.NewEpochCommit(
		flow.UntrustedEpochCommit{
			Counter:            uint64(counter),
			ClusterQCs:         clusterQCs,
			DKGGroupKey:        dKGGroupKey,
			DKGParticipantKeys: dKGParticipantKeys,
			DKGIndexMap:        dKGIndexMap,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not construct epoch commit: %w", err)
	}

	// create the service event
	serviceEvent := &flow.ServiceEvent{
		Type:  flow.ServiceEventCommit,
		Event: commit,
	}

	return serviceEvent, nil
}

// convertServiceEventEpochRecover converts a service event encoded as the generic
// flow.Event type to a ServiceEvent type for an EpochRecover event.
// CAUTION: This function must only be used for input events computed locally, by an
// Execution or Verification Node; it is not resilient to malicious inputs.
// No errors are expected during normal operation.
func convertServiceEventEpochRecover(event flow.Event) (*flow.ServiceEvent, error) {
	// decode bytes using ccf
	payload, err := ccf.Decode(nil, event.Payload)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event payload: %w", err)
	}

	// NOTE: variable names prefixed with cdc represent cadence types
	cdcEvent, ok := payload.(cadence.Event)
	if !ok {
		return nil, invalidCadenceTypeError("payload", payload, cadence.Event{})
	}

	if cdcEvent.Type() == nil {
		return nil, fmt.Errorf("EpochRecover event doesn't have type")
	}

	fields := cadence.FieldsMappedByName(cdcEvent)

	const expectedFieldCount = 15
	if len(fields) < expectedFieldCount {
		return nil, fmt.Errorf(
			"insufficient fields in EpochRecover event (%d < %d)",
			len(fields),
			expectedFieldCount,
		)
	}

	// parse EpochRecover event

	counter, err := getField[cadence.UInt64](fields, "counter")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	cdcParticipants, err := getField[cadence.Array](fields, "nodeInfo")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	firstView, err := getField[cadence.UInt64](fields, "firstView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	finalView, err := getField[cadence.UInt64](fields, "finalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	cdcClusters, err := getField[cadence.Array](fields, "clusterAssignments")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	randomSrcHex, err := getField[cadence.String](fields, "randomSource")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	targetDuration, err := getField[cadence.UInt64](fields, "targetDuration") // Epoch duration [seconds]
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	targetEndTimeUnix, err := getField[cadence.UInt64](fields, "targetEndTime") // Unix time [seconds]
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	dkgPhase1FinalView, err := getField[cadence.UInt64](fields, "DKGPhase1FinalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	dkgPhase2FinalView, err := getField[cadence.UInt64](fields, "DKGPhase2FinalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	dkgPhase3FinalView, err := getField[cadence.UInt64](fields, "DKGPhase3FinalView")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	cdcClusterQCVoteData, err := getField[cadence.Array](fields, "clusterQCVoteData")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	cdcDKGKeys, err := getField[cadence.Array](fields, "dkgPubKeys")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	cdcDKGGroupKey, err := getField[cadence.String](fields, "dkgGroupKey")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	cdcDKGIndexMap, err := getField[cadence.Dictionary](fields, "dkgIdMapping")
	if err != nil {
		return nil, fmt.Errorf("failed to decode EpochRecover event: %w", err)
	}

	// random source from the event must be a hex string
	// containing exactly 128 bits (equivalent to 16 bytes or 32 hex characters)
	randomSource, err := hex.DecodeString(string(randomSrcHex))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decode random source hex (%v) from EpochRecover event: %w",
			randomSrcHex,
			err,
		)
	}

	// parse cluster assignments; returned assignments are in canonical order
	assignments, err := convertEpochRecoverCollectorClusterAssignments(cdcClusters.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cluster assignments from EpochRecover event: %w", err)
	}

	// parse epoch participants; returned node identities are in canonical order
	participants, err := convertParticipants(cdcParticipants.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to convert participants from EpochRecover event: %w", err)
	}

	setup, err := flow.NewEpochSetup(
		flow.UntrustedEpochSetup{
			Counter:            uint64(counter),
			FirstView:          uint64(firstView),
			DKGPhase1FinalView: uint64(dkgPhase1FinalView),
			DKGPhase2FinalView: uint64(dkgPhase2FinalView),
			DKGPhase3FinalView: uint64(dkgPhase3FinalView),
			FinalView:          uint64(finalView),
			Participants:       participants,
			Assignments:        assignments,
			RandomSource:       randomSource,
			TargetDuration:     uint64(targetDuration),
			TargetEndTime:      uint64(targetEndTimeUnix),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not construct epoch setup: %w", err)
	}

	// parse cluster qc votes
	clusterQCs, err := convertClusterQCVoteData(cdcClusterQCVoteData.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to decode clusterQCVoteData from EpochRecover event: %w", err)
	}

	// parse DKG participants
	dKGParticipantKeys, err := convertDKGKeys(cdcDKGKeys.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Random Beacon key shares from EpochRecover event: %w", err)
	}

	// parse DKG group key
	dKGGroupKey, err := convertDKGKey(cdcDKGGroupKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Random Beacon group key from EpochRecover event: %w", err)
	}

	// parse DKG Index Map
	//
	// CAUTION: When the Execution or Verification Node serializes the EpochCommit to compute its ID, the DKGIndexMap
	// is converted from a map to a slice. This is necessary because maps don't have a deterministic order. For *valid*
	// EpochCommit events, the following convention holds (see DKGIndexMap type declaration for details):
	//   - For the DKG committee 𝒟, its size is n = |𝒟| = len(DKGIndexMap).
	//   - The values in DKGIndexMap must form the set {0, 1, …, n-1}, as required by the low level cryptography
	//     module (convention simplifying the implementation).
	// Therefore, a valid `DKGIndexMap` can always be represented as an `IdentifierList` slice `s` such that
	// nodeID := s[i] for i ∈ {0, …, n-1} corresponds to a key value pair (nodeID, i) in DKGIndexMap. The
	// `EpochCommit.EncodeRLP` method performs this conversion (and panics when the convention is violated).
	//    Generally, execution should be permissive and forward all system events to the Protocol State, which then
	// performs comprehensive validity checks and decides whether events are accepted or rejected. However, we can
	// only compute the ID of an EpochCommit whose DKGIndexMap satisfies the convention above. Furthermore, we do
	// not want to depend on the System Smart Contracts to _always_ emit valid DKGIndexMap - especially for Epoch
	// Recovery, where humans provide some of the parameters in the EpochCommit event.
	//    Therefore, we check here that DKGIndexMap satisfies the convention required by `EncodeRLP` and error
	// otherwise. When we error here, the corresponding event will just be omitted from `ExecutionResult.ServiceEvents`.
	// (In contrast, erroring during the ID computation will result in an irrecoverable execution halt, because the
	// ExecutionResult has already been fully constructed, but can't be broadcast).
	//    We will only drop service events whose DKGIndexMap is invalid. As the Protocol State will anyway discard
	// such events, it is fine to not relay them in the first place.
	dKGIndexMap := make(flow.DKGIndexMap, len(cdcDKGIndexMap.Pairs))
	for _, pair := range cdcDKGIndexMap.Pairs {
		nodeID, err := flow.HexStringToIdentifier(string(pair.Key.(cadence.String)))
		if err != nil {
			return nil, fmt.Errorf("failed to decode flow.Identifer in DKGIndexMap entry from EpochRecover event: %w", err)
		}
		index := pair.Value.(cadence.Int).Int()
		dKGIndexMap[nodeID] = index
	}

	commit, err := flow.NewEpochCommit(
		flow.UntrustedEpochCommit{
			Counter:            uint64(counter),
			ClusterQCs:         clusterQCs,
			DKGGroupKey:        dKGGroupKey,
			DKGParticipantKeys: dKGParticipantKeys,
			DKGIndexMap:        dKGIndexMap,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not construct epoch commit: %w", err)
	}

	// create the service event
	epochRecover, err := flow.NewEpochRecover(
		flow.UntrustedEpochRecover{
			EpochSetup:  *setup,
			EpochCommit: *commit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not construct epoch recover: %w", err)
	}

	serviceEvent := &flow.ServiceEvent{
		Type:  flow.ServiceEventRecover,
		Event: epochRecover,
	}

	return serviceEvent, nil
}

// convertEpochRecoverCollectorClusterAssignments converts collector cluster assignments for EpochRecover event.
// This is a simplified version compared to the `convertClusterAssignments` function since we are dealing with
// a list of participants that don't need to be ordered by index or node weights.
// No errors are expected during normal operation.
func convertEpochRecoverCollectorClusterAssignments(cdcClusters []cadence.Value) (flow.AssignmentList, error) {
	// parse cluster assignments to Go types
	clusterAssignments := make([]flow.IdentifierList, 0, len(cdcClusters))
	// we are dealing with a nested array where each element is a list of node IDs,
	// this way we represent the cluster assignments.
	for _, value := range cdcClusters {
		cdcCluster, ok := value.(cadence.Array)
		if !ok {
			return nil, invalidCadenceTypeError("collectorClusters[i]", cdcCluster, cadence.Array{})
		}

		clusterMembers := make(flow.IdentifierList, 0, len(cdcCluster.Values))
		for _, cdcClusterParticipant := range cdcCluster.Values {
			nodeIDString, ok := cdcClusterParticipant.(cadence.String)
			if !ok {
				return nil, invalidCadenceTypeError(
					"collectorClusters[i][j]",
					cdcClusterParticipant,
					cadence.String(""),
				)
			}

			nodeID, err := flow.HexStringToIdentifier(string(nodeIDString))
			if err != nil {
				return nil, fmt.Errorf(
					"could not convert hex string to identifer: %w",
					err,
				)
			}
			clusterMembers = append(clusterMembers, nodeID)
		}

		// IMPORTANT: for each cluster, node IDs must be in *canonical order*
		clusterAssignments = append(clusterAssignments, clusterMembers.Sort(flow.IdentifierCanonical))
	}

	return clusterAssignments, nil
}

// convertClusterAssignments converts the Cadence representation of cluster
// assignments included in the EpochSetup into the protocol AssignmentList
// representation.
// CONVENTION: for each cluster assignment (i.e. element in `AssignmentList`), the nodeIDs are listed in CANONICAL ORDER
// No errors are expected during normal operation.
func convertClusterAssignments(cdcClusters []cadence.Value) (flow.AssignmentList, error) {
	// ensure we don't have duplicate cluster indices
	indices := make(map[uint]struct{})

	// parse cluster assignments to Go types
	clusterAssignments := make([]flow.IdentifierList, len(cdcClusters))
	for _, value := range cdcClusters {
		cdcCluster, ok := value.(cadence.Struct)
		if !ok {
			return nil, invalidCadenceTypeError("cluster", cdcCluster, cadence.Struct{})
		}

		if cdcCluster.Type() == nil {
			return nil, fmt.Errorf("cluster struct doesn't have type")
		}

		fields := cadence.FieldsMappedByName(cdcCluster)

		const expectedFieldCount = 2
		if len(fields) < expectedFieldCount {
			return nil, fmt.Errorf(
				"insufficient fields (%d < %d)",
				len(fields),
				expectedFieldCount,
			)
		}

		// Extract cluster fields

		clusterIndex, err := getField[cadence.UInt16](fields, "index")
		if err != nil {
			return nil, fmt.Errorf("failed to decode cluster struct: %w", err)
		}
		// ensure cluster index is valid
		if int(clusterIndex) >= len(cdcClusters) {
			return nil, fmt.Errorf(
				"invalid cdcCluster index (%d) outside range [0,%d]",
				clusterIndex,
				len(cdcClusters)-1,
			)
		}
		_, dup := indices[uint(clusterIndex)]
		if dup {
			return nil, fmt.Errorf("duplicate cdcCluster index (%d)", clusterIndex)
		}

		weightsByNodeID, err := getField[cadence.Dictionary](fields, "nodeWeights")
		if err != nil {
			return nil, fmt.Errorf("failed to decode cluster struct: %w", err)
		}

		// read weights to retrieve node IDs of cdcCluster members
		clusterMembers := make(flow.IdentifierList, 0, len(weightsByNodeID.Pairs))
		for _, pair := range weightsByNodeID.Pairs {
			nodeIDString, ok := pair.Key.(cadence.String)
			if !ok {
				return nil, invalidCadenceTypeError(
					"clusterWeights.nodeID",
					pair.Key,
					cadence.String(""),
				)
			}
			nodeID, err := flow.HexStringToIdentifier(string(nodeIDString))
			if err != nil {
				return nil, fmt.Errorf(
					"could not convert hex string to identifer: %w",
					err,
				)
			}
			clusterMembers = append(clusterMembers, nodeID)
		}

		// IMPORTANT: for each cluster, node IDs must be in *canonical order*
		clusterAssignments[clusterIndex] = clusterMembers.Sort(flow.IdentifierCanonical)
	}

	return clusterAssignments, nil
}

// convertParticipants converts the network participants specified in the
// EpochSetup event into an IdentityList.
// CONVENTION: returned IdentityList is in CANONICAL ORDER
func convertParticipants(cdcParticipants []cadence.Value) (flow.IdentitySkeletonList, error) {
	participants := make(flow.IdentitySkeletonList, 0, len(cdcParticipants))

	for _, value := range cdcParticipants {
		// checking compliance with expected format
		cdcNodeInfoStruct, ok := value.(cadence.Struct)
		if !ok {
			return nil, invalidCadenceTypeError(
				"cdcNodeInfoFields",
				value,
				cadence.Struct{},
			)
		}

		if cdcNodeInfoStruct.Type() == nil {
			return nil, fmt.Errorf("nodeInfo struct doesn't have type")
		}

		fields := cadence.FieldsMappedByName(cdcNodeInfoStruct)

		const expectedFieldCount = 14
		if len(fields) < expectedFieldCount {
			return nil, fmt.Errorf(
				"insufficient fields (%d < %d)",
				len(fields),
				expectedFieldCount,
			)
		}

		nodeIDHex, err := getField[cadence.String](fields, "id")
		if err != nil {
			return nil, fmt.Errorf("failed to decode nodeInfo struct: %w", err)
		}

		role, err := getField[cadence.UInt8](fields, "role")
		if err != nil {
			return nil, fmt.Errorf("failed to decode nodeInfo struct: %w", err)
		}
		if !flow.Role(role).Valid() {
			return nil, fmt.Errorf("invalid role %d", role)
		}

		address, err := getField[cadence.String](fields, "networkingAddress")
		if err != nil {
			return nil, fmt.Errorf("failed to decode nodeInfo struct: %w", err)
		}

		networkKeyHex, err := getField[cadence.String](fields, "networkingKey")
		if err != nil {
			return nil, fmt.Errorf("failed to decode nodeInfo struct: %w", err)
		}

		stakingKeyHex, err := getField[cadence.String](fields, "stakingKey")
		if err != nil {
			return nil, fmt.Errorf("failed to decode nodeInfo struct: %w", err)
		}

		initialWeight, err := getField[cadence.UInt64](fields, "initialWeight")
		if err != nil {
			return nil, fmt.Errorf("failed to decode nodeInfo struct: %w", err)
		}

		identity := &flow.IdentitySkeleton{
			InitialWeight: uint64(initialWeight),
			Address:       string(address),
			Role:          flow.Role(role),
		}

		// convert nodeID string into identifier
		identity.NodeID, err = flow.HexStringToIdentifier(string(nodeIDHex))
		if err != nil {
			return nil, fmt.Errorf("could not convert hex string to identifer: %w", err)
		}

		// parse to PublicKey the networking key hex string
		networkKeyBytes, err := hex.DecodeString(string(networkKeyHex))
		if err != nil {
			return nil, fmt.Errorf(
				"could not decode network public key into bytes: %w",
				err,
			)
		}
		identity.NetworkPubKey, err = crypto.DecodePublicKey(
			crypto.ECDSAP256,
			networkKeyBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("could not decode network public key: %w", err)
		}

		// parse to PublicKey the staking key hex string
		stakingKeyBytes, err := hex.DecodeString(string(stakingKeyHex))
		if err != nil {
			return nil, fmt.Errorf(
				"could not decode staking public key into bytes: %w",
				err,
			)
		}
		identity.StakingPubKey, err = crypto.DecodePublicKey(
			crypto.BLSBLS12381,
			stakingKeyBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("could not decode staking public key: %w", err)
		}

		participants = append(participants, identity)
	}

	// IMPORTANT: returned identities must be in *canonical order*
	participants = participants.Sort(flow.Canonical[flow.IdentitySkeleton])
	return participants, nil
}

// convertClusterQCVoteData converts cluster QC vote data from the EpochRecover event
// to a representation suitable for inclusion in the protocol state. Votes are
// aggregated as part of this conversion.
// TODO(efm-recovery): update this function for new QCVoteData structure (see https://github.com/onflow/flow-go/pull/5943#discussion_r1605267444)
func convertClusterQCVoteData(cdcClusterQCVoteData []cadence.Value) ([]flow.ClusterQCVoteData, error) {
	qcVoteDatas := make([]flow.ClusterQCVoteData, 0, len(cdcClusterQCVoteData))

	// CAUTION: Votes are not validated prior to aggregation. This means a single
	// invalid vote submission will result in a fully invalid QC for that cluster.
	// Votes must be validated by the ClusterQC smart contract.

	for _, cdcClusterQC := range cdcClusterQCVoteData {
		cdcClusterQCStruct, ok := cdcClusterQC.(cadence.Struct)
		if !ok {
			return nil, invalidCadenceTypeError(
				"clusterQC",
				cdcClusterQC,
				cadence.Struct{},
			)
		}

		if cdcClusterQCStruct.Type() == nil {
			return nil, fmt.Errorf("clusterQCVoteData struct doesn't have type")
		}

		fields := cadence.FieldsMappedByName(cdcClusterQCStruct)

		const expectedFieldCount = 2
		if len(fields) < expectedFieldCount {
			return nil, fmt.Errorf(
				"insufficient fields (%d < %d)",
				len(fields),
				expectedFieldCount,
			)
		}

		cdcVoterIDs, err := getField[cadence.Array](fields, "voterIDs")
		if err != nil {
			return nil, fmt.Errorf("failed to decode clusterQCVoteData struct: %w", err)
		}

		voterIDs := make([]flow.Identifier, 0, len(cdcVoterIDs.Values))
		for _, cdcVoterID := range cdcVoterIDs.Values {
			voterIDHex, ok := cdcVoterID.(cadence.String)
			if !ok {
				return nil, invalidCadenceTypeError(
					"clusterQC[i].voterID",
					cdcVoterID,
					cadence.String(""),
				)
			}
			voterID, err := flow.HexStringToIdentifier(string(voterIDHex))
			if err != nil {
				return nil, fmt.Errorf("could not convert voter ID from hex: %w", err)
			}
			voterIDs = append(voterIDs, voterID)
		}

		cdcAggSignature, err := getField[cadence.String](fields, "aggregatedSignature")
		if err != nil {
			return nil, fmt.Errorf("failed to decode clusterQCVoteData struct: %w", err)
		}

		aggregatedSignature, err := hex.DecodeString(string(cdcAggSignature))
		if err != nil {
			return nil, fmt.Errorf("could not convert raw vote from hex: %w", err)
		}

		// check that aggregated signature is not identity, because an identity signature
		// is invalid if verified under an identity public key. This can happen in two cases:
		//  - If the quorum has at least one honest signer, and given all staking key proofs of possession
		//    are valid, it's extremely unlikely for the aggregated public key (and the corresponding
		//    aggregated signature) to be identity.
		//  - If all quorum is malicious and intentionally forge an identity aggregate. As of the previous point,
		//    this is only possible if there is no honest collector involved in constructing the cluster QC.
		//    Hence, the cluster would need to contain a supermajority of malicious collectors.
		//    As we are assuming that the fraction of malicious collectors overall does not exceed 1/3  (measured
		//    by stake), the probability for randomly assigning 2/3 or more byzantine collectors to a single cluster
		//    vanishes (provided a sufficiently high collector count in total).
		//
		//  Note that at this level, all individual signatures are guaranteed to be valid
		//  w.r.t their corresponding staking public key. It is therefore enough to check
		//  the aggregated signature to conclude whether the aggregated public key is identity.
		//  This check is therefore a sanity check to catch a potential issue early.
		if crypto.IsBLSSignatureIdentity(aggregatedSignature) {
			return nil, fmt.Errorf("cluster qc vote aggregation failed because resulting BLS signature is identity")
		}

		// set the fields on the QC vote data object
		qcVoteDatas = append(qcVoteDatas, flow.ClusterQCVoteData{
			SigData:  aggregatedSignature,
			VoterIDs: voterIDs,
		})
	}

	return qcVoteDatas, nil
}

// convertClusterQCVotes converts raw cluster QC votes from the EpochCommit event
// to a representation suitable for inclusion in the protocol state. Votes are
// aggregated as part of this conversion.
func convertClusterQCVotes(cdcClusterQCs []cadence.Value) (
	[]flow.ClusterQCVoteData,
	error,
) {

	// avoid duplicate indices
	indices := make(map[uint]struct{})
	qcVoteDatas := make([]flow.ClusterQCVoteData, len(cdcClusterQCs))

	// CAUTION: Votes are not validated prior to aggregation. This means a single
	// invalid vote submission will result in a fully invalid QC for that cluster.
	// Votes must be validated by the ClusterQC smart contract.

	for _, cdcClusterQC := range cdcClusterQCs {
		cdcClusterQCStruct, ok := cdcClusterQC.(cadence.Struct)
		if !ok {
			return nil, invalidCadenceTypeError(
				"clusterQC",
				cdcClusterQC,
				cadence.Struct{},
			)
		}

		if cdcClusterQCStruct.Type() == nil {
			return nil, fmt.Errorf("clusterQC struct doesn't have type")
		}

		fields := cadence.FieldsMappedByName(cdcClusterQCStruct)

		const expectedFieldCount = 4
		if len(fields) < expectedFieldCount {
			return nil, fmt.Errorf(
				"insufficient fields (%d < %d)",
				len(fields),
				expectedFieldCount,
			)
		}

		index, err := getField[cadence.UInt16](fields, "index")
		if err != nil {
			return nil, fmt.Errorf("failed to decode clusterQC struct: %w", err)
		}

		cdcRawVotes, err := getField[cadence.Array](fields, "voteSignatures")
		if err != nil {
			return nil, fmt.Errorf("failed to decode clusterQC struct: %w", err)
		}

		cdcVoterIDs, err := getField[cadence.Array](fields, "voterIDs")
		if err != nil {
			return nil, fmt.Errorf("failed to decode clusterQC struct: %w", err)
		}

		if int(index) >= len(cdcClusterQCs) {
			return nil, fmt.Errorf(
				"invalid index (%d) not in range [0,%d]",
				index,
				len(cdcClusterQCs),
			)
		}
		_, dup := indices[uint(index)]
		if dup {
			return nil, fmt.Errorf("duplicate cluster QC index (%d)", index)
		}

		voterIDs := make([]flow.Identifier, 0, len(cdcVoterIDs.Values))
		for _, cdcVoterID := range cdcVoterIDs.Values {
			voterIDHex, ok := cdcVoterID.(cadence.String)
			if !ok {
				return nil, invalidCadenceTypeError(
					"clusterQC[i].voterID",
					cdcVoterID,
					cadence.String(""),
				)
			}
			voterID, err := flow.HexStringToIdentifier(string(voterIDHex))
			if err != nil {
				return nil, fmt.Errorf("could not convert voter ID from hex: %w", err)
			}
			voterIDs = append(voterIDs, voterID)
		}

		// gather all the vote signatures
		signatures := make([]crypto.Signature, 0, len(cdcRawVotes.Values))
		for _, cdcRawVote := range cdcRawVotes.Values {
			rawVoteHex, ok := cdcRawVote.(cadence.String)
			if !ok {
				return nil, invalidCadenceTypeError(
					"clusterQC[i].vote",
					cdcRawVote,
					cadence.String(""),
				)
			}
			rawVoteBytes, err := hex.DecodeString(string(rawVoteHex))
			if err != nil {
				return nil, fmt.Errorf("could not convert raw vote from hex: %w", err)
			}
			signatures = append(signatures, rawVoteBytes)
		}
		// Aggregate BLS signatures
		aggregatedSignature, err := crypto.AggregateBLSSignatures(signatures)
		if err != nil {
			// expected errors of the function are:
			//  - empty list of signatures
			//  - an input signature does not deserialize to a valid point
			// Both are not expected at this stage because list is guaranteed not to be
			// empty and individual signatures have been validated.
			return nil, fmt.Errorf("cluster qc vote aggregation failed: %w", err)
		}

		// check that aggregated signature is not identity, because an identity signature
		// is invalid if verified under an identity public key. This can happen in two cases:
		//  - If the quorum has at least one honest signer, and given all staking key proofs of possession
		//    are valid, it's extremely unlikely for the aggregated public key (and the corresponding
		//    aggregated signature) to be identity.
		//  - If all quorum is malicious and intentionally forge an identity aggregate. As of the previous point,
		//    this is only possible if there is no honest collector involved in constructing the cluster QC.
		//    Hence, the cluster would need to contain a supermajority of malicious collectors.
		//    As we are assuming that the fraction of malicious collectors overall does not exceed 1/3  (measured
		//    by stake), the probability for randomly assigning 2/3 or more byzantine collectors to a single cluster
		//    vanishes (provided a sufficiently high collector count in total).
		//
		//  Note that at this level, all individual signatures are guaranteed to be valid
		//  w.r.t their corresponding staking public key. It is therefore enough to check
		//  the aggregated signature to conclude whether the aggregated public key is identity.
		//  This check is therefore a sanity check to catch a potential issue early.
		if crypto.IsBLSSignatureIdentity(aggregatedSignature) {
			return nil, fmt.Errorf("cluster qc vote aggregation failed because resulting BLS signature is identity")
		}

		// set the fields on the QC vote data object
		qcVoteDatas[int(index)] = flow.ClusterQCVoteData{
			SigData:  aggregatedSignature,
			VoterIDs: voterIDs,
		}
	}

	return qcVoteDatas, nil
}

// convertDKGKeys converts hex-encoded public beacon keys as received by the DKG
// smart contract into crypto.PublicKey representations suitable for inclusion
// in the protocol state.
func convertDKGKeys(cdcDKGKeys []cadence.Value) ([]crypto.PublicKey, error) {
	convertedKeys := make([]crypto.PublicKey, 0, len(cdcDKGKeys))
	for _, value := range cdcDKGKeys {
		pubKey, err := convertDKGKey(value)
		if err != nil {
			return nil, fmt.Errorf("could not decode public beacon key share: %w", err)
		}
		convertedKeys = append(convertedKeys, pubKey)
	}
	return convertedKeys, nil
}

// convertDKGKey converts a single hex-encoded public beacon keys as received by the DKG
// smart contract into crypto.PublicKey representations suitable for inclusion
// in the protocol state.
func convertDKGKey(cdcDKGKeys cadence.Value) (crypto.PublicKey, error) {
	// extract string representation from Cadence Value
	keyHex, ok := cdcDKGKeys.(cadence.String)
	if !ok {
		return nil, invalidCadenceTypeError("dkgKey", cdcDKGKeys, cadence.String(""))
	}

	// decode individual public keys
	pubKeyBytes, err := hex.DecodeString(string(keyHex))
	if err != nil {
		return nil, fmt.Errorf("converting hex to bytes failed: %w", err)
	}
	pubKey, err := crypto.DecodePublicKey(crypto.BLSBLS12381, pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("could not decode bytes into a public key: %w", err)
	}
	return pubKey, nil
}

func invalidCadenceTypeError(
	fieldName string,
	actualType, expectedType cadence.Value,
) error {
	// NOTE: This error is reported if the Go-types are different (not if Cadence types are different).
	// Therefore, print the Go-type instead of cadence type.
	// Cadence type can be `nil`, since the `expectedType` is always the zero-value of the Go type.
	return fmt.Errorf(
		"invalid Cadence type for field %s (got=%T, expected=%T)",
		fieldName,
		actualType,
		expectedType,
	)
}

// convertServiceEventProtocolStateVersionUpgrade converts a Cadence instance of the VersionBeacon
// service event to the protocol-internal representation.
// CAUTION: This function must only be used for input events computed locally, by an
// Execution or Verification Node; it is not resilient to malicious inputs.
// No errors are expected during normal operation.
func convertServiceEventProtocolStateVersionUpgrade(event flow.Event) (*flow.ServiceEvent, error) {
	payload, err := ccf.Decode(nil, event.Payload)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event payload: %w", err)
	}

	versionUpgrade, err := DecodeCadenceValue("ProtocolStateVersionUpgrade payload", payload,
		func(cdcEvent cadence.Event) (*flow.ProtocolStateVersionUpgrade, error) {

			if cdcEvent.Type() == nil {
				return nil, fmt.Errorf("ProtocolStateVersionUpgrade event doesn't have type")
			}

			fields := cadence.FieldsMappedByName(cdcEvent)

			const expectedFieldCount = 2
			if len(fields) < expectedFieldCount {
				return nil, fmt.Errorf("unexpected number of fields in ProtocolStateVersionUpgrade (%d < %d)",
					len(fields), expectedFieldCount)
			}

			newProtocolVersionValue, err := getField[cadence.Value](fields, "newProtocolVersion")
			if err != nil {
				return nil, fmt.Errorf("failed to decode VersionBeacon event: %w", err)
			}

			activeViewValue, err := getField[cadence.Value](fields, "activeView")
			if err != nil {
				return nil, fmt.Errorf("failed to decode VersionBeacon event: %w", err)
			}

			newProtocolVersion, err := DecodeCadenceValue(
				".newProtocolVersion", newProtocolVersionValue, func(cadenceVal cadence.UInt64) (uint64, error) {
					return uint64(cadenceVal), err
				},
			)
			if err != nil {
				return nil, err
			}
			activeView, err := DecodeCadenceValue(
				".activeView", activeViewValue, func(cadenceVal cadence.UInt64) (uint64, error) {
					return uint64(cadenceVal), err
				},
			)
			if err != nil {
				return nil, err
			}

			return &flow.ProtocolStateVersionUpgrade{
				NewProtocolStateVersion: newProtocolVersion,
				ActiveView:              activeView,
			}, nil
		})
	if err != nil {
		return nil, fmt.Errorf("could not decode cadence value: %w", err)
	}

	// create the service event
	serviceEvent := &flow.ServiceEvent{
		Type:  flow.ServiceEventProtocolStateVersionUpgrade,
		Event: versionUpgrade,
	}
	return serviceEvent, nil
}

// convertServiceEventVersionBeacon converts a Cadence instance of the VersionBeacon
// service event to the protocol-internal representation.
// CAUTION: This function must only be used for input events computed locally, by an
// Execution or Verification Node; it is not resilient to malicious inputs.
// No errors are expected during normal operation.
func convertServiceEventVersionBeacon(event flow.Event) (*flow.ServiceEvent, error) {
	payload, err := ccf.Decode(nil, event.Payload)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event payload: %w", err)
	}

	versionBeacon, err := DecodeCadenceValue(
		"VersionBeacon payload", payload, func(cdcEvent cadence.Event) (*flow.VersionBeacon, error) {

			if cdcEvent.Type() == nil {
				return nil, fmt.Errorf("VersionBeacon event doesn't have type")
			}

			fields := cadence.FieldsMappedByName(cdcEvent)

			const expectedFieldCount = 2
			if len(fields) != expectedFieldCount {
				return nil, fmt.Errorf(
					"unexpected number of fields in VersionBeacon event (%d != %d)",
					len(fields),
					expectedFieldCount,
				)
			}

			versionBoundariesValue, err := getField[cadence.Value](fields, "versionBoundaries")
			if err != nil {
				return nil, fmt.Errorf("failed to decode VersionBeacon event: %w", err)
			}

			sequenceValue, err := getField[cadence.Value](fields, "sequence")
			if err != nil {
				return nil, fmt.Errorf("failed to decode VersionBeacon event: %w", err)
			}

			versionBoundaries, err := DecodeCadenceValue(
				".versionBoundaries", versionBoundariesValue, convertVersionBoundaries,
			)
			if err != nil {
				return nil, err
			}

			sequence, err := DecodeCadenceValue(
				".sequence", sequenceValue, func(cadenceVal cadence.UInt64) (
					uint64,
					error,
				) {
					return uint64(cadenceVal), nil
				},
			)
			if err != nil {
				return nil, err
			}

			return &flow.VersionBeacon{
				VersionBoundaries: versionBoundaries,
				Sequence:          sequence,
			}, err
		},
	)
	if err != nil {
		return nil, err
	}

	// a converted version beacon event should also be valid
	if err := versionBeacon.Validate(); err != nil {
		return nil, fmt.Errorf("invalid VersionBeacon event: %w", err)
	}

	// create the service event
	serviceEvent := &flow.ServiceEvent{
		Type:  flow.ServiceEventVersionBeacon,
		Event: versionBeacon,
	}

	return serviceEvent, nil
}

func convertVersionBoundaries(array cadence.Array) (
	[]flow.VersionBoundary,
	error,
) {
	boundaries := make([]flow.VersionBoundary, len(array.Values))

	for i, cadenceVal := range array.Values {
		boundary, err := VersionBoundary(cadenceVal)
		if err != nil {
			return nil, decodeError{
				location: fmt.Sprintf(".Values[%d]", i),
				err:      err,
			}
		}
		boundaries[i] = boundary
	}

	return boundaries, nil
}

// VersionBoundary decodes a single version boundary from the given Cadence value.
func VersionBoundary(value cadence.Value) (
	flow.VersionBoundary,
	error,
) {
	boundary, err := DecodeCadenceValue(
		"VersionBoundary",
		value,
		func(structVal cadence.Struct) (
			flow.VersionBoundary,
			error,
		) {
			if structVal.Type() == nil {
				return flow.VersionBoundary{}, fmt.Errorf("VersionBoundary struct doesn't have type")
			}

			fields := cadence.FieldsMappedByName(structVal)

			const expectedFieldCount = 2
			if len(fields) < expectedFieldCount {
				return flow.VersionBoundary{}, fmt.Errorf(
					"incorrect number of fields (%d != %d)",
					len(fields),
					expectedFieldCount,
				)
			}

			blockHeightValue, err := getField[cadence.Value](fields, "blockHeight")
			if err != nil {
				return flow.VersionBoundary{}, fmt.Errorf("failed to decode VersionBoundary struct: %w", err)
			}

			versionValue, err := getField[cadence.Value](fields, "version")
			if err != nil {
				return flow.VersionBoundary{}, fmt.Errorf("failed to decode VersionBoundary struct: %w", err)
			}

			height, err := DecodeCadenceValue(
				".blockHeight",
				blockHeightValue,
				func(cadenceVal cadence.UInt64) (
					uint64,
					error,
				) {
					return uint64(cadenceVal), nil
				},
			)
			if err != nil {
				return flow.VersionBoundary{}, err
			}

			version, err := DecodeCadenceValue(
				".version",
				versionValue,
				convertSemverVersion,
			)
			if err != nil {
				return flow.VersionBoundary{}, err
			}

			return flow.VersionBoundary{
				BlockHeight: height,
				Version:     version,
			}, nil
		},
	)
	return boundary, err
}

func convertSemverVersion(structVal cadence.Struct) (
	string,
	error,
) {
	if structVal.Type() == nil {
		return "", fmt.Errorf("Semver struct doesn't have type")
	}

	fields := cadence.FieldsMappedByName(structVal)

	const expectedFieldCount = 4
	if len(fields) < expectedFieldCount {
		return "", fmt.Errorf(
			"incorrect number of fields (%d != %d)",
			len(fields),
			expectedFieldCount,
		)
	}

	majorValue, err := getField[cadence.Value](fields, "major")
	if err != nil {
		return "", fmt.Errorf("failed to decode SemVer struct: %w", err)
	}

	minorValue, err := getField[cadence.Value](fields, "minor")
	if err != nil {
		return "", fmt.Errorf("failed to decode SemVer struct: %w", err)
	}

	patchValue, err := getField[cadence.Value](fields, "patch")
	if err != nil {
		return "", fmt.Errorf("failed to decode SemVer struct: %w", err)
	}

	preReleaseValue, err := getField[cadence.Value](fields, "preRelease")
	if err != nil {
		return "", fmt.Errorf("failed to decode SemVer struct: %w", err)
	}

	major, err := DecodeCadenceValue(
		".major",
		majorValue,
		func(cadenceVal cadence.UInt8) (
			uint64,
			error,
		) {
			return uint64(cadenceVal), nil
		},
	)
	if err != nil {
		return "", err
	}

	minor, err := DecodeCadenceValue(
		".minor",
		minorValue,
		func(cadenceVal cadence.UInt8) (
			uint64,
			error,
		) {
			return uint64(cadenceVal), nil
		},
	)
	if err != nil {
		return "", err
	}

	patch, err := DecodeCadenceValue(
		".patch",
		patchValue,
		func(cadenceVal cadence.UInt8) (
			uint64,
			error,
		) {
			return uint64(cadenceVal), nil
		},
	)
	if err != nil {
		return "", err
	}

	preRelease, err := DecodeCadenceValue(
		".preRelease",
		preReleaseValue,
		func(cadenceVal cadence.Optional) (
			string,
			error,
		) {
			if cadenceVal.Value == nil {
				return "", nil
			}

			return DecodeCadenceValue(
				"!",
				cadenceVal.Value,
				func(cadenceVal cadence.String) (
					string,
					error,
				) {
					return string(cadenceVal), nil
				},
			)
		},
	)
	if err != nil {
		return "", err
	}

	version := semver.Version{
		Major:      int64(major),
		Minor:      int64(minor),
		Patch:      int64(patch),
		PreRelease: semver.PreRelease(preRelease),
	}

	return version.String(), nil

}

type decodeError struct {
	location string
	err      error
}

func (e decodeError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("decoding error %s: %s", e.location, e.err.Error())
	}
	return fmt.Sprintf("decoding error %s", e.location)
}

func (e decodeError) Unwrap() error {
	return e.err
}

func DecodeCadenceValue[From cadence.Value, Into any](
	location string,
	value cadence.Value,
	decodeInner func(From) (Into, error),
) (Into, error) {
	var defaultInto Into
	if value == nil {
		return defaultInto, decodeError{
			location: location,
			err:      nil,
		}
	}

	convertedValue, is := value.(From)
	if !is {
		return defaultInto, decodeError{
			location: location,
			err: fmt.Errorf(
				"invalid Cadence type (got=%T, expected=%T)",
				value,
				*new(From),
			),
		}
	}

	inner, err := decodeInner(convertedValue)
	if err != nil {
		if err, is := err.(decodeError); is {
			return defaultInto, decodeError{
				location: location + err.location,
				err:      err.err,
			}
		}
		return defaultInto, decodeError{
			location: location,
			err:      err,
		}
	}

	return inner, nil
}
