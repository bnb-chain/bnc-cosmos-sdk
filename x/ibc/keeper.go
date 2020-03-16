package ibc

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC Keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	codespace sdk.CodespaceType
}

func NewKeeper(storeKey sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  storeKey,
		codespace: codespace,
	}
}

func (k Keeper) CreateIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID, value []byte) sdk.Error {
	sequence := k.GetNextSequence(ctx, destChainID, channelID)
	key := buildIBCPackageKey(sdk.GetSourceChainID(), destChainID, channelID, sequence)
	kvStore := ctx.KVStore(k.storeKey)
	if kvStore.Has(key) {
		return ErrDuplicatedSequence(DefaultCodespace, "duplicated sequence")
	}
	kvStore.Set(key, value)
	k.incrSequence(ctx, destChainID, channelID)
	return nil
}

func (k *Keeper) GetIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID, sequence uint64) []byte {
	kvStore := ctx.KVStore(k.storeKey)
	key := buildIBCPackageKey(sdk.GetSourceChainID(), destChainID, channelID, sequence)
	return kvStore.Get(key)
}

func (k Keeper) CleanupIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID, confirmedSequence uint64) {
	prefixKey := buildIBCPackageKeyPrefix(sdk.GetSourceChainID(), destChainID, channelID)
	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, prefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		packageKey := iterator.Key()
		if len(packageKey) != prefixLength+sourceChainIDLength+destChainIDLength+channelIDLength+sequenceLength {
			continue
		}
		sequence := binary.BigEndian.Uint64(packageKey[prefixLength+sourceChainIDLength+destChainIDLength+channelIDLength:])
		if sequence > confirmedSequence {
			break
		}
		kvStore.Delete(packageKey)
	}
}

func (k *Keeper) GetNextSequence(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) incrSequence(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID) {
	sequence := k.GetNextSequence(ctx, destChainID, channelID)

	sequenceBytes := make([]byte, sequenceLength)
	binary.BigEndian.PutUint64(sequenceBytes, sequence+1)

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(buildChannelSequenceKey(destChainID, channelID), sequenceBytes)
}
