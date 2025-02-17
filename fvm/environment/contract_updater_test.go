package environment

import (
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/fvm/programs"

	"github.com/onflow/flow-go/fvm/blueprints"
	stateMock "github.com/onflow/flow-go/fvm/mock/state"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/fvm/utils"
	"github.com/onflow/flow-go/model/flow"
)

type testContractUpdaterStubs struct {
	deploymentEnabled    bool
	removalEnabled       bool
	deploymentAuthorized []common.Address
	removalAuthorized    []common.Address

	auditFunc func(address runtime.Address, code []byte) (bool, error)
}

func (p testContractUpdaterStubs) RestrictedDeploymentEnabled() bool {
	return p.deploymentEnabled
}

func (p testContractUpdaterStubs) RestrictedRemovalEnabled() bool {
	return p.removalEnabled
}

func (p testContractUpdaterStubs) GetAuthorizedAccounts(
	path cadence.Path,
) []common.Address {
	if path == blueprints.ContractDeploymentAuthorizedAddressesPath {
		return p.deploymentAuthorized
	}
	return p.removalAuthorized
}

func (p testContractUpdaterStubs) UseContractAuditVoucher(
	address runtime.Address,
	code []byte,
) (
	bool,
	error,
) {
	return p.auditFunc(address, code)
}

func TestContract_ChildMergeFunctionality(t *testing.T) {
	txnState := state.NewTransactionState(
		utils.NewSimpleView(),
		state.DefaultParameters())
	accounts := NewAccounts(txnState)
	address := flow.HexToAddress("01")
	rAdd := runtime.Address(address)
	err := accounts.Create(nil, address)
	require.NoError(t, err)

	contractUpdater := NewContractUpdaterForTesting(
		accounts,
		testContractUpdaterStubs{})

	// no contract initially
	names, err := accounts.GetContractNames(address)
	require.NoError(t, err)
	require.Equal(t, len(names), 0)

	// set contract no need for signing accounts
	err = contractUpdater.setContract(rAdd, "testContract", []byte("ABC"), nil)
	require.NoError(t, err)
	require.True(t, contractUpdater.hasUpdates())

	// should not be readable from draft
	cont, err := accounts.GetContract("testContract", address)
	require.NoError(t, err)
	require.Equal(t, len(cont), 0)

	// commit
	_, err = contractUpdater.Commit()
	require.NoError(t, err)
	cont, err = accounts.GetContract("testContract", address)
	require.NoError(t, err)
	require.Equal(t, cont, []byte("ABC"))

	// rollback
	err = contractUpdater.setContract(rAdd, "testContract2", []byte("ABC"), nil)
	require.NoError(t, err)
	contractUpdater.Reset()
	require.False(t, contractUpdater.hasUpdates())
	_, err = contractUpdater.Commit()
	require.NoError(t, err)

	// test contract shouldn't be there
	cont, err = accounts.GetContract("testContract2", address)
	require.NoError(t, err)
	require.Equal(t, len(cont), 0)

	// test contract should be there
	cont, err = accounts.GetContract("testContract", address)
	require.NoError(t, err)
	require.Equal(t, cont, []byte("ABC"))

	// remove
	err = contractUpdater.removeContract(rAdd, "testContract", nil)
	require.NoError(t, err)

	// contract still there because no commit yet
	cont, err = accounts.GetContract("testContract", address)
	require.NoError(t, err)
	require.Equal(t, cont, []byte("ABC"))

	// commit removal
	_, err = contractUpdater.Commit()
	require.NoError(t, err)

	// contract should no longer be there
	cont, err = accounts.GetContract("testContract", address)
	require.NoError(t, err)
	require.Equal(t, []byte(nil), cont)
}

