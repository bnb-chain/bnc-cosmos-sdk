package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestInitCrossChainID(t *testing.T) {
	sourceChainID := sdk.IbcChainID(0x0001)
	_, keeper := createTestInput(t, true)
	keeper.SetSrcIbcChainID(sourceChainID)

	require.Equal(t, sourceChainID, keeper.GetSrcIbcChainID())
}

func TestRegisterCrossChainChannel(t *testing.T) {
	_, keeper := createTestInput(t, true)
	require.NoError(t, keeper.RegisterChannel("bind", sdk.IbcChannelID(1)))
	require.NoError(t, keeper.RegisterChannel("transfer", sdk.IbcChannelID(2)))
	require.NoError(t, keeper.RegisterChannel("timeout", sdk.IbcChannelID(3)))
	require.NoError(t, keeper.RegisterChannel("staking", sdk.IbcChannelID(4)))
	require.Error(t, keeper.RegisterChannel("staking", sdk.IbcChannelID(5)))
	require.Error(t, keeper.RegisterChannel("staking-new", sdk.IbcChannelID(4)))

	channeID, err := keeper.GetChannelID("transfer")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChannelID(2), channeID)

	channeID, err = keeper.GetChannelID("staking")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChannelID(4), channeID)
}

func TestRegisterDestChainID(t *testing.T) {
	_, keeper := createTestInput(t, true)
	require.NoError(t, keeper.RegisterDestChain("bsc", sdk.IbcChainID(1)))
	require.NoError(t, keeper.RegisterDestChain("ethereum", sdk.IbcChainID(2)))
	require.NoError(t, keeper.RegisterDestChain("btc", sdk.IbcChainID(3)))
	require.NoError(t, keeper.RegisterDestChain("cosmos", sdk.IbcChainID(4)))
	require.Error(t, keeper.RegisterDestChain("cosmos", sdk.IbcChainID(5)))
	require.Error(t, keeper.RegisterDestChain("mock", sdk.IbcChainID(4)))

	destChainID, err := keeper.GetDestIbcChainID("bsc")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChainID(1), destChainID)

	destChainID, err = keeper.GetDestIbcChainID("btc")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChainID(3), destChainID)
}

func TestCrossChainID(t *testing.T) {
	chainID, err := sdk.ParseIbcChainID("123")
	require.NoError(t, err)
	require.Equal(t, sdk.IbcChainID(123), chainID)

	_, err = sdk.ParseIbcChainID("65537")
	require.Error(t, err)
}
