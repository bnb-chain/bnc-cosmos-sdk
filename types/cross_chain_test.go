package types_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func TestParseChannelID(t *testing.T) {
	channelID, err := types.ParseIbcChannelID("12")
	require.NoError(t, err)
	require.Equal(t, types.IbcChannelID(12), channelID)

	_, err = types.ParseIbcChannelID("1024")
	require.Error(t, err)
}

func TestParseCrossChainID(t *testing.T) {
	chainID, err := types.ParseIbcChainID("12")
	require.NoError(t, err)
	require.Equal(t, types.IbcChainID(12), chainID)

	chainID, err = types.ParseIbcChainID("10000")
	require.NoError(t, err)
	require.Equal(t, types.IbcChainID(10000), chainID)

	_, err = types.ParseIbcChainID("65536")
	require.Error(t, err)
}