func TestContract_AuthorizationFunctionality(t *testing.T) {
	txnState := state.NewTransactionState(
		utils.NewSimpleView(),
		state.DefaultParameters())
	accounts := NewAccounts(txnState)

	authAdd := flow.HexToAddress("01")
	rAdd := runtime.Address(authAdd)
	err := accounts.Create(nil, authAdd)
	require.NoError(t, err)

	authRemove := flow.HexToAddress("02")
	rRemove := runtime.Address(authRemove)
	err = accounts.Create(nil, authRemove)
	require.NoError(t, err)

	authBoth := flow.HexToAddress("03")
	rBoth := runtime.Address(authBoth)
	err = accounts.Create(nil, authBoth)
	require.NoError(t, err)

	unAuth := flow.HexToAddress("04")
	unAuthR := runtime.Address(unAuth)
	err = accounts.Create(nil, unAuth)
	require.NoError(t, err)

	makeUpdater := func() *contractUpdater {
		return NewContractUpdaterForTesting(
			accounts,
			testContractUpdaterStubs{
				deploymentEnabled:    true,
				removalEnabled:       true,
				deploymentAuthorized: []common.Address{rAdd, rBoth},
				removalAuthorized:    []common.Address{rRemove, rBoth},
				auditFunc:            func(address runtime.Address, code []byte) (bool, error) { return false, nil },
			})

	}

	t.Run("try to set contract with unauthorized account", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract1", []byte("ABC"), []common.Address{unAuthR})
		require.Error(t, err)
		require.False(t, contractUpdater.hasUpdates())
	})

	t.Run("try to set contract with account only authorized for removal", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract1", []byte("ABC"), []common.Address{rRemove})
		require.Error(t, err)
		require.False(t, contractUpdater.hasUpdates())
	})

	t.Run("set contract with account authorized for adding", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract2", []byte("ABC"), []common.Address{rAdd})
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())
	})

	t.Run("set contract with account authorized for adding and removing", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract2", []byte("ABC"), []common.Address{rBoth})
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())
	})

	t.Run("try to remove contract with unauthorized account", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract1", []byte("ABC"), []common.Address{rAdd})
		require.NoError(t, err)
		_, err = contractUpdater.Commit()
		require.NoError(t, err)

		err = contractUpdater.removeContract(unAuthR, "testContract2", []common.Address{unAuthR})
		require.Error(t, err)
		require.False(t, contractUpdater.hasUpdates())
	})

	t.Run("remove contract account authorized for removal", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract1", []byte("ABC"), []common.Address{rAdd})
		require.NoError(t, err)
		_, err = contractUpdater.Commit()
		require.NoError(t, err)

		err = contractUpdater.removeContract(rRemove, "testContract2", []common.Address{rRemove})
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())
	})

	t.Run("try to remove contract with account only authorized for adding", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract1", []byte("ABC"), []common.Address{rAdd})
		require.NoError(t, err)
		_, err = contractUpdater.Commit()
		require.NoError(t, err)

		err = contractUpdater.removeContract(rAdd, "testContract2", []common.Address{rAdd})
		require.Error(t, err)
		require.False(t, contractUpdater.hasUpdates())
	})

	t.Run("remove contract with account authorized for adding and removing", func(t *testing.T) {
		contractUpdater := makeUpdater()

		err = contractUpdater.setContract(rAdd, "testContract1", []byte("ABC"), []common.Address{rAdd})
		require.NoError(t, err)
		_, err = contractUpdater.Commit()
		require.NoError(t, err)

		err = contractUpdater.removeContract(rBoth, "testContract2", []common.Address{rBoth})
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())
	})
}

func TestContract_DeploymentVouchers(t *testing.T) {

	txnState := state.NewTransactionState(
		utils.NewSimpleView(),
		state.DefaultParameters())
	accounts := NewAccounts(txnState)

	addressWithVoucher := flow.HexToAddress("01")
	addressWithVoucherRuntime := runtime.Address(addressWithVoucher)
	err := accounts.Create(nil, addressWithVoucher)
	require.NoError(t, err)

	addressNoVoucher := flow.HexToAddress("02")
	addressNoVoucherRuntime := runtime.Address(addressNoVoucher)
	err = accounts.Create(nil, addressNoVoucher)
	require.NoError(t, err)

	contractUpdater := NewContractUpdaterForTesting(
		accounts,
		testContractUpdaterStubs{
			deploymentEnabled: true,
			removalEnabled:    true,
			auditFunc: func(address runtime.Address, code []byte) (bool, error) {
				if address.String() == addressWithVoucher.String() {
					return true, nil
				}
				return false, nil
			},
		})

	// set contract without voucher
	err = contractUpdater.setContract(
		addressNoVoucherRuntime,
		"TestContract1",
		[]byte("pub contract TestContract1 {}"),
		[]common.Address{
			addressNoVoucherRuntime,
		},
	)
	require.Error(t, err)
	require.False(t, contractUpdater.hasUpdates())

	// try to set contract with voucher
	err = contractUpdater.setContract(
		addressWithVoucherRuntime,
		"TestContract2",
		[]byte("pub contract TestContract2 {}"),
		[]common.Address{
			addressWithVoucherRuntime,
		},
	)
	require.NoError(t, err)
	require.True(t, contractUpdater.hasUpdates())
}

func TestContract_ContractUpdate(t *testing.T) {

	txnState := state.NewTransactionState(
		utils.NewSimpleView(),
		state.DefaultParameters())
	accounts := NewAccounts(txnState)

	flowAddress := flow.HexToAddress("01")
	runtimeAddress := runtime.Address(flowAddress)
	err := accounts.Create(nil, flowAddress)
	require.NoError(t, err)

	var authorizationChecked bool

	contractUpdater := NewContractUpdaterForTesting(
		accounts,
		testContractUpdaterStubs{
			deploymentEnabled: true,
			removalEnabled:    true,
			auditFunc: func(address runtime.Address, code []byte) (bool, error) {
				// Ensure the voucher check is only called once,
				// for the initial contract deployment,
				// and not for the subsequent update
				require.False(t, authorizationChecked)
				authorizationChecked = true
				return true, nil
			},
		})

	// deploy contract with voucher
	err = contractUpdater.setContract(
		runtimeAddress,
		"TestContract",
		[]byte("pub contract TestContract {}"),
		[]common.Address{
			runtimeAddress,
		},
	)
	require.NoError(t, err)
	require.True(t, contractUpdater.hasUpdates())

	contractUpdateKeys, err := contractUpdater.Commit()
	require.NoError(t, err)
	require.Equal(
		t,
		[]programs.ContractUpdateKey{
			{
				Address: flowAddress,
				Name:    "TestContract",
			},
		},
		contractUpdateKeys,
	)

	// try to update contract without voucher
	err = contractUpdater.setContract(
		runtimeAddress,
		"TestContract",
		[]byte("pub contract TestContract {}"),
		[]common.Address{
			runtimeAddress,
		},
	)
	require.NoError(t, err)
	require.True(t, contractUpdater.hasUpdates())
}

