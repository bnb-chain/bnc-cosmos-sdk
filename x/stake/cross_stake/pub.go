package cross_stake

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/pubsub"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
)

const (
	CrossStakeTopic = pubsub.Topic("cross-stake")

	CrossStakeDelegateType         string = "CSD"
	CrossStakeUndelegateType       string = "CSU"
	CrossStakeClaimRewardType      string = "CSCR"
	CrossStakeClaimUndelegatedType string = "CSCU"
	CrossStakeReinvestType         string = "CSRI"
	CrossStakeRedelegateType       string = "CSRD"
)

type CrossStakeEvent struct {
	TxHash           string
	ChainId          string
	Type             string
	Delegator        string
	Receiver         string
	ValidatorSrc     string
	ValidatorDst     string
	OracleRelayerFee int64
	BSCRelayerFee    int64
}

type CrossReceiver struct {
	Addr   string
	Amount int64
}

func (event CrossStakeEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

func publishCrossChainEvent(ctx types.Context, keeper keeper.Keeper, delegator string, receiver string, valSrc string,
	valDst string, eventType string, oracleRelayerFee int64, bSCRelayerFee int64) {
	chainId := keeper.ScKeeper.BscSideChainId(ctx)
	if keeper.PbsbServer != nil {
		txHash := ctx.Value(baseapp.TxHashKey)
		if txHashStr, ok := txHash.(string); ok {
			event := CrossStakeEvent{
				TxHash:           txHashStr,
				ChainId:          chainId,
				Type:             eventType,
				Delegator:        delegator,
				Receiver:         receiver,
				ValidatorSrc:     valSrc,
				ValidatorDst:     valDst,
				OracleRelayerFee: oracleRelayerFee,
				BSCRelayerFee:    bSCRelayerFee,
			}
			keeper.PbsbServer.Publish(event)
		} else {
			ctx.Logger().Error("failed to get tx hash, will not publish cross stake event ")
		}
	}
}
