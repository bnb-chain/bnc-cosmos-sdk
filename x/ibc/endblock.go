package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
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
	if sdk.IsUpgrade(sdk.FinalSunsetFork) && !keeper.sideKeeper.IsBSCAllChannelClosed(ctx) {
		events = events.AppendEvents(closeSideChainChannels(ctx, keeper))
	}
	ctx.EventManager().EmitEvents(events)
}

func closeSideChainChannels(ctx sdk.Context, k Keeper) sdk.Events {
	var events sdk.Events
	sideChainId := k.sideKeeper.BscSideChainId(ctx)
	// disable side chain channels
	id := k.sideKeeper.Config().DestChainNameToID(sideChainId)
	govChannelId := sdk.ChannelID(gov.ProposalTypeManageChanPermission)
	permissions := k.sideKeeper.GetChannelSendPermissions(ctx, id)
	for _, channelId := range k.sideKeeper.Config().ChannelIDs() {
		if channelId == govChannelId {
			// skip gov channel
			continue
		}
		if permissions[channelId] == sdk.ChannelForbidden {
			// skip forbidden channel
			continue
		}

		events = events.AppendEvents(closeChannelOnSideChanAndKeeper(ctx, k, id, channelId))
	}

	// disable side chain gov channel
	if permissions[govChannelId] == sdk.ChannelAllow {
		events = events.AppendEvents(closeChannelOnSideChanAndKeeper(ctx, k, id, govChannelId))
	}
	k.sideKeeper.SetBSCAllChannelClosed(ctx)
	return events
}

func closeChannelOnSideChanAndKeeper(ctx sdk.Context, k Keeper,
	destChainID sdk.ChainID, channelID sdk.ChannelID) sdk.Events {
	var events sdk.Events
	_, err := k.sideKeeper.SaveChannelSettingChangeToIbc(ctx, destChainID, channelID, sdk.ChannelForbidden)
	if err != nil {
		ctx.Logger().Error("failed to save ibc channel change", "err", err.Error())
		events.AppendEvent(sdk.NewEvent(EventTypeSaveIBCChannelSettingFailed,
			sdk.NewAttribute(AttributeKeySideChainId, fmt.Sprint(destChainID)),
			sdk.NewAttribute(AttributeKeyChannelId, fmt.Sprint(channelID)),
			sdk.NewAttribute(AttributeKeyError, err.Error()),
		))
		return events
	}
	events.AppendEvent(sdk.NewEvent(EventTypeSaveIBCChannelSettingSucceed,
		sdk.NewAttribute(AttributeKeySideChainId, fmt.Sprint(destChainID)),
		sdk.NewAttribute(AttributeKeyChannelId, fmt.Sprint(channelID)),
	))
	// close bc side chain channel
	k.sideKeeper.SetChannelSendPermission(ctx, destChainID, channelID, sdk.ChannelForbidden)
	return events
}
