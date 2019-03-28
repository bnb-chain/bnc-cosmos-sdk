package gov

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Procedure around Deposits for governance
type DepositProcedure struct {
	MinDeposit       sdk.Coins     `json:"min_deposit"`        //  Minimum deposit for a proposal to enter voting period.
	MaxDepositPeriod time.Duration `json:"max_deposit_period"` //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
}

// Procedure around Tallying votes in governance
type TallyingProcedure struct {
	Quorum    sdk.Dec `json:"quorum"`    //  Minimum percentage of total stake needed to vote for a result to be considered valid
	Threshold sdk.Dec `json:"threshold"` //  Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
	Veto      sdk.Dec `json:"veto"`      //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
}
