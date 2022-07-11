package cross_stake

import (
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	CrossStakeTopic = pubsub.Topic("cross-stake")

	CrossStakeDelegateType               string = "CSD"
	CrossStakeUndelegateType             string = "CSU"
	CrossStakeTransferOutRewardType      string = "CSTR"
	CrossStakeTransferOutUndelegatedType string = "CSTU"
	CrossStakeRedelegateType             string = "CSRD"
)

type CrossStakeEvent struct {
	ChainId      string
	Type         string
	Delegator    sdk.AccAddress
	ValidatorSrc sdk.ValAddress
	ValidatorDst sdk.ValAddress
	RelayFee     int64
}

func (event CrossStakeEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type TransferOutRewardEvent struct {
	ChainId       string
	Type          string
	Delegators    []sdk.AccAddress
	Receivers     []types.SmartChainAddress
	Amounts       []int64
	BSCRelayerFee int64
}

func (event TransferOutRewardEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type TransferOutUndelegatedEvent struct {
	ChainId       string
	Type          string
	Delegator     sdk.AccAddress
	Receiver      types.SmartChainAddress
	Amount        int64
	BSCRelayerFee int64
}

func (event TransferOutUndelegatedEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

func PublishCrossChainEvent(ctx sdk.Context, keeper keeper.Keeper, delegator sdk.AccAddress, valSrc sdk.ValAddress,
	valDst sdk.ValAddress, eventType string, relayFee int64) {
	chainId := keeper.ScKeeper.BscSideChainId(ctx)
	if keeper.PbsbServer != nil {
		event := CrossStakeEvent{
			ChainId:      chainId,
			Type:         eventType,
			Delegator:    delegator,
			ValidatorSrc: valSrc,
			ValidatorDst: valDst,
			RelayFee:     relayFee,
		}
		keeper.PbsbServer.Publish(event)
	}
}
