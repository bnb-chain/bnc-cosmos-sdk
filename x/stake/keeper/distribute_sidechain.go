package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"strings"
)

const numberOfDecimalPlace = 8
const threshold = 5

func (k Keeper) Distribute(ctx sdk.Context, isCurrentBreatheHeightMarked bool) {

	height, found := getTargetValidatorsStoreHeight(ctx, k, isCurrentBreatheHeightMarked)
	if !found { // no data stored at expected breathe block height
		return
	}
	validators, found := k.GetValidatorsByHeight(ctx, height)
	if !found { // do nothing, if there is no validators to be rewarded.
		return
	}

	bondDenom := k.BondDenom(ctx)

	for _, validator := range validators {

		distAccCoins := k.bankKeeper.GetCoins(ctx, validator.DistributionAddr)

		totalReward := getTotalRewardThenClear(distAccCoins, bondDenom)

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
		if len(shouldCarry) > 0 {
			for _, eachReward := range shouldCarry {
				if _, _, err := k.bankKeeper.AddCoins(ctx, eachReward.AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, eachReward.Reward)}); err != nil {
					panic(err)
				}
			}
		}
		if len(shouldNotCarry) > 0 {
			for _, eachReward := range shouldNotCarry {
				if _, _, err := k.bankKeeper.AddCoins(ctx, eachReward.AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, eachReward.Reward)}); err != nil {
					panic(err)
				}
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

//func allocateReward(delegations []types.SimplifiedDelegation, commission sdk.Dec, totalShares int64, remainInt int64) ([]*delReward, []*delReward, int64) {
//	shouldCarry := make([]*delReward, 0)
//	shouldNotCarry := make([]*delReward, 0)
//	var minToDistribute int64
//	for _, del := range delegations {
//
//		afterRoundDown, firstDecimalValue := Div(del.Shares.Mul(commission).RawInt(), totalShares, 1)
//
//		if firstDecimalValue < threshold {
//			shouldNotCarry = append(shouldNotCarry, &delReward{del.DelegatorAddr, afterRoundDown})
//		} else {
//			shouldCarry = append(shouldCarry, &delReward{del.DelegatorAddr, afterRoundDown})
//		}
//		minToDistribute += afterRoundDown
//	}
//	leftCommission := commission.RawInt() - minToDistribute
//	if leftCommission > 0 {
//		for _, delR := range shouldCarry {
//			if leftCommission == 0 {
//				if remainInt == 0 {
//					break
//				}
//				delR.reward++
//				remainInt--
//			} else {
//				delR.reward++
//				leftCommission--
//			}
//		}
//		if leftCommission > 0 {
//			for _, delR := range shouldNotCarry {
//				if leftCommission == 0 {
//					break
//				} else {
//					delR.reward++
//					leftCommission--
//				}
//			}
//		}
//	}
//	return shouldCarry, shouldNotCarry, remainInt
//}

func getTotalRewardThenClear(distAccCoins sdk.Coins, bondDenom string) int64 {
	var totalReward int64
	for i := 0; i < distAccCoins.Len(); i++ {
		if strings.Compare(distAccCoins[i].Denom, bondDenom) == 0 {
			totalReward = distAccCoins[i].Amount
			distAccCoins[i].Amount = 0
			break
		}
	}
	return totalReward
}

/**
 * If the current day's breathe block height is marked before this query, then find the third height from the bottom.
 * Otherwise, find the second-to-last one.
 */
func getTargetValidatorsStoreHeight(ctx sdk.Context, k Keeper, isCurrentBreatheHeightMarked bool) (height int64, found bool) {
	index := 2
	if isCurrentBreatheHeightMarked {
		index = 3
	}
	return k.GetBreatheBlockHeight(ctx, index)
}
