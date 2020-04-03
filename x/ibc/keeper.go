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

func (k Keeper) CreateIBCPackage(ctx sdk.Context, destChainName string, channelName string, value []byte) (uint64, sdk.Error) {
	destIbcChainID, err := GetDestIbcChainID(destChainName)
	if err != nil {
		return 0, sdk.ErrInternal(err.Error())
	}
	channelID, err := GetChannelID(channelName)
	if err != nil {
		return 0, sdk.ErrInternal(err.Error())
	}

	sequence := k.getSequence(ctx, destIbcChainID, channelID)
	key := buildIBCPackageKey(GetSrcIbcChainID(), destIbcChainID, channelID, sequence)
	kvStore := ctx.KVStore(k.storeKey)
	if kvStore.Has(key) {
		return 0, ErrDuplicatedSequence(DefaultCodespace, "duplicated sequence")
	}
	kvStore.Set(key, value)
	k.incrSequence(ctx, destIbcChainID, channelID)
	return sequence, nil
}

func (k *Keeper) GetIBCPackage(ctx sdk.Context, destChainName string, channelName string, sequence uint64) ([]byte, error) {
	destChainID, err := GetDestIbcChainID(destChainName)
	if err != nil {
		return nil, err
	}
	channelID, err := GetChannelID(channelName)
	if err != nil {
		return nil, err
	}

	kvStore := ctx.KVStore(k.storeKey)
	key := buildIBCPackageKey(GetSrcIbcChainID(), destChainID, channelID, sequence)
	return kvStore.Get(key), nil
}

func (k Keeper) CleanupIBCPackage(ctx sdk.Context, destChainName string, channelName string, confirmedSequence uint64) {
	destChainID, err := GetDestIbcChainID(destChainName)
	if err != nil {
		return
	}
	channelID, err := GetChannelID(channelName)
	if err != nil {
		return
	}
	prefixKey := buildIBCPackageKeyPrefix(GetSrcIbcChainID(), destChainID, channelID)
	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, prefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		packageKey := iterator.Key()
		if len(packageKey) != totalPackageKeyLength {
			continue
		}
		sequence := binary.BigEndian.Uint64(packageKey[totalPackageKeyLength-sequenceLength:])
		if sequence > confirmedSequence {
			break
		}
		kvStore.Delete(packageKey)
	}
}

func (k *Keeper) getSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) incrSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID) {
	var sequence uint64
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID))
	if bz == nil {
		sequence = 0
	} else {
		sequence = binary.BigEndian.Uint64(bz)
	}

	sequenceBytes := make([]byte, sequenceLength)
	binary.BigEndian.PutUint64(sequenceBytes, sequence+1)
	kvStore.Set(buildChannelSequenceKey(destChainID, channelID), sequenceBytes)
}
