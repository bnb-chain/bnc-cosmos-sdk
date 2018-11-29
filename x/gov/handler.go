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

	proposal := keeper.NewTextProposal(ctx, msg.Title, msg.Description, msg.ProposalType)

	err, votingStarted := keeper.AddDeposit(ctx, proposal.GetProposalID(), msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(proposal.GetProposalID())

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
	isValidator := false
	keeper.vs.IterateValidatorsBonded(ctx, func(index int64, validator sdk.Validator) (stop bool) {
		if sdk.ValAddress(msg.Voter).Equals(validator.GetOperator()) {
			isValidator = true
			return true
		}
		return false
	})

	if !isValidator {
		return sdk.ErrUnauthorized("Non validator").Result()
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
func EndBlocker(ctx sdk.Context, keeper Keeper) (resTags sdk.Tags) {

	logger := ctx.Logger().With("module", "x/gov")

	resTags = sdk.NewTags()

	// Delete proposals that haven't met minDeposit
	for shouldPopInactiveProposalQueue(ctx, keeper) {
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
	for shouldPopActiveProposalQueue(ctx, keeper) {
		activeProposal := keeper.ActiveProposalQueuePop(ctx)

		proposalStartTime := activeProposal.GetVotingStartTime()
		votingPeriod := keeper.GetVotingProcedure(ctx).VotingPeriod
		if ctx.BlockHeader().Time.Before(proposalStartTime.Add(votingPeriod)) {
			continue
		}

		passes, tallyResults := tally(ctx, keeper, activeProposal)
		proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(activeProposal.GetProposalID())
		var action []byte
		if passes {
			activeProposal.SetStatus(StatusPassed)
			action = tags.ActionProposalPassed

			// refund deposits
			keeper.RefundDeposits(ctx, activeProposal.GetProposalID())
		} else {
			activeProposal.SetStatus(StatusRejected)
			action = tags.ActionProposalRejected

			// distribute deposits to proposer
			keeper.DistributeDeposits(ctx, activeProposal.GetProposalID())
		}

		activeProposal.SetTallyResult(tallyResults)
		keeper.SetProposal(ctx, activeProposal)

		logger.Info(fmt.Sprintf("proposal %d (%s) tallied; passed: %v",
			activeProposal.GetProposalID(), activeProposal.GetTitle(), passes))

		resTags.AppendTag(tags.Action, action)
		resTags.AppendTag(tags.ProposalID, proposalIDBytes)
	}

	return resTags
}
func shouldPopInactiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
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

func shouldPopActiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	votingProcedure := keeper.GetVotingProcedure(ctx)
	peekProposal := keeper.ActiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if !ctx.BlockHeader().Time.Before(peekProposal.GetVotingStartTime().Add(votingProcedure.VotingPeriod)) {
		return true
	}
	return false
}
