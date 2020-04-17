package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, keeper Keeper) {
	chainIDList := keeper.cfg.getChainIDList()
	channelList := keeper.cfg.getChannelIDList()

	var attributes []sdk.Attribute
	for _, destChainID := range chainIDList {
		for _, channelID := range channelList {

			lastHeightSequence := keeper.getLastHeightSequence(ctx, destChainID, channelID)
			curSequence := keeper.getSequence(ctx, destChainID, channelID)

			destChainName := keeper.cfg.destChainIDToName[destChainID]

			for sequence := lastHeightSequence; sequence < curSequence; sequence++ {
				attributes = append(attributes,
					sdk.NewAttribute(ibcPackageInfoAttributeKey, buildIBCPackageAttributeValue(destChainName, destChainID, channelID, sequence)))
			}
			//update last height sequence
			if lastHeightSequence != curSequence{
				keeper.setLastHeightSequence(ctx, destChainID, channelID, curSequence)
			}
		}
	}
	if len(attributes) > 0 {
		event := sdk.NewEvent(ibcEventType, attributes...)
		ctx.EventManager().EmitEvent(event)
	}
}
