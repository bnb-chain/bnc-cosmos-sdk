package sidechain

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	pTypes "github.com/cosmos/cosmos-sdk/x/paramHub/types"
	"github.com/cosmos/cosmos-sdk/x/sidechain/types"
)

const (
	EnableOrDisableChannelKey = "enableOrDisableChannel"
	AddOrUpdateChannelKey     = "addOrUpdateChannel"
	AddOperatorKey            = "addOperator"
	DeleteOperatorKey         = "deleteOperator"
)

var (
	SystemRewardContractAddr, _ = hex.DecodeString("0000000000000000000000000000000000001002")
	CrossChainContractAddr, _   = hex.DecodeString("0000000000000000000000000000000000002000")
)

func (k *Keeper) CreateNewCrossChainChannel(ctx sdk.Context, sideChainId sdk.ChainID, channelId sdk.ChannelID, rewardConfig sdk.RewardConfig, handleContract []byte) (seq uint64, sdkErr sdk.Error) {
	valueBytes := []byte{byte(channelId), byte(rewardConfig)}
	valueBytes = append(valueBytes, handleContract...)

	return k.sendParamChangeToIbc(ctx, sideChainId, AddOrUpdateChannelKey, valueBytes, CrossChainContractAddr)
}

func (k *Keeper) AddSystemRewardOperator(ctx sdk.Context, sideChainId sdk.ChainID, operator sdk.SmartChainAddress) (seq uint64, sdkErr sdk.Error) {
	return k.sendParamChangeToIbc(ctx, sideChainId, AddOperatorKey, operator[:], SystemRewardContractAddr)
}

func (k *Keeper) sendParamChangeToIbc(ctx sdk.Context, sideChainId sdk.ChainID, key string, valueBytes []byte, targetBytes []byte) (seq uint64, sdkErr sdk.Error) {
	paramChange := pTypes.CSCParamChange{
		Key:         key,
		ValueBytes:  valueBytes,
		TargetBytes: targetBytes,
	}

	bz, err := rlp.EncodeToBytes(&paramChange)
	if err != nil {
		return 0, sdk.ErrInternal("failed to encode paramChange")
	}
	return k.ibcKeeper.CreateRawIBCPackageById(ctx, sideChainId, types.GovChannelId, sdk.SynCrossChainPackageType, bz)
}
