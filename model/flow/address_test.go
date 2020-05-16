package flow_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/utils/unittest"
)

type addressWrapper struct {
	Address flow.Address
}

func TestAddressJSON(t *testing.T) {
	addr := unittest.AddressFixture()
	data, err := json.Marshal(addressWrapper{Address: addr})
	require.Nil(t, err)

	t.Log(string(data))

	var out addressWrapper
	err = json.Unmarshal(data, &out)
	require.Nil(t, err)
	assert.Equal(t, addr, out.Address)
}

func TestShort(t *testing.T) {
	type testcase struct {
		addr     flow.Address
		expected string
	}

	cases := []testcase{
		{
			addr:     flow.RootAddress,
			expected: "01",
		},
		{
			addr:     flow.HexToAddress("0000000002"),
			expected: "02",
		},
		{
			addr:     flow.HexToAddress("1f10"),
			expected: "1f10",
		},
		{
			addr:     flow.HexToAddress("0f10"),
			expected: "0f10",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.addr.Short(), c.expected)
	}
}

func TestConstants(t *testing.T) {
	//check the Zero and Root constants
	var expected [flow.AddressLength]byte
	assert.Equal(t, flow.ZeroAddress, flow.Address(expected))
	expected[flow.AddressLength-1] = 1
	assert.Equal(t, flow.RootAddress, flow.Address(expected))

	// check the transition from account zero to root
	state := flow.AddressState(0)
	address, _, err := flow.AccountAddress(state)
	require.NoError(t, err)
	assert.Equal(t, address, flow.RootAddress)
	// check high states
	state = flow.AddressState(1<<45 - 1)
	_, state, err = flow.AccountAddress(state)
	assert.NoError(t, err)
	_, _, err = flow.AccountAddress(state)
	assert.Error(t, err)
}


