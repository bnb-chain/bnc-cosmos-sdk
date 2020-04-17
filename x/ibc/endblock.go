package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, keeper Keeper) {
	if keeper.packageCollector == nil {
		return
	}
	var attributes []sdk.Attribute
	for _, ibcPackageRecord := range keeper.packageCollector {
		attributes = append(attributes,
			sdk.NewAttribute(ibcPackageInfoAttributeKey,
				buildIBCPackageAttributeValue(ibcPackageRecord.destChainName, ibcPackageRecord.destChainID, ibcPackageRecord.channelID, ibcPackageRecord.sequence)))
	}
	keeper.packageCollector = nil
	event := sdk.NewEvent(ibcEventType, attributes...)
	ctx.EventManager().EmitEvent(event)
}
