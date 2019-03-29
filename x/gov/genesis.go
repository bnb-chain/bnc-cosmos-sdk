package gov

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultDepositDenom = "steak"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	StartingProposalID int64             `json:"starting_proposalID"`
	DepositProcedure   DepositProcedure  `json:"deposit_period"`
	TallyingProcedure  TallyingProcedure `json:"tallying_procedure"`
}

func NewGenesisState(startingProposalID int64, dp DepositProcedure, tp TallyingProcedure) GenesisState {
	return GenesisState{
		StartingProposalID: startingProposalID,
		DepositProcedure:   dp,
		TallyingProcedure:  tp,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		StartingProposalID: 1,
		DepositProcedure: DepositProcedure{
			MinDeposit:       sdk.Coins{sdk.NewCoin(DefaultDepositDenom, 2000e8)},
			MaxDepositPeriod: time.Duration(2*24) * time.Hour, // 2 days
		},
		TallyingProcedure: TallyingProcedure{
			Quorum:    sdk.NewDecWithPrec(5, 1),
			Threshold: sdk.NewDecWithPrec(5, 1),
			Veto:      sdk.NewDecWithPrec(334, 3),
		},
	}
}

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	err := k.SetInitialProposalID(ctx, data.StartingProposalID)
	if err != nil {
		// TODO: Handle this with #870
		panic(err)
	}
	k.setDepositProcedure(ctx, data.DepositProcedure)
	k.setTallyingProcedure(ctx, data.TallyingProcedure)
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	startingProposalID, _ := k.getNewProposalID(ctx)
	depositProcedure := k.GetDepositProcedure(ctx)
	tallyingProcedure := k.GetTallyingProcedure(ctx)

	return GenesisState{
		StartingProposalID: startingProposalID,
		DepositProcedure:   depositProcedure,
		TallyingProcedure:  tallyingProcedure,
	}
}
