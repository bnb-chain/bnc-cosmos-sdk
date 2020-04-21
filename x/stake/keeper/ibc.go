package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const IbcChannelName = "staking"
const IbcChannelId = sdk.IbcChannelID(8)

func (k Keeper) SaveValidatorSetToIbc(ctx sdk.Context, sideChainId string, ibcVals types.IbcValidatorSet) (seq uint64, sdkErr sdk.Error) {
	if k.ibcKeeper == nil {
		return 0, sdk.ErrInternal("the keeper is not prepared for side chain")
	}
	bz, err := ibcVals.Serialize()
	if err != nil {
		k.Logger(ctx).Error("serialize failed: " + err.Error())
		return 0, sdk.ErrInternal(err.Error())
	}
	// prepend a flag 0x00
	bz = append([]byte{0x00}, bz...)
	return k.ibcKeeper.CreateIBCPackage(ctx, sideChainId, IbcChannelName, bz)
}

func (k Keeper) SaveJailedValidatorToIbc(ctx sdk.Context, sideChainId string, ibcVal types.IbcValidator) (seq uint64, sdkErr sdk.Error) {
	bz, err := ibcVal.Serialize()
	if err != nil {
		k.Logger(ctx).Error("serialize failed: " + err.Error())
		return 0, sdk.ErrInternal(err.Error())
	}
	newBz := addPrefix(byte(0x01),bz)
	return k.ibcKeeper.CreateIBCPackage(ctx, sideChainId, IbcChannelName, newBz)
}

func addPrefix(prefix byte,bz []byte) []byte {
	newBz := make([]byte, len(bz)+1)
	newBz[0] = prefix
	copy(newBz[1:], bz)
	return newBz
}
