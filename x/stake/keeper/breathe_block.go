package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

func (k Keeper) SetBreatheBlockHeight(ctx sdk.Context, height int64, blockTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinaryBare(height)
	if err != nil {
		panic(err)
	}
	store.Set(GetBreatheBlockHeightKey(blockTime),bz)
}

func (k Keeper) GetBreatheBlockHeight(ctx sdk.Context, indexCountBackwards int) (height int64, found bool){
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStoreReversePrefixIterator(store, BreatheBlockHeightKey)
	defer iterator.Close()

	i := 1
	for ;iterator.Valid();iterator.Next() {
		if indexCountBackwards == i {
			k.cdc.MustUnmarshalBinaryBare(iterator.Value(),&height)
			return height,true
		}
		i++
	}
	return height, false
}


