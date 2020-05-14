package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/oracle/types"
)

func RegisterUpgradeBeginBlocker(keeper Keeper) {
	sdk.UpgradeMgr.RegisterBeginBlocker(sdk.LaunchBscUpgrade, func(ctx sdk.Context) {
		keeper.SetParams(ctx, types.Params{ConsensusNeeded: types.DefaultConsensusNeeded})
	})
}
