package keeper

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetBreatheBlockHeight(ctx sdk.Context, height int64, blockTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(height))
	store.Set(GetBreatheBlockHeightKey(blockTime), bz)
}

func (k Keeper) GetBreatheBlockHeight(ctx sdk.Context, indexCountBackwards int) (height int64, found bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStoreReversePrefixIterator(store, BreatheBlockHeightKey)
	defer iterator.Close()

	i := 1
	for ; iterator.Valid(); iterator.Next() {
		if indexCountBackwards == i {
			height := int64(binary.BigEndian.Uint64(iterator.Value()))
			return height, true
		}
		i++
	}
	return height, false
}
