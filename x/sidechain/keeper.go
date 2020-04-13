package sidechain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
}

func NewKeeper(storeKey sdk.StoreKey) Keeper {
	return Keeper{
		storeKey: storeKey,
	}
}

func (k Keeper) PrepareCtxForSideChain(ctx sdk.Context, sideChainId string) (sdk.Context, error) {
	storePrefix := k.GetSideChainStorePrefix(ctx, sideChainId)
	if storePrefix == nil {
		return sdk.Context{}, fmt.Errorf("invalid sideChainId: %s", sideChainId)
	}

	// add store prefix to ctx for side chain use
	return ctx.WithSideChainKeyPrefix(storePrefix), nil
}

// TODO: to support multi side chains in the future. We will enable a registration mechanism and add these chain ids to db.
// then we need to check if the sideChainId already exists
func (k Keeper) SetSideChainIdAndStorePrefix(ctx sdk.Context, sideChainId string, storePrefix []byte) {
	store := ctx.KVStore(k.storeKey)
	key := GetSideChainStorePrefixKey(sideChainId)
	store.Set(key, storePrefix)
}

// get side chain store key prefix
func (k Keeper) GetSideChainStorePrefix(ctx sdk.Context, sideChainId string) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(GetSideChainStorePrefixKey(sideChainId))
}

func (k Keeper) GetAllSideChainPrefixes(ctx sdk.Context) ([]string, [][]byte) {
	store := ctx.KVStore(k.storeKey)
	sideChainIds := make([]string, 0, 1)
	prefixes := make([][]byte, 0, 1)
	iterator := sdk.KVStorePrefixIterator(store, SideChainStorePrefixByIdKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		sideChainId := iterator.Key()[len(SideChainStorePrefixByIdKey):]
		sideChainIds = append(sideChainIds, string(sideChainId))
		prefixes = append(prefixes, iterator.Value())
	}
	return sideChainIds, prefixes
}
