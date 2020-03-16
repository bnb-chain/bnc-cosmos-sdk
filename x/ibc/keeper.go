package ibc

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC Keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		codespace: codespace,
	}
}

func (k Keeper) CreateIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID, value []byte) sdk.Error {
	sequence := k.GetSequence(ctx, destChainID, channelID)
	key := buildIBCPackageKey(sdk.GetSourceChainID(), destChainID, channelID, sequence)
	kvStore := ctx.KVStore(k.storeKey)
	if kvStore.Has(key) {
		return ErrDuplicatedSequence(DefaultCodespace, "duplicated sequence")
	}
	kvStore.Set(key, value)
	k.incrSequence(ctx, destChainID, channelID)
	return nil
}

func (k Keeper) CleanupIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID, confirmedSequence uint64) {
	prefixKey := buildIBCPackageKeyPrefix(sdk.GetSourceChainID(), destChainID, channelID)
	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, prefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		packageKey := iterator.Key()
		if len(packageKey) != sourceChainIDLength+destChainIDLength+channelIDLength+sequenceLength {
			continue
		}
		sequence := binary.BigEndian.Uint64(packageKey[sourceChainIDLength+destChainIDLength+channelIDLength:])
		if sequence > confirmedSequence {
			break
		}
		kvStore.Delete(packageKey)
	}
}

func (k *Keeper) GetSequence(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) incrSequence(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID) {
	sequence := k.GetSequence(ctx, destChainID, channelID)

	sequenceBytes := make([]byte, sequenceLength)
	binary.BigEndian.PutUint64(sequenceBytes, sequence+1)

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(buildChannelSequenceKey(destChainID, channelID), sequenceBytes)
}
