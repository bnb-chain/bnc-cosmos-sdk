package slashing

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) getSlashRecord(ctx sdk.Context, sideConsAddr []byte, sideHeight int64) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(getSlashRecordKey(sideConsAddr, sideHeight))
}

func (k Keeper) setSlashRecord(ctx sdk.Context, sideConsAddr []byte, sideHeight int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(getSlashRecordKey(sideConsAddr, sideHeight), []byte{})
}
