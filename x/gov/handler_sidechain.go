package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/events"
)

func handleMsgSideChainSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSideChainSubmitProposal) sdk.Result {
	ctx, err := keeper.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId)
	if err != nil {
		return ErrInvalidSideChainId(keeper.codespace, msg.SideChainId).Result()
	}

	result := handleMsgSubmitProposal(ctx, keeper, msg.MsgSubmitProposal)
	if result.IsOK() {
		result.Tags = result.Tags.AppendTag(events.SideChainID, []byte(msg.SideChainId))
	}
	return result
}

func handleMsgSideChainDeposit(ctx sdk.Context, keeper Keeper, msg MsgSideChainDeposit) sdk.Result {
	ctx, err := keeper.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId)
	if err != nil {
		return ErrInvalidSideChainId(keeper.codespace, msg.SideChainId).Result()
	}

	result := handleMsgDeposit(ctx, keeper, msg.MsgDeposit)
	if result.IsOK() {
		result.Tags = result.Tags.AppendTag(events.SideChainID, []byte(msg.SideChainId))
	}
	return result
}

func handleMsgSideChainVote(ctx sdk.Context, keeper Keeper, msg MsgSideChainVote) sdk.Result {
	ctx, err := keeper.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId)
	if err != nil {
		return ErrInvalidSideChainId(keeper.codespace, msg.SideChainId).Result()
	}
	result := handleMsgVote(ctx, keeper, msg.MsgVote)
	if result.IsOK() {
		result.Tags = result.Tags.AppendTag(events.SideChainID, []byte(msg.SideChainId))
	}
	return result
}
