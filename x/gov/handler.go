package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// Handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDeposit:
			return handleMsgDeposit(ctx, keeper, msg)
		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)
		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized gov msg type"
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {

	proposal := keeper.NewTextProposal(ctx, msg.Title, msg.Description, msg.ProposalType, msg.VotingPeriod)

	hooksErr := keeper.OnProposalSubmitted(ctx, proposal)
	if hooksErr != nil {
		return ErrInvalidProposal(keeper.codespace, hooksErr.Error()).Result()
	}

	proposalID := proposal.GetProposalID()
	proposalIDBytes := []byte(fmt.Sprintf("%d", proposalID))

	err, votingStarted := keeper.AddDeposit(ctx, proposal.GetProposalID(), msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		tags.Action, tags.ActionSubmitProposal,
		tags.Proposer, []byte(msg.Proposer.String()),
		tags.ProposalID, proposalIDBytes,
	)

	if votingStarted {
		resTags.AppendTag(tags.VotingPeriodStart, proposalIDBytes)
	}

	return sdk.Result{
		Data: proposalIDBytes,
		Tags: resTags,
	}
}

func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {

	err, votingStarted := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(msg.ProposalID)

	// TODO: Add tag for if voting period started
	resTags := sdk.NewTags(
		tags.Action, tags.ActionDeposit,
		tags.Depositer, []byte(msg.Depositer.String()),
		tags.ProposalID, proposalIDBytes,
	)

	if votingStarted {
		resTags.AppendTag(tags.VotingPeriodStart, proposalIDBytes)
	}

	return sdk.Result{
		Tags: resTags,
	}
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {
	validator := keeper.vs.Validator(ctx, sdk.ValAddress(msg.Voter))

	if validator == nil {
		return sdk.ErrUnauthorized("Vote is not from a validator operator").Result()
	}

	if validator.GetPower().IsZero() {
		return sdk.ErrUnauthorized("Validator is not bonded").Result()
	}

	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(msg.ProposalID)

	resTags := sdk.NewTags(
		tags.Action, tags.ActionVote,
		tags.Voter, []byte(msg.Voter.String()),
		tags.ProposalID, proposalIDBytes,
	)
	return sdk.Result{
		Tags: resTags,
	}
}

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) (resTags sdk.Tags, passedProposals, failedProposals []int64) {

	logger := ctx.Logger().With("module", "x/gov")

	resTags = sdk.NewTags()
	passedProposals = make([]int64, 0)
	failedProposals = make([]int64, 0)

	// Delete proposals that haven't met minDeposit
	for ShouldPopInactiveProposalQueue(ctx, keeper) {
		inactiveProposal := keeper.InactiveProposalQueuePop(ctx)
		if inactiveProposal.GetStatus() != StatusDepositPeriod {
			continue
		}

		proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(inactiveProposal.GetProposalID())

		// distribute deposits to proposer
		keeper.DistributeDeposits(ctx, inactiveProposal.GetProposalID())

		keeper.DeleteProposal(ctx, inactiveProposal)
		resTags.AppendTag(tags.Action, tags.ActionProposalDropped)
		resTags.AppendTag(tags.ProposalID, proposalIDBytes)

		logger.Info(
			fmt.Sprintf("proposal %d (%s) didn't meet minimum deposit of %v (had only %v); distribute to validator",
				inactiveProposal.GetProposalID(),
				inactiveProposal.GetTitle(),
				keeper.GetDepositProcedure(ctx).MinDeposit,
				inactiveProposal.GetTotalDeposit(),
			),
		)
	}

	// Check if earliest Active Proposal ended voting period yet
	for ShouldPopActiveProposalQueue(ctx, keeper) {
		activeProposal := keeper.ActiveProposalQueuePop(ctx)

		proposalStartTime := activeProposal.GetVotingStartTime()
		votingPeriod := activeProposal.GetVotingPeriod()
		if ctx.BlockHeader().Time.Before(proposalStartTime.Add(votingPeriod)) {
			continue
		}

		passes, refundDeposits, tallyResults := Tally(ctx, keeper, activeProposal)
		proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(activeProposal.GetProposalID())
		var action []byte
		if passes {
			activeProposal.SetStatus(StatusPassed)
			action = tags.ActionProposalPassed

			// refund deposits
			keeper.RefundDeposits(ctx, activeProposal.GetProposalID())
			passedProposals = append(passedProposals, activeProposal.GetProposalID())
		} else {
			activeProposal.SetStatus(StatusRejected)
			action = tags.ActionProposalRejected

			// if votes reached quorum and not all votes are abstain, distribute deposits to validator, else refund deposits
			if refundDeposits {
				keeper.RefundDeposits(ctx, activeProposal.GetProposalID())
			} else {
				keeper.DistributeDeposits(ctx, activeProposal.GetProposalID())
			}
			failedProposals = append(failedProposals, activeProposal.GetProposalID())
		}

		activeProposal.SetTallyResult(tallyResults)
		keeper.SetProposal(ctx, activeProposal)

		logger.Info(fmt.Sprintf("proposal %d (%s) tallied; passed: %v",
			activeProposal.GetProposalID(), activeProposal.GetTitle(), passes))

		resTags.AppendTag(tags.Action, action)
		resTags.AppendTag(tags.ProposalID, proposalIDBytes)
	}

	return
}

func ShouldPopInactiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	depositProcedure := keeper.GetDepositProcedure(ctx)
	peekProposal := keeper.InactiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if peekProposal.GetStatus() != StatusDepositPeriod {
		return true
	} else if !ctx.BlockHeader().Time.Before(peekProposal.GetSubmitTime().Add(depositProcedure.MaxDepositPeriod)) {
		return true
	}
	return false
}

func ShouldPopActiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	peekProposal := keeper.ActiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if !ctx.BlockHeader().Time.Before(peekProposal.GetVotingStartTime().Add(peekProposal.GetVotingPeriod())) {
		return true
	}
	return false
}
