package ibc

import (
	"encoding/binary"
	"fmt"

	codec "github.com/cosmos/cosmos-sdk/codec"
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

func (k Keeper) CreateIBCPackage(ctx sdk.Context, destinationChainID string, channelID ChannelID, value []byte) sdk.Error {
	sequence, err:= k.getSequence(ctx, channelID)
	if err != nil {
		return ErrUnsupportedChannel(DefaultCodespace, err.Error())
	}

	if len(ctx.ChainID()) > 32 || len(destinationChainID) > 32 {
		return ErrChainIDTooLong(DefaultCodespace, "chainID should be no more than 32")
	}

	key := BuildIBCPackageKey(ctx.ChainID(), destinationChainID, channelID, sequence)

	kvStore := ctx.KVStore(k.storeKey)
	if kvStore.Has(key) {
		panic("duplicated key for cross chain package")
	}
	kvStore.Set(key, value)

	k.incrSequence(ctx, channelID)
	return nil
}

func (k Keeper) CleanupIBCPackage(ctx sdk.Context, destinationChainID string, channelID int8, confirmedSequence int64) sdk.Error {
	if len(ctx.ChainID()) > 32 || len(destinationChainID) > 32 {
		return ErrChainIDTooLong(DefaultCodespace, "chainID should be no more than 32")
	}

	prefixKey := buildIBCPackageKeyPrefix(ctx.ChainID(), destinationChainID, channelID)
	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, prefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		packageKey := iterator.Key()
		if len(packageKey) != 32+32+1+8 {
			continue
		}
		sequence := int64(binary.BigEndian.Uint64(packageKey[65:]))
		if sequence > confirmedSequence {
			break
		}
		//TODO double check if the delete operation affect iteration
		kvStore.Delete(packageKey)
	}

	return nil
}

func (k *Keeper) getSequence(ctx sdk.Context, channelID ChannelID) (int64, error) {
	switch channelID {
	case BindChannelID:
		return k.getBindChannelSequence(ctx), nil
	case TransferChannelID:
		return k.getTransferChannelSequence(ctx), nil
	case TimeoutChannelID:
		return k.getTimeoutChannelSequence(ctx), nil
	case StakingChannelID:
		return k.getStakingChannelSequence(ctx), nil
	default:
		return 0, fmt.Errorf("unsupported channelID")
	}
}

func (k *Keeper) incrSequence(ctx sdk.Context, channelID ChannelID) {
	switch channelID {
	case BindChannelID:
		k.incrBindChannelSequence(ctx)
	case TransferChannelID:
		k.incrTransferChannelSequence(ctx)
	case TimeoutChannelID:
		k.incrTimeoutChannelSequence(ctx)
	case StakingChannelID:
		k.incrStakingChannelSequence(ctx)
	}
}

func (k *Keeper) getBindChannelSequence(ctx sdk.Context) int64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(KeyForBindChannelSequence)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k *Keeper) incrBindChannelSequence(ctx sdk.Context) {
	sequence := k.getBindChannelSequence(ctx)
	sequenceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceBytes, uint64(sequence+1))

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(KeyForBindChannelSequence, sequenceBytes)
}

func (k *Keeper) getTransferChannelSequence(ctx sdk.Context) int64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(KeyForTransferChannelSequence)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k *Keeper) incrTransferChannelSequence(ctx sdk.Context) {
	sequence := k.getTransferChannelSequence(ctx)
	sequenceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceBytes, uint64(sequence+1))

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(KeyForTransferChannelSequence, sequenceBytes)
}

func (k *Keeper) getTimeoutChannelSequence(ctx sdk.Context) int64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(KeyForTimeoutChannelSequence)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k *Keeper) incrTimeoutChannelSequence(ctx sdk.Context) {
	sequence := k.getTimeoutChannelSequence(ctx)
	sequenceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceBytes, uint64(sequence+1))

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(KeyForTimeoutChannelSequence, sequenceBytes)
}

func (k *Keeper) getStakingChannelSequence(ctx sdk.Context) int64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(KeyForStakingChannelSequence)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k *Keeper) incrStakingChannelSequence(ctx sdk.Context) {
	sequence := k.getStakingChannelSequence(ctx)
	sequenceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceBytes, uint64(sequence+1))

	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(KeyForStakingChannelSequence, sequenceBytes)
}
