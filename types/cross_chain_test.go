package types_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func TestParseChannelID(t *testing.T) {
	channelID, err := types.ParseCrossChainChannelID("12")
	require.NoError(t, err)
	require.Equal(t, types.CrossChainChannelID(12), channelID)

	_, err = types.ParseCrossChainChannelID("1024")
	require.Error(t, err)
}

func TestParseCrossChainID(t *testing.T) {
	chainID, err := types.ParseCrossChainID("12")
	require.NoError(t, err)
	require.Equal(t, types.CrossChainID(12), chainID)

	chainID, err = types.ParseCrossChainID("10000")
	require.NoError(t, err)
	require.Equal(t, types.CrossChainID(10000), chainID)

	_, err = types.ParseCrossChainID("65536")
	require.Error(t, err)
}
