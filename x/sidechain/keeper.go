package sidechain

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
)

var (
	separator = "::"
)

type Keeper struct {
	storeKey   sdk.StoreKey
	paramspace params.Subspace
	cfg        *crossChainConfig
	cdc        *codec.Codec

	govKeeper *gov.Keeper
	ibcKeeper IbcKeeper
}

type IbcKeeper interface {
	CreateRawIBCPackageById(ctx sdk.Context, destIbcChainID sdk.IbcChainID, channelID sdk.IbcChannelID,
		packageType sdk.CrossChainPackageType, packageLoad []byte) (uint64, sdk.Error)
}

func NewKeeper(storeKey sdk.StoreKey, paramspace params.Subspace, cdc *codec.Codec) Keeper {
	return Keeper{
		storeKey:   storeKey,
		paramspace: paramspace.WithTypeTable(ParamTypeTable()),
		cfg:        newCrossChainCfg(),
		cdc:        cdc,
	}
}

func (k *Keeper) SetGovKeeper(govKeeper *gov.Keeper) {
	k.govKeeper = govKeeper
}

func (k *Keeper) SetIbcKeeper(ibcKeeper IbcKeeper) {
	k.ibcKeeper = ibcKeeper
}

func (k *Keeper) PrepareCtxForSideChain(ctx sdk.Context, sideChainId string) (sdk.Context, error) {
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

func (k *Keeper) GetAllSideChainPrefixes(ctx sdk.Context) ([]string, [][]byte) {
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

func (k *Keeper) RegisterChannel(name string, id sdk.IbcChannelID, app sdk.CrossChainApplication) error {
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
	k.cfg.channelIDToApp[id] = app
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

func (k *Keeper) SetChannelSendPermission(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID, permission sdk.ChannelPermission) {
	kvStore := ctx.KVStore(k.storeKey)
	kvStore.Set(buildChannelPermissionKey(destChainID, channelID), []byte{byte(permission)})
}

func (k *Keeper) GetChannelSendPermission(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID) sdk.ChannelPermission {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelPermissionKey(destChainID, channelID))
	if bz == nil {
		return sdk.ChannelForbidden
	}
	return sdk.ChannelPermission(bz[0])
}

func (k *Keeper) GetChannelSendPermissions(ctx sdk.Context, destChainID sdk.IbcChainID) map[sdk.IbcChannelID]sdk.ChannelPermission {
	kvStore := ctx.KVStore(k.storeKey).Prefix(buildChannelPermissionsPrefixKey(destChainID))
	ite := kvStore.Iterator(nil, nil)
	permissions := make(map[sdk.IbcChannelID]sdk.ChannelPermission, 0)
	for ; ite.Valid(); ite.Next() {
		key := ite.Key()
		if len(key) < 1 {
			continue
		}
		channelId := sdk.IbcChannelID(key[0])
		value := ite.Value()
		permissions[channelId] = sdk.ChannelPermission(value[0])
	}
	return permissions
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
		return sdk.IbcChainID(0), fmt.Errorf("non-existing destination ibcChainName ")
	}
	return destChainID, nil
}

func (k *Keeper) GetDestIbcChainName(id sdk.IbcChainID) (string, error) {
	destChainName, exist := k.cfg.destChainIDToName[id]
	if !exist {
		return "", fmt.Errorf("non-existing destination ibcChainID")
	}
	return destChainName, nil
}

func (k *Keeper) GetSendSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID) uint64 {
	return k.getSequence(ctx, destChainID, channelID, PrefixForSendSequenceKey)
}

func (k *Keeper) IncrSendSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID) {
	k.incrSequence(ctx, destChainID, channelID, PrefixForSendSequenceKey)
}

func (k *Keeper) GetReceiveSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID) uint64 {
	return k.getSequence(ctx, destChainID, channelID, PrefixForReceiveSequenceKey)
}

func (k *Keeper) IncrReceiveSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID) {
	k.incrSequence(ctx, destChainID, channelID, PrefixForReceiveSequenceKey)
}

func (k *Keeper) GetCrossChainApp(ctx sdk.Context, channelID sdk.IbcChannelID) sdk.CrossChainApplication {
	return k.cfg.channelIDToApp[channelID]
}

func (k *Keeper) getSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID, prefix []byte) uint64 {
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID, prefix))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) incrSequence(ctx sdk.Context, destChainID sdk.IbcChainID, channelID sdk.IbcChannelID, prefix []byte) {
	var sequence uint64
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID, prefix))
	if bz == nil {
		sequence = 0
	} else {
		sequence = binary.BigEndian.Uint64(bz)
	}

	sequenceBytes := make([]byte, sequenceLength)
	binary.BigEndian.PutUint64(sequenceBytes, sequence+1)
	kvStore.Set(buildChannelSequenceKey(destChainID, channelID, prefix), sequenceBytes)
}

func EndBlock(ctx sdk.Context, k Keeper) {
	if sdk.IsUpgrade(sdk.LaunchBscUpgrade) && k.govKeeper != nil {
		chanPermissions := k.getLastChanPermissionChanges(ctx)
		// should in reverse order
		for j := len(chanPermissions) - 1; j >= 0; j-- {
			change := chanPermissions[j]
			// must exist
			id, _ := k.cfg.destChainNameToID[change.SideChainId]
			k.SetChannelSendPermission(ctx, id, change.ChannelId, change.Permission)
			_, err := k.SaveChannelSettingChangeToIbc(ctx, id, change.ChannelId, change.Permission)
			if err != nil {
				ctx.Logger().With("module", "side_chain").Error("failed to write cross chain channel permission change message ",
					"err", err)
			}
		}
	}
	return
}
