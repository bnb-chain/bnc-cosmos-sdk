package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// event hooks for governance
type GovHooks interface {
	OnProposalSubmitted(ctx sdk.Context, proposal Proposal) error // Must be called when a proposal submitted
}

func (keeper Keeper) OnProposalSubmitted(ctx sdk.Context, proposal Proposal) error {
	if keeper.hooks != nil {
		return keeper.hooks.OnProposalSubmitted(ctx, proposal)
	}
	return nil
}
