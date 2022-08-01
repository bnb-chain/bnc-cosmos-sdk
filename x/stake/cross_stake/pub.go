package cross_stake

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func PublishCrossStakeEvent(ctx sdk.Context, keeper Keeper, from string, to []types.CrossReceiver, symbol string,
	eventType string, relayerFee int64) {
	if keeper.PbsbServer != nil {
		txHash := ctx.Value(baseapp.TxHashKey)
		if txHashStr, ok := txHash.(string); ok {
			event := types.CrossTransferEvent{
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
