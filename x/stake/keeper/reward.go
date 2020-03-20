package keeper

import (
	"math"
	"math/big"

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

func allocate(sharers []Sharer, totalRewards sdk.Dec, totalShares sdk.Dec, extra int64) (shouldCarry []*EachReward, shouldNotCarry []*EachReward, newExtra int64) {
	var minToDistribute int64
	for _, sharer := range sharers {

		afterRoundDown, firstDecimalValue := Div(sharer.Shares.Mul(totalRewards).RawInt(), totalShares.RawInt(), 1)

		if firstDecimalValue < threshold {
			shouldNotCarry = append(shouldNotCarry, &EachReward{sharer.AccAddr, afterRoundDown})
		} else {
			shouldCarry = append(shouldCarry, &EachReward{sharer.AccAddr, afterRoundDown})
		}
		minToDistribute += afterRoundDown
	}
	remainingRewards := totalRewards.RawInt() - minToDistribute
	if remainingRewards > 0 {
		for _, eachReward := range shouldCarry {
			if remainingRewards == 0 {
				if extra == 0 {
					break
				}
				eachReward.Reward++
				extra--
			} else {
				eachReward.Reward++
				remainingRewards--
			}
		}
		if remainingRewards > 0 {
			for _, eachReward := range shouldNotCarry {
				if remainingRewards == 0 {
					break
				} else {
					eachReward.Reward++
					remainingRewards--
				}
			}
		}
	}
	return shouldCarry, shouldNotCarry, extra

}

func int64Div(x, y int64, extraDecimalPlace int) (afterRoundDown int64, extraDecimalValue int) {
	resultOfAddDecimalPlace := (x * int64(Pow(10, numberOfDecimalPlace+extraDecimalPlace))) / y
	dived := int64(Pow(10, int(extraDecimalPlace)))
	afterRoundDown = resultOfAddDecimalPlace / dived
	extraDecimalValue = int(resultOfAddDecimalPlace % dived)
	return afterRoundDown, extraDecimalValue
}

func Div(x, y int64, extraDecimalPlace int) (afterRoundDown int64, extraDecimalValue int) {

	minAllow := math.MaxInt64 / int64(Pow(10, numberOfDecimalPlace+extraDecimalPlace))
	if x <= minAllow {
		return int64Div(x, y, extraDecimalPlace)
	}

	z := &big.Int{}
	z.Mul(big.NewInt(x), big.NewInt(int64(Pow(10, numberOfDecimalPlace+extraDecimalPlace)))).Div(z, big.NewInt(y))

	dived := big.NewInt(int64(Pow(10, int(extraDecimalPlace))))

	expectedDecimalValueBig := &big.Int{}
	afterRoundDownBig, expectedDecimalValueBig := z.QuoRem(z, dived, expectedDecimalValueBig)
	afterRoundDown = afterRoundDownBig.Int64()
	extraDecimalValue = int(expectedDecimalValueBig.Int64())
	return afterRoundDown, extraDecimalValue
}

func Pow(x, n int) int {
	ret := 1
	for n != 0 {
		if n%2 != 0 {
			ret = ret * x
		}
		n /= 2
		x = x * x
	}
	return ret
}
