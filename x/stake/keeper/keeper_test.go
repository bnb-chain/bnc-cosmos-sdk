package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expParams := types.DefaultParams()
	// params changed or not activated yet, so different with default
	expParams.MinSelfDelegation = int64(10)
	expParams.MinDelegationChange = int64(2)
	expParams.BaseProposerRewardRatio = sdk.ZeroDec()
	expParams.BonusProposerRewardRatio = sdk.ZeroDec()
	expParams.MaxStakeSnapshots = uint16(0)
	expParams.FeeFromBscToBcRatio = sdk.ZeroDec()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	require.Equal(t, expParams, resParams)

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
