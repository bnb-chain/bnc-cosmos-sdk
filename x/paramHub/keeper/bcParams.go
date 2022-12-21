package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/paramHub/types"
)

func (keeper *Keeper) getLastBCParamChanges(ctx sdk.Context) *types.BCChangeParams {
	var latestProposal *gov.Proposal
	lastProposalId := keeper.GetLastBCParamChangeProposalId(ctx)
	keeper.govKeeper.Iterate(ctx, nil, nil, gov.StatusPassed, lastProposalId.ProposalID, true, func(proposal gov.Proposal) bool {
		if proposal.GetProposalType() == gov.ProposalTypeParameterChange {
			latestProposal = &proposal
			return true
		}
		return false
	})

	if latestProposal != nil {
		var changeParam types.BCChangeParams
		strProposal := (*latestProposal).GetDescription()
		err := keeper.cdc.UnmarshalJSON([]byte(strProposal), &changeParam)
		if err != nil {
			keeper.Logger(ctx).Error("Get broken data when unmarshal BCParamsChange msg, will skip.", "proposalId", (*latestProposal).GetProposalID(), "err", err)
			return nil
		}
		// SetLastBCParamChangeProposalId first. If invalid, the proposal before it will not been processed too.
		keeper.SetLastBCParamChangeProposalId(ctx, types.LastProposalID{ProposalID: (*latestProposal).GetProposalID()})
		if err := changeParam.Check(); err != nil {
			keeper.Logger(ctx).Error("The BCParamsChange proposal is invalid, will skip.", "proposalId", (*latestProposal).GetProposalID(), "param", changeParam, "err", err)
			return nil
		}
		return &changeParam
	}
	return nil
}

func (keeper *Keeper) GetLastBCParamChangeProposalId(ctx sdk.Context) types.LastProposalID {
	var id types.LastProposalID
	keeper.paramSpace.GetIfExists(ctx, ParamStoreKeyBCLastParamsChangeProposalID, &id)
	return id
}

func (keeper *Keeper) SetLastBCParamChangeProposalId(ctx sdk.Context, id types.LastProposalID) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyBCLastParamsChangeProposalID, &id)
	return
}

func (keeper *Keeper) GetBCParams(ctx sdk.Context) ([]types.BCParam, sdk.Error) {
	params := make([]types.BCParam, 0)
	for _, subSpace := range keeper.GetSubscriberBCParamSpace() {
		param := subSpace.Proto()
		subSpace.ParamSpace.GetParamSet(ctx, param)
		params = append(params, param)
	}
	return params, nil
}
