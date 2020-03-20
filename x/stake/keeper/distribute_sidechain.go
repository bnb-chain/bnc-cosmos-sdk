package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	numberOfDecimalPlace = 8
	threshold            = 5
)

func (k Keeper) Distribute(ctx sdk.Context, height int64) {

	validators, found := k.GetValidatorsByHeight(ctx, height)
	if !found { // do nothing, if there is no validators to be rewarded.
		return
	}

	bondDenom := k.BondDenom(ctx)

	for _, validator := range validators {

		distAccCoins := k.bankKeeper.GetCoins(ctx, validator.DistributionAddr)

		totalReward := distAccCoins.AmountOf(bondDenom)

		if totalReward == 0 { // there is no reward for this validator
			continue
		}

		delegations, found := k.GetSimplifiedDelegations(ctx, height, validator.OperatorAddr)
		if !found {
			panic(fmt.Sprintf("no delegations found with height=%d, validator=%s", height, validator.OperatorAddr))
		}

		totalRewardDec := sdk.NewDec(totalReward)
		commission := totalRewardDec.Mul(validator.Commission.Rate)
		remainInt := totalRewardDec.Sub(commission).RawInt()

		// remove all balance of bondDenom from Distribution account
		distAccCoins = distAccCoins.Minus(sdk.Coins{sdk.NewCoin(bondDenom, totalReward)})
		if err := k.bankKeeper.SetCoins(ctx, validator.DistributionAddr, distAccCoins); err != nil {
			panic(err)
		}

		//shouldCarry, shouldNotCarry, remainInt := allocateReward(delegations, commission, validator.DelegatorShares.RawInt(), remainInt)
		shouldCarry, shouldNotCarry, remainInt := allocate(simDelsToSharers(delegations), commission, validator.DelegatorShares, remainInt)

		if remainInt > 0 { // assign rewards to self-delegator
			if _, _, err := k.bankKeeper.AddCoins(ctx, validator.GetFeeAddr(), sdk.Coins{sdk.NewCoin(bondDenom, remainInt)}); err != nil {
				panic(err)
			}
		}

		// assign rewards to delegator
		for _, eachReward := range shouldCarry {
			if _, _, err := k.bankKeeper.AddCoins(ctx, eachReward.AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, eachReward.Reward)}); err != nil {
				panic(err)
			}
		}

		for _, eachReward := range shouldNotCarry {
			if _, _, err := k.bankKeeper.AddCoins(ctx, eachReward.AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, eachReward.Reward)}); err != nil {
				panic(err)
			}
		}

	}

	removeValidatorsAndDelegationsAtHeight(height, k, ctx, validators)
}

func simDelsToSharers(simDels []types.SimplifiedDelegation) []Sharer {
	sharers := make([]Sharer, len(simDels))
	for i, del := range simDels {
		sharers[i] = Sharer{AccAddr: del.DelegatorAddr, Shares: del.Shares}
	}
	return sharers
}

func removeValidatorsAndDelegationsAtHeight(height int64, k Keeper, ctx sdk.Context, validators []types.Validator) {
	for _, validator := range validators {
		k.RemoveSimplifiedDelegations(ctx, height, validator.OperatorAddr)
	}
	k.RemoveValidatorsByHeight(ctx, height)
}