func TestContract_DeterministicErrorOnCommit(t *testing.T) {
	mockAccounts := &stateMock.Accounts{}

	mockAccounts.On("ContractExists", mock.Anything, mock.Anything).
		Return(false, nil)

	mockAccounts.On("SetContract", mock.Anything, mock.Anything, mock.Anything).
		Return(func(contractName string, address flow.Address, contract []byte) error {
			return fmt.Errorf("%s %s", contractName, address.Hex())
		})

	contractUpdater := NewContractUpdaterForTesting(
		mockAccounts,
		testContractUpdaterStubs{})

	address1 := runtime.Address(flow.HexToAddress("0000000000000001"))
	address2 := runtime.Address(flow.HexToAddress("0000000000000002"))

	err := contractUpdater.setContract(address2, "A", []byte("ABC"), nil)
	require.NoError(t, err)

	err = contractUpdater.setContract(address1, "B", []byte("ABC"), nil)
	require.NoError(t, err)

	err = contractUpdater.setContract(address1, "A", []byte("ABC"), nil)
	require.NoError(t, err)

	_, err = contractUpdater.Commit()
	require.EqualError(t, err, "A 0000000000000001")
}

func TestContract_ContractRemoval(t *testing.T) {

	txnState := state.NewTransactionState(
		utils.NewSimpleView(),
		state.DefaultParameters())
	accounts := NewAccounts(txnState)

	flowAddress := flow.HexToAddress("01")
	runtimeAddress := runtime.Address(flowAddress)
	err := accounts.Create(nil, flowAddress)
	require.NoError(t, err)

	t.Run("contract removal with restriction", func(t *testing.T) {
		var authorizationChecked bool

		contractUpdater := NewContractUpdaterForTesting(
			accounts,
			testContractUpdaterStubs{
				removalEnabled: true,
				auditFunc: func(address runtime.Address, code []byte) (bool, error) {
					// Ensure the voucher check is only called once,
					// for the initial contract deployment,
					// and not for the subsequent update
					require.False(t, authorizationChecked)
					authorizationChecked = true
					return true, nil
				},
			})

		// deploy contract with voucher
		err = contractUpdater.setContract(
			runtimeAddress,
			"TestContract",
			[]byte("pub contract TestContract {}"),
			[]common.Address{
				runtimeAddress,
			},
		)
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())

		contractUpdateKeys, err := contractUpdater.Commit()
		require.NoError(t, err)
		require.Equal(
			t,
			[]programs.ContractUpdateKey{
				{
					Address: flowAddress,
					Name:    "TestContract",
				},
			},
			contractUpdateKeys,
		)

		// update should work
		err = contractUpdater.setContract(
			runtimeAddress,
			"TestContract",
			[]byte("pub contract TestContract {}"),
			[]common.Address{
				runtimeAddress,
			},
		)
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())

		// try remove contract should fail
		err = contractUpdater.removeContract(
			runtimeAddress,
			"TestContract",
			[]common.Address{
				runtimeAddress,
			},
		)
		require.Error(t, err)
	})

	t.Run("contract removal without restriction", func(t *testing.T) {
		var authorizationChecked bool

		contractUpdater := NewContractUpdaterForTesting(
			accounts,
			testContractUpdaterStubs{
				auditFunc: func(address runtime.Address, code []byte) (bool, error) {
					// Ensure the voucher check is only called once,
					// for the initial contract deployment,
					// and not for the subsequent update
					require.False(t, authorizationChecked)
					authorizationChecked = true
					return true, nil
				},
			})

		// deploy contract with voucher
		err = contractUpdater.setContract(
			runtimeAddress,
			"TestContract",
			[]byte("pub contract TestContract {}"),
			[]common.Address{
				runtimeAddress,
			},
		)
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())

		contractUpdateKeys, err := contractUpdater.Commit()
		require.NoError(t, err)
		require.Equal(
			t,
			[]programs.ContractUpdateKey{
				{
					Address: flowAddress,
					Name:    "TestContract",
				},
			},
			contractUpdateKeys,
		)

		// update should work
		err = contractUpdater.setContract(
			runtimeAddress,
			"TestContract",
			[]byte("pub contract TestContract {}"),
			[]common.Address{
				runtimeAddress,
			},
		)
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())

		// try remove contract should fail
		err = contractUpdater.removeContract(
			runtimeAddress,
			"TestContract",
			[]common.Address{
				runtimeAddress,
			},
		)
		require.NoError(t, err)
		require.True(t, contractUpdater.hasUpdates())

	})
}
