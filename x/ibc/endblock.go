package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, keeper Keeper) {
	if len(keeper.packageCollector.collectedPackages) == 0 {
		return
	}
	var (
		attributes []sdk.Attribute
		events     sdk.Events
	)
	for _, ibcPackageRecord := range keeper.packageCollector.collectedPackages {
		attributes = append(attributes,
			sdk.NewAttribute(ibcPackageInfoAttributeKey,
				buildIBCPackageAttributeValue(ibcPackageRecord.destChainID, ibcPackageRecord.channelID, ibcPackageRecord.sequence)))
	}

	keeper.packageCollector.collectedPackages = keeper.packageCollector.collectedPackages[:0]
	event := sdk.NewEvent(ibcEventType, attributes...)
	events.AppendEvent(event)
	if sdk.IsUpgrade(sdk.BCFusionThirdHardFork) {
		events = events.AppendEvents(closeSideChainChannels(ctx, keeper))
	}
	ctx.EventManager().EmitEvents(events)
}

func closeSideChainChannels(ctx sdk.Context, k Keeper) sdk.Events {
	var events sdk.Events
	sideChainId := k.sideKeeper.BscSideChainId(ctx)
	// disable side chain channels
	id := k.sideKeeper.Config().DestChainNameToID(sideChainId)
	govChannelId := sdk.ChannelID(0x09)
	for _, channelId := range k.sideKeeper.Config().ChannelIDs() {
		if channelId == govChannelId {
			// skip gov channel
			continue
		}
		permissions := k.sideKeeper.GetChannelSendPermissions(ctx, id)
		if permissions[channelId] == sdk.ChannelForbidden {
			// skip forbidden channel
			continue
		}

		events = events.AppendEvents(saveChannelSetting(ctx, k, id, channelId, sdk.ChannelForbidden))
	}

	// disable side chain gov channel
	events = events.AppendEvents(saveChannelSetting(ctx, k, id, govChannelId, sdk.ChannelForbidden))
	return events
}

func saveChannelSetting(ctx sdk.Context, k Keeper,
	destChainID sdk.ChainID, channelID sdk.ChannelID, permission sdk.ChannelPermission) sdk.Events {
	var events sdk.Events
	_, err := k.sideKeeper.SaveChannelSettingChangeToIbc(ctx, destChainID, channelID, sdk.ChannelForbidden)
	if err != nil {
		ctx.Logger().Error("closeSideChainChannels", "err", err.Error())
		events.AppendEvent(sdk.NewEvent("failed to closeSideChainChannels ",
			sdk.NewAttribute("sideChainId", fmt.Sprint(destChainID)),
			sdk.NewAttribute("channelId", fmt.Sprint(channelID)),
			sdk.NewAttribute("error", err.Error()),
		))
		return events
	}
	events.AppendEvent(sdk.NewEvent("closeSideChainChannels",
		sdk.NewAttribute("sideChainId", fmt.Sprint(destChainID)),
		sdk.NewAttribute("channelId", fmt.Sprint(channelID)),
	))
	// close bc side chain channel
	k.sideKeeper.SetChannelSendPermission(ctx, destChainID, channelID, sdk.ChannelForbidden)

	return events
}
