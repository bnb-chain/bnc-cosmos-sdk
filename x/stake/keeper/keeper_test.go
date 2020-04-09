package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestParams(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))

	//modify a params, save, and retrieve
	expParams.MaxValidators = 777
	keeper.SetParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	require.True(t, expParams.Equal(resParams))
}

func TestPool(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expPool := types.InitialPool()

	//check that the empty keeper loads the default
	resPool := keeper.GetPool(ctx)
	require.True(t, expPool.Equal(resPool))

	//modify a params, save, and retrieve
	expPool.BondedTokens = sdk.NewDec(777)
	keeper.SetPool(ctx, expPool)
	resPool = keeper.GetPool(ctx)
	require.True(t, expPool.Equal(resPool))
}

func TestKeeper_SetSideChainIdAndStorePrefix(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	scIds, scPrefixes := keeper.GetAllSideChainPrefixes(ctx)
	require.Equal(t, len(scIds),0)
	require.Equal(t, len(scPrefixes), 0)

	keeper.SetSideChainIdAndStorePrefix(ctx, "abc", []byte{0x11, 0x12})
	keeper.SetSideChainIdAndStorePrefix(ctx, "xyz", []byte{0xab})
	scIds, scPrefixes = keeper.GetAllSideChainPrefixes(ctx)
	require.Equal(t, len(scIds),2)
	require.Equal(t, len(scPrefixes), 2)
	require.Equal(t, scIds[0], "abc")
	require.Equal(t, scPrefixes[0], []byte{0x11, 0x12})
	require.Equal(t, scIds[1], "xyz")
	require.Equal(t, scPrefixes[1], []byte{0xab})
}
