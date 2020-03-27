package ibc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func TestInitCrossChainID(t *testing.T) {
	sourceChainID := sdk.CrossChainID(0x0001)
	ibc.SetSourceChainID(sourceChainID)

	require.Equal(t, sourceChainID, ibc.GetSourceChainID())
}

func TestRegisterCrossChainChannel(t *testing.T) {
	require.NoError(t, ibc.RegisterCrossChainChannel("bind", sdk.CrossChainChannelID(1)))
	require.NoError(t, ibc.RegisterCrossChainChannel("transfer", sdk.CrossChainChannelID(2)))
	require.NoError(t, ibc.RegisterCrossChainChannel("timeout", sdk.CrossChainChannelID(3)))
	require.NoError(t, ibc.RegisterCrossChainChannel("staking", sdk.CrossChainChannelID(4)))
	require.Error(t, ibc.RegisterCrossChainChannel("staking", sdk.CrossChainChannelID(5)))
	require.Error(t, ibc.RegisterCrossChainChannel("staking-new", sdk.CrossChainChannelID(4)))

	channeID, err := ibc.GetChannelID("transfer")
	require.NoError(t, err)
	require.Equal(t, sdk.CrossChainChannelID(2), channeID)

	channeID, err = ibc.GetChannelID("staking")
	require.NoError(t, err)
	require.Equal(t, sdk.CrossChainChannelID(4), channeID)
}

func TestRegisterDestChainID(t *testing.T) {
	require.NoError(t, ibc.RegisterDestChainID("bsc", sdk.CrossChainID(1)))
	require.NoError(t, ibc.RegisterDestChainID("ethereum", sdk.CrossChainID(2)))
	require.NoError(t, ibc.RegisterDestChainID("btc", sdk.CrossChainID(3)))
	require.NoError(t, ibc.RegisterDestChainID("cosmos", sdk.CrossChainID(4)))
	require.Error(t, ibc.RegisterDestChainID("cosmos", sdk.CrossChainID(5)))
	require.Error(t, ibc.RegisterDestChainID("mock", sdk.CrossChainID(4)))

	destChainID, err := ibc.GetDestChainID("bsc")
	require.NoError(t, err)
	require.Equal(t, sdk.CrossChainID(1), destChainID)

	destChainID, err = ibc.GetDestChainID("btc")
	require.NoError(t, err)
	require.Equal(t, sdk.CrossChainID(3), destChainID)
}

func TestCrossChainID(t *testing.T) {
	chainID, err := ibc.ParseCrossChainID("123")
	require.NoError(t, err)
	require.Equal(t, sdk.CrossChainID(123), chainID)

	_, err = ibc.ParseCrossChainID("65537")
	require.Error(t, err)
}
