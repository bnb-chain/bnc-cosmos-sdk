package ibc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func TestInitCrossChainID(t *testing.T) {
	sourceChainID := sdk.IbcChainID(0x0001)
	ibc.SetSrcIbcChainID(sourceChainID)

	require.Equal(t, sourceChainID, ibc.GetSrcIbcChainID())
}

func TestRegisterCrossChainChannel(t *testing.T) {
	require.NoError(t, ibc.RegisterChannel("bind", sdk.IbcChannelID(1)))
	require.NoError(t, ibc.RegisterChannel("transfer", sdk.IbcChannelID(2)))
	require.NoError(t, ibc.RegisterChannel("timeout", sdk.IbcChannelID(3)))
	require.NoError(t, ibc.RegisterChannel("staking", sdk.IbcChannelID(4)))
	require.Error(t, ibc.RegisterChannel("staking", sdk.IbcChannelID(5)))
	require.Error(t, ibc.RegisterChannel("staking-new", sdk.IbcChannelID(4)))

	channeID, err := ibc.GetChannelID("transfer")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChannelID(2), channeID)

	channeID, err = ibc.GetChannelID("staking")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChannelID(4), channeID)
}

func TestRegisterDestChainID(t *testing.T) {
	require.NoError(t, ibc.RegisterDestChain("bsc", sdk.IbcChainID(1)))
	require.NoError(t, ibc.RegisterDestChain("ethereum", sdk.IbcChainID(2)))
	require.NoError(t, ibc.RegisterDestChain("btc", sdk.IbcChainID(3)))
	require.NoError(t, ibc.RegisterDestChain("cosmos", sdk.IbcChainID(4)))
	require.Error(t, ibc.RegisterDestChain("cosmos", sdk.IbcChainID(5)))
	require.Error(t, ibc.RegisterDestChain("mock", sdk.IbcChainID(4)))

	destChainID, err := ibc.GetDestIbcChainID("bsc")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChainID(1), destChainID)

	destChainID, err = ibc.GetDestIbcChainID("btc")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChainID(3), destChainID)
}

func TestCrossChainID(t *testing.T) {
	chainID, err := ibc.ParseCrossChainID("123")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChainID(123), chainID)

	_, err = ibc.ParseCrossChainID("65537")
	require.Error(t, err)
}
