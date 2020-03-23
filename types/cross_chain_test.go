package types_test

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitCrossChainID(t *testing.T) {
	sourceChainID := types.CrossChainID(0x0001)
	types.SetSourceChainID(sourceChainID)

	require.Equal(t, sourceChainID, types.GetSourceChainID())
}

func TestRegisterNewCrossChainChannel(t *testing.T) {
	require.NoError(t, types.RegisterNewCrossChainChannel("bind", types.ChannelID(1)))
	require.NoError(t, types.RegisterNewCrossChainChannel("transfer", types.ChannelID(2)))
	require.NoError(t, types.RegisterNewCrossChainChannel("timeout", types.ChannelID(3)))
	require.NoError(t, types.RegisterNewCrossChainChannel("staking", types.ChannelID(4)))

	channeID, err := types.GetChannelID("transfer")
	require.NoError(t, err)
	require.Equal(t, types.ChannelID(2), channeID)

	channeID, err = types.GetChannelID("staking")
	require.NoError(t, err)
	require.Equal(t, types.ChannelID(4), channeID)
}

func TestCrossChainID(t *testing.T) {
	chainID, err := types.ParseCrossChainID("123")
	require.NoError(t, err)
	require.Equal(t, types.CrossChainID(123), chainID)

	_, err = types.ParseCrossChainID("65537")
	require.Error(t, err)
}