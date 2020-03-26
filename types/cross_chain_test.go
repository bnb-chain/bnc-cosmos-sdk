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

func TestRegisterCrossChainChannel(t *testing.T) {
	require.NoError(t, types.RegisterCrossChainChannel("bind", types.ChannelID(1)))
	require.NoError(t, types.RegisterCrossChainChannel("transfer", types.ChannelID(2)))
	require.NoError(t, types.RegisterCrossChainChannel("timeout", types.ChannelID(3)))
	require.NoError(t, types.RegisterCrossChainChannel("staking", types.ChannelID(4)))
	require.Error(t, types.RegisterCrossChainChannel("staking", types.ChannelID(5)))
	require.Error(t, types.RegisterCrossChainChannel("staking-new", types.ChannelID(4)))

	channeID, err := types.GetChannelID("transfer")
	require.NoError(t, err)
	require.Equal(t, types.ChannelID(2), channeID)

	channeID, err = types.GetChannelID("staking")
	require.NoError(t, err)
	require.Equal(t, types.ChannelID(4), channeID)
}

func TestRegisterDestChainID(t *testing.T) {
	require.NoError(t, types.RegisterDestChainID("bsc", types.CrossChainID(1)))
	require.NoError(t, types.RegisterDestChainID("ethereum", types.CrossChainID(2)))
	require.NoError(t, types.RegisterDestChainID("btc", types.CrossChainID(3)))
	require.NoError(t, types.RegisterDestChainID("cosmos", types.CrossChainID(4)))
	require.Error(t, types.RegisterDestChainID("cosmos", types.CrossChainID(5)))
	require.Error(t, types.RegisterDestChainID("mock", types.CrossChainID(4)))

	destChainID, err := types.GetDestChainID("bsc")
	require.NoError(t, err)
	require.Equal(t, types.CrossChainID(1), destChainID)

	destChainID, err = types.GetDestChainID("btc")
	require.NoError(t, err)
	require.Equal(t, types.CrossChainID(3), destChainID)
}

func TestCrossChainID(t *testing.T) {
	chainID, err := types.ParseCrossChainID("123")
	require.NoError(t, err)
	require.Equal(t, types.CrossChainID(123), chainID)

	_, err = types.ParseCrossChainID("65537")
	require.Error(t, err)
}