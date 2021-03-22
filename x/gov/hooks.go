package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// event hooks for governance
type GovHooks interface {
	OnProposalSubmitted(ctx sdk.Context, proposal Proposal) error // Must be called when a proposal submitted
}

type ExtGovHooks interface {
	GovHooks
	OnProposalPassed(ctx sdk.Context, proposal Proposal) error
}

func (keeper Keeper) OnProposalSubmitted(ctx sdk.Context, proposal Proposal) error {
	hs := keeper.hooks[proposal.GetProposalType()]
	for _, hooks := range hs {
		err := hooks.OnProposalSubmitted(ctx, proposal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (keeper Keeper) OnProposalPassed(ctx sdk.Context, proposal Proposal) error {
	hs := keeper.hooks[proposal.GetProposalType()]
	for _, hooks := range hs {
		switch hook := hooks.(type) {
		case ExtGovHooks:
			err := hook.OnProposalPassed(ctx, proposal)
			if err != nil {
				return err
			}
		default: // do nothing
		}
	}
	return nil
}
