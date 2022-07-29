package cross_stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func PublishCrossStakeEvent(keeper Keeper, delegator sdk.AccAddress, valSrc sdk.ValAddress, valDst sdk.ValAddress,
	eventType string, relayFee int64) {
	chainId := keeper.DestChainName
	if keeper.PbsbServer != nil {
		event := types.CrossStakeEvent{
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
