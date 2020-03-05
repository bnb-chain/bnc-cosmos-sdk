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
	sequence, err:= k.getSequence(ctx, channelID)
	if err != nil {
		return ErrUnsupportedChannel(DefaultCodespace, err.Error())
	}

	key := BuildIBCPackageKey(sdk.GetSourceChainID(), destChainID, channelID, sequence)

	kvStore := ctx.KVStore(k.storeKey)
	if kvStore.Has(key) {
		panic("duplicated key for cross chain package")
	}
	kvStore.Set(key, value)

	k.incrSequence(ctx, channelID)
	return nil
}

func (k Keeper) CleanupIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID, confirmedSequence uint64) sdk.Error {
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
		//TODO double check if the delete operation affect iteration
		kvStore.Delete(packageKey)
	}

	return nil
}

func (k *Keeper) getSequence(ctx sdk.Context, channelID sdk.ChannelID) (uint64, error) {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(channelID))
	if bz == nil {
		return 0, nil
	}
	return binary.BigEndian.Uint64(bz), nil
}

func (k *Keeper) incrSequence(ctx sdk.Context, channelID sdk.ChannelID) {
	sequence, _ := k.getSequence(ctx, channelID)

	sequenceBytes := make([]byte, sequenceLength)
	binary.BigEndian.PutUint64(sequenceBytes ,sequence+1)

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(buildChannelSequenceKey(channelID), sequenceBytes)
}
