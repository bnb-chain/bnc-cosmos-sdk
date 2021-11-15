package keeper

import (
	"encoding/binary"
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const threshold = 5

func allocate(sharers []types.Sharer, totalRewards sdk.Dec) (rewards []types.Reward) {
	var minToDistribute int64
	var shouldCarry []types.Reward
	var shouldNotCarry []types.Reward

	totalShares := sdk.ZeroDec()
	for _, sharer := range sharers {
		totalShares = totalShares.Add(sharer.Shares)
	}

	for _, sharer := range sharers {

		afterRoundDown, firstDecimalValue := mulQuoDecWithExtraDecimal(sharer.Shares, totalRewards, totalShares, 1)

		if firstDecimalValue < threshold {
			shouldNotCarry = append(shouldNotCarry, types.Reward{AccAddr: sharer.AccAddr, Shares: sharer.Shares, Amount: afterRoundDown})
		} else {
			shouldCarry = append(shouldCarry, types.Reward{AccAddr: sharer.AccAddr, Shares: sharer.Shares, Amount: afterRoundDown})
		}
		minToDistribute += afterRoundDown
	}
	remainingRewards := totalRewards.RawInt() - minToDistribute
	if remainingRewards > 0 {
		for i := range shouldCarry {
			if remainingRewards <= 0 {
				break
			} else {
				shouldCarry[i].Amount++
				remainingRewards--
			}
		}
		if remainingRewards > 0 {
			for i := range shouldNotCarry {
				if remainingRewards <= 0 {
					break
				} else {
					shouldNotCarry[i].Amount++
					remainingRewards--
				}
			}
		}
	}

	return append(shouldCarry, shouldNotCarry...)
}

// calculate a * b / c, getting the extra decimal digital as result of extraDecimalValue. For example:
// 0.00000003 * 2 / 0.00000004 = 0.000000015,
// assume that decimal place number of Dec is 8, and the extraDecimalPlace was given 1, then
// we take the 8th decimal place value '1' as afterRoundDown, and extra decimal value(9th) '5' as extraDecimalValue
func mulQuoDecWithExtraDecimal(a, b, c sdk.Dec, extraDecimalPlace int) (afterRoundDown int64, extraDecimalValue int) {
	extra := int64(math.Pow(10, float64(extraDecimalPlace)))
	product, ok := sdk.Mul64(a.RawInt(), b.RawInt())
	if !ok { // int64 exceed
		return mulQuoBigIntWithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
	} else {
		if product, ok = sdk.Mul64(product, extra); !ok {
			return mulQuoBigIntWithExtraDecimal(big.NewInt(a.RawInt()), big.NewInt(b.RawInt()), big.NewInt(c.RawInt()), big.NewInt(extra))
		}
		resultOfAddDecimalPlace := product / c.RawInt()
		afterRoundDown = resultOfAddDecimalPlace / extra
		extraDecimalValue = int(resultOfAddDecimalPlace % extra)
		return afterRoundDown, extraDecimalValue
	}
}

func mulQuoBigIntWithExtraDecimal(a, b, c, extra *big.Int) (afterRoundDown int64, extraDecimalValue int) {
	product := sdk.MulBigInt(sdk.MulBigInt(a, b), extra)
	result := sdk.QuoBigInt(product, c)

	expectedDecimalValueBig := &big.Int{}
	afterRoundDownBig, expectedDecimalValueBig := result.QuoRem(result, extra, expectedDecimalValueBig)
	afterRoundDown = afterRoundDownBig.Int64()
	extraDecimalValue = int(expectedDecimalValueBig.Int64())
	return afterRoundDown, extraDecimalValue
}

func (k Keeper) SetRewards(ctx sdk.Context, sideChainId string, batchNo int64, rewards []types.Reward) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalRewards(k.cdc, rewards)
	store.Set(GetSideChainBatchKey(sideChainId, batchNo), bz)
}

func GetSideChainBatchKey(sideChainId string, batchNo int64) []byte {
	bz1 := make([]byte, 8)
	copy(bz1, sideChainId)

	bz2 := []byte{'-'}

	bz3 := make([]byte, 8)
	binary.BigEndian.PutUint64(bz3, uint64(batchNo))

	bz := append(RewardKey, bz1...)
	bz = append(bz, bz2...)
	bz = append(bz, bz3...)

	return bz
}

func (k Keeper) GetRewards(ctx sdk.Context, sideChainId string, batchNo int64) (rewards []types.Reward) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(GetSideChainBatchKey(sideChainId, batchNo))
	rewards = types.MustUnmarshalRewards(k.cdc, value)
	return
}
