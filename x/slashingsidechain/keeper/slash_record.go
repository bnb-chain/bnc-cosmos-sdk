package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

var SlashRecordKey = []byte{0x01}

func (k Keeper) GetSlashRecord(ctx sdk.Context, sideConsAddr []byte, sideHeight int64) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(getSlashRecordKey(sideConsAddr,sideHeight))
}

func (k Keeper) SetSlashRecord(ctx sdk.Context, sideConsAddr []byte, sideHeight int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(getSlashRecordKey(sideConsAddr,sideHeight),[]byte{})
}

func getSlashRecordKey(sideConsAddr []byte, sideHeight int64) []byte {
	return append(append(SlashRecordKey, sideConsAddr...),[]byte(strconv.FormatInt(sideHeight, 16))...)
}
