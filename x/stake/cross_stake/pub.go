package cross_stake

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func PublishCrossStakeEvent(ctx sdk.Context, keeper Keeper, from string, to []pubsub.CrossReceiver, symbol string,
	eventType string, relayerFee int64) {
	if keeper.PbsbServer != nil {
		txHash := ctx.Value(baseapp.TxHashKey)
		if txHashStr, ok := txHash.(string); ok {
			event := pubsub.CrossTransferEvent{
				TxHash:     txHashStr,
				ChainId:    keeper.DestChainName,
				RelayerFee: relayerFee,
				Type:       eventType,
				From:       from,
				Denom:      symbol,
				To:         to,
			}
			keeper.PbsbServer.Publish(event)
		} else {
			ctx.Logger().With("module", "stake").Error("failed to get txhash, will not publish cross transfer event ")
		}
	}
}
