package stake_migration

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/cross_stake"
)

func PublishStakeMigrationEvent(ctx sdk.Context, keeper cross_stake.Keeper, from string, to []pubsub.CrossReceiver, symbol string,
	eventType string, relayerFee int64,
) {
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
