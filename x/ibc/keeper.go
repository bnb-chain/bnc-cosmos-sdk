package ibc

import (
	"encoding/binary"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC Keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	codespace sdk.CodespaceType

	cfg              *crossChainConfig
	packageCollector *packageCollector
}

func NewKeeper(storeKey sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:         storeKey,
		codespace:        codespace,
		cfg:              newCrossChainCfg(),
		packageCollector: newPackageCollector(),
	}
}

func (k Keeper) CreateIBCPackage(ctx sdk.Context, destChainName string, channelName string, value []byte) (uint64, sdk.Error) {
	destIbcChainID, err := k.GetDestIbcChainID(destChainName)
	if err != nil {
		return 0, sdk.ErrInternal(err.Error())
	}
	channelID, err := k.GetChannelID(channelName)
	if err != nil {
		return 0, sdk.ErrInternal(err.Error())
	}

	sequence := k.getSequence(ctx, destIbcChainID, channelID)
	key := buildIBCPackageKey(k.GetSrcIbcChainID(), destIbcChainID, channelID, sequence)
	kvStore := ctx.KVStore(k.storeKey)
	if kvStore.Has(key) {
		return 0, ErrDuplicatedSequence(DefaultCodespace, "duplicated sequence")
	}
	kvStore.Set(key, value)
	k.incrSequence(ctx, destIbcChainID, channelID)

	k.packageCollector.collectedPackages = append(k.packageCollector.collectedPackages, packageRecord{
		destChainName: destChainName,
		destChainID:   destIbcChainID,
		channelID:     channelID,
		sequence:      sequence,
	})
	return sequence, nil
}

func (k *Keeper) GetIBCPackage(ctx sdk.Context, destChainName string, channelName string, sequence uint64) ([]byte, error) {
	destChainID, err := k.GetDestIbcChainID(destChainName)
	if err != nil {
		return nil, err
	}
	channelID, err := k.GetChannelID(channelName)
	if err != nil {
		return nil, err
	}

	kvStore := ctx.KVStore(k.storeKey)
	key := buildIBCPackageKey(k.GetSrcIbcChainID(), destChainID, channelID, sequence)
	return kvStore.Get(key), nil
}

func (k Keeper) CleanupIBCPackage(ctx sdk.Context, destChainName string, channelName string, confirmedSequence uint64) {
	destChainID, err := k.GetDestIbcChainID(destChainName)
	if err != nil {
		return
	}
	channelID, err := k.GetChannelID(channelName)
	if err != nil {
		return
	}
	prefixKey := buildIBCPackageKeyPrefix(k.GetSrcIbcChainID(), destChainID, channelID)
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

func (k *Keeper) RegisterChannel(name string, id sdk.IbcChannelID) error {
	_, ok := k.cfg.nameToChannelID[name]
	if ok {
		return fmt.Errorf("duplicated channel name")
	}
	_, ok = k.cfg.channelIDToName[id]
	if ok {
		return fmt.Errorf("duplicated channel id")
	}
	k.cfg.nameToChannelID[name] = id
	k.cfg.channelIDToName[id] = name
	return nil
}

// internally, we use name as the id of the chain, must be unique
func (k *Keeper) RegisterDestChain(name string, ibcChainID sdk.IbcChainID) error {
	if strings.Contains(name, separator) {
		return fmt.Errorf("destination chain name should not contains %s", separator)
	}
	_, ok := k.cfg.destChainNameToID[name]
	if ok {
		return fmt.Errorf("duplicated destination chain name")
	}
	_, ok = k.cfg.destChainIDToName[ibcChainID]
	if ok {
		return fmt.Errorf("duplicated destination chain ibcChainID")
	}
	k.cfg.destChainNameToID[name] = ibcChainID
	k.cfg.destChainIDToName[ibcChainID] = name
	return nil
}

func (k *Keeper) GetChannelID(channelName string) (sdk.IbcChannelID, error) {
	id, ok := k.cfg.nameToChannelID[channelName]
	if !ok {
		return sdk.IbcChannelID(0), fmt.Errorf("non-existing channel")
	}
	return id, nil
}

func (k *Keeper) SetSrcIbcChainID(srcIbcChainID sdk.IbcChainID) {
	k.cfg.srcIbcChainID = srcIbcChainID
}

func (k *Keeper) GetSrcIbcChainID() sdk.IbcChainID {
	return k.cfg.srcIbcChainID
}

func (k *Keeper) GetDestIbcChainID(name string) (sdk.IbcChainID, error) {
	destChainID, exist := k.cfg.destChainNameToID[name]
	if !exist {
		return sdk.IbcChainID(0), fmt.Errorf("non-existing destination ibcChainID")
	}
	return destChainID, nil
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
