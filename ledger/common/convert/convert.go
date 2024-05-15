package convert

import (
	"fmt"

	"github.com/onflow/cadence/runtime/common"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"
)

const (
	KeyPartOwner = uint16(0)
	// Deprecated: KeyPartController was only used by the very first
	// version of Cadence for access control, which was later retired
	_          = uint16(1) // DO NOT REUSE
	KeyPartKey = uint16(2)
)

// UnexpectedLedgerKeyFormat is returned when a ledger key is not in the expected format
var UnexpectedLedgerKeyFormat = fmt.Errorf("unexpected ledger key format")

// AddressToRegisterOwner converts 8-byte address to register owner.
// If given address is ZeroAddress, register owner is "" (global register).
func AddressToRegisterOwner(address common.Address) string {
	// Global registers have address zero and an empty owner field
	if address == common.ZeroAddress {
		return ""
	}

	// All other registers have the account's address
	return string(address.Bytes())
}

// LedgerKeyToRegisterID converts a ledger key to a register id
// returns an UnexpectedLedgerKeyFormat error if the key is not in the expected format
func LedgerKeyToRegisterID(key ledger.Key) (flow.RegisterID, error) {
	parts := key.KeyParts
	if len(parts) != 2 ||
		parts[0].Type != KeyPartOwner ||
		parts[1].Type != KeyPartKey {
		return flow.RegisterID{}, fmt.Errorf("ledger key %s: %w", key.String(), UnexpectedLedgerKeyFormat)
	}

	return flow.NewRegisterID(
		flow.BytesToAddress(parts[0].Value),
		string(parts[1].Value),
	), nil
}

// RegisterIDToLedgerKey converts a register id to a ledger key
func RegisterIDToLedgerKey(registerID flow.RegisterID) ledger.Key {
	return ledger.Key{
		KeyParts: []ledger.KeyPart{
			{
				Type:  KeyPartOwner,
				Value: []byte(registerID.Owner),
			},
			{
				Type:  KeyPartKey,
				Value: []byte(registerID.Key),
			},
		},
	}
}

// PayloadToRegister converts a payload to a register id and value
func PayloadToRegister(payload *ledger.Payload) (flow.RegisterID, flow.RegisterValue, error) {
	key, err := payload.Key()
	if err != nil {
		return flow.RegisterID{}, flow.RegisterValue{}, fmt.Errorf("could not parse register key from payload: %w", err)
	}
	regID, err := LedgerKeyToRegisterID(key)
	if err != nil {
		return flow.RegisterID{}, flow.RegisterValue{}, fmt.Errorf("could not convert register key into register id: %w", err)
	}

	return regID, payload.Value(), nil
}
