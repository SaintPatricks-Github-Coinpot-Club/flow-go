package flow

import (
	"fmt"

	"github.com/onflow/crypto"
)

// Attestation confirms correctness of a chunk of an exec result
//
//structwrite:immutable - mutations allowed only within the constructor
type Attestation struct {
	BlockID           Identifier // ID of the block included the collection
	ExecutionResultID Identifier // ID of the execution result
	ChunkIndex        uint64     // index of the approved chunk
}

// UntrustedAttestation is an untrusted input-only representation of an Attestation,
// used for construction.
//
// This type exists to ensure that constructor functions are invoked explicitly
// with named fields, which improves clarity and reduces the risk of incorrect field
// ordering during construction.
//
// An instance of UntrustedAttestation should be validated and converted into
// a trusted Attestation using NewAttestation constructor.
type UntrustedAttestation Attestation

// NewAttestation creates a new instance of Attestation.
// Construction Attestation allowed only within the constructor.
//
// All errors indicate a valid Attestation cannot be constructed from the input.
// ChunkIndex can be zero in principle, so we don’t check it.
func NewAttestation(untrusted UntrustedAttestation) (*Attestation, error) {
	if untrusted.BlockID == ZeroID {
		return nil, fmt.Errorf("BlockID must not be empty")
	}

	if untrusted.ExecutionResultID == ZeroID {
		return nil, fmt.Errorf("ExecutionResultID must not be empty")
	}

	return &Attestation{
		BlockID:           untrusted.BlockID,
		ExecutionResultID: untrusted.ExecutionResultID,
		ChunkIndex:        untrusted.ChunkIndex,
	}, nil
}

// ID generates a unique identifier using attestation
func (a Attestation) ID() Identifier {
	return MakeID(a)
}

// ResultApprovalBody holds body part of a result approval
//
//structwrite:immutable - mutations allowed only within the constructor
type ResultApprovalBody struct {
	Attestation
	ApproverID           Identifier       // node id generating this result approval
	AttestationSignature crypto.Signature // signature over attestation, this has been separated for BLS aggregation
	Spock                crypto.Signature // proof of re-computation, one per each chunk
}

// UntrustedResultApprovalBody is an untrusted input-only representation of an ResultApprovalBody,
// used for construction.
//
// This type exists to ensure that constructor functions are invoked explicitly
// with named fields, which improves clarity and reduces the risk of incorrect field
// ordering during construction.
//
// An instance of UntrustedResultApprovalBody should be validated and converted into
// a trusted ResultApprovalBody using NewResultApprovalBody constructor.
type UntrustedResultApprovalBody ResultApprovalBody

// NewResultApprovalBody creates a new instance of ResultApprovalBody.
// Construction ResultApprovalBody allowed only within the constructor.
//
// All errors indicate a valid Collection cannot be constructed from the input.
func NewResultApprovalBody(untrusted UntrustedResultApprovalBody) (*ResultApprovalBody, error) {
	att, err := NewAttestation(UntrustedAttestation(untrusted.Attestation))
	if err != nil {
		return nil, fmt.Errorf("invalid attestation: %w", err)
	}

	if untrusted.ApproverID == ZeroID {
		return nil, fmt.Errorf("ApproverID must not be empty")
	}

	if len(untrusted.AttestationSignature) == 0 {
		return nil, fmt.Errorf("AttestationSignature must not be empty")
	}

	if len(untrusted.Spock) == 0 {
		return nil, fmt.Errorf("Spock proof must not be empty")
	}

	return &ResultApprovalBody{
		Attestation:          *att,
		ApproverID:           untrusted.ApproverID,
		AttestationSignature: untrusted.AttestationSignature,
		Spock:                untrusted.Spock,
	}, nil
}

// PartialID generates a unique identifier using Attestation + ApproverID
func (rab ResultApprovalBody) PartialID() Identifier {
	data := struct {
		Attestation Attestation
		ApproverID  Identifier
	}{
		Attestation: rab.Attestation,
		ApproverID:  rab.ApproverID,
	}

	return MakeID(data)
}

// ID generates a unique identifier using ResultApprovalBody
func (rab ResultApprovalBody) ID() Identifier {
	return MakeID(rab)
}

// ResultApproval includes an approval for a chunk, verified by a verification node
//
//structwrite:immutable - mutations allowed only within the constructor
type ResultApproval struct {
	Body              ResultApprovalBody
	VerifierSignature crypto.Signature // signature over all above fields
}

// UntrustedResultApproval is an untrusted input-only representation of an ResultApproval,
// used for construction.
//
// This type exists to ensure that constructor functions are invoked explicitly
// with named fields, which improves clarity and reduces the risk of incorrect field
// ordering during construction.
//
// An instance of UntrustedResultApproval should be validated and converted into
// a trusted ResultApproval using NewResultApproval constructor.
type UntrustedResultApproval ResultApproval

// NewResultApproval creates a new instance of ResultApproval.
// Construction ResultApproval allowed only within the constructor.
//
// All errors indicate a valid Collection cannot be constructed from the input.
func NewResultApproval(untrusted UntrustedResultApproval) (*ResultApproval, error) {
	rab, err := NewResultApprovalBody(UntrustedResultApprovalBody(untrusted.Body))
	if err != nil {
		return nil, fmt.Errorf("invalid result approval body: %w", err)
	}

	if len(untrusted.VerifierSignature) == 0 {
		return nil, fmt.Errorf("VerifierSignature must not be empty")
	}

	return &ResultApproval{
		Body:              *rab,
		VerifierSignature: untrusted.VerifierSignature,
	}, nil
}

// ID generates a unique identifier using result approval body
func (ra ResultApproval) ID() Identifier {
	return MakeID(ra.Body)
}

// Checksum generates checksum using the result approval full content
func (ra ResultApproval) Checksum() Identifier {
	return MakeID(ra)
}
