package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const IbcChannelName = "staking"
const IbcChannelId = sdk.IbcChannelID(8)

func (k Keeper) SaveValidatorSetToIbc(ctx sdk.Context, sideChainId string, ibcVals types.IbcValidatorSet) (seq uint64, sdkErr sdk.Error) {
	bz, err := ibcVals.Serialize()
	if err != nil {
		k.Logger(ctx).Error("serialize failed: " + err.Error())
		return 0, sdk.ErrInternal(err.Error())
	}
	return k.ibcKeeper.CreateIBCPackage(ctx, sideChainId, IbcChannelName, bz)
}
