package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Sharer struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
}

type EachReward struct {
	AccAddr sdk.AccAddress
	Reward  int64
}

func allocate(sharers []Sharer, totalRewards sdk.Dec, totalShares sdk.Dec, extra int64) (shouldCarry []EachReward, shouldNotCarry []EachReward, newExtra int64) {
	var minToDistribute int64
	for _, sharer := range sharers {

		afterRoundDown, firstDecimalValue := sdk.MulQuoDecWithExtraDecimal(sharer.Shares,totalRewards,totalShares,1)

		if firstDecimalValue < threshold {
			shouldNotCarry = append(shouldNotCarry, EachReward{sharer.AccAddr, afterRoundDown})
		} else {
			shouldCarry = append(shouldCarry, EachReward{sharer.AccAddr, afterRoundDown})
		}
		minToDistribute += afterRoundDown
	}
	remainingRewards := totalRewards.RawInt() - minToDistribute
	if remainingRewards > 0 {
		for i := range shouldCarry {
			if remainingRewards == 0 {
				if extra == 0 {
					break
				}
				shouldCarry[i].Reward++
				extra--
			} else {
				shouldCarry[i].Reward++
				remainingRewards--
			}
		}
		if remainingRewards > 0 {
			for i := range shouldNotCarry {
				if remainingRewards == 0 {
					break
				} else {
					shouldNotCarry[i].Reward++
					remainingRewards--
				}
			}
		}
	}
	return shouldCarry, shouldNotCarry, extra
}
