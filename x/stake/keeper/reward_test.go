package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestAllocateReward(t *testing.T) {

	simDels := make([]Sharer, 11)
	// set 5% shares for the previous 10 delegator,
	// set 50% shares for the last delegator
	for i := 0; i < 11; i++ {
		delAddr := CreateTestAddr()
		shares := sdk.NewDecWithoutFra(5)
		if i == 10 {
			shares = sdk.NewDecWithoutFra(50)
		}
		simDel := Sharer{AccAddr: delAddr, Shares: shares}
		simDels[i] = simDel
	}
	totalShares := int64(100 * pow(8))

	commission := sdk.NewDec(10)

	/**
	 * case1:
	 *  commission: 10
	 *  remain: 5
	 *  delegator1-10: 5%, delegator11: 50%
	 * expected:
	 *  shouldCarry: delegator1-10, $1 for each(take all in 'remain')
	 *  shouldNotCarry: delegator11, $5 for delegator11
	 *  remainAfter: 0
	 */
	remainInt := int64(5)

	shouldCarry, shouldNotCarry, remainIntAfter := allocate(simDels, commission, sdk.NewDec(totalShares), remainInt)
	require.Len(t, shouldCarry, 10)
	require.Len(t, shouldNotCarry, 1)
	require.EqualValues(t, 0, remainIntAfter)
	for _, sc := range shouldCarry {
		require.EqualValues(t, 1, sc.Reward)
	}
	require.EqualValues(t, 5, shouldNotCarry[0].Reward)

	/**
	 * case2:
	 *  commission: 10
	 *  remain: 10
	 *  delegator1-10: 5%, delegator11: 50%
	 * expected:
	 *  shouldCarry: delegator1-10, $1 for each(take 5 from 'remain')
	 *  shouldNotCarry: delegator11, $5 for delegator11
	 *  remainAfter: 5
	 */
	remainInt = int64(10)
	shouldCarry, shouldNotCarry, remainIntAfter = allocate(simDels, commission, sdk.NewDec(totalShares), remainInt)
	require.Len(t, shouldCarry, 10)
	require.Len(t, shouldNotCarry, 1)
	require.EqualValues(t, 5, remainIntAfter)
	for _, sc := range shouldCarry {
		require.EqualValues(t, 1, sc.Reward)
	}
	require.EqualValues(t, 5, shouldNotCarry[0].Reward)

	/**
	 * case3:
	 *  commission: 10
	 *  remain: 3
	 *  delegator1-10: 5%, delegator11: 50%
	 * expected:
	 *  shouldCarry: delegator1-10, $1 for delegator1-8, $0 for delegator9-10 (take $3 from remain)
	 *  shouldNotCarry: delegator11, $5 for delegator11
	 *  remainAfter: 0
	 *
	 */
	remainInt = int64(3)
	shouldCarry, shouldNotCarry, remainIntAfter = allocate(simDels, commission, sdk.NewDec(totalShares), remainInt)
	require.Len(t, shouldCarry, 10)
	require.Len(t, shouldNotCarry, 1)
	require.EqualValues(t, 0, remainIntAfter)
	for i, sc := range shouldCarry {
		if i == 8 || i == 9 {
			require.EqualValues(t, 0, sc.Reward)
		} else {
			require.EqualValues(t, 1, sc.Reward)
		}
	}
	require.EqualValues(t, 5, shouldNotCarry[0].Reward)

	/**
	 * case4:
	 *  commission: 21
	 *  remain: 10
	 *  delegator1-10: 5%, delegator11: 50%
	 * expected:
	 *  shouldCarry: delegator11, $11 for delegator11
	 *  shouldNotCarry: delegator1-11, $1 for each one
	 *  remainAfter: 10
	 *
	 */
	commission = sdk.NewDec(21)
	remainInt = int64(10)
	shouldCarry, shouldNotCarry, remainIntAfter = allocate(simDels, commission, sdk.NewDec(totalShares), remainInt)
	require.Len(t, shouldCarry, 1)
	require.Len(t, shouldNotCarry, 10)
	require.EqualValues(t, 10, remainIntAfter)
	for _, sc := range shouldCarry {
		require.EqualValues(t, 11, sc.Reward)
	}
	for _, sc := range shouldNotCarry {
		require.EqualValues(t, 1, sc.Reward)
	}

}

func TestDiv(t *testing.T) {
	x := int64(1)
	y := int64(3)
	afterRoundDown, extraDecimalValue := Div(x, y, 1)
	require.EqualValues(t, 33333333, afterRoundDown)
	require.EqualValues(t, 3, extraDecimalValue)

	y = int64(300000000)
	afterRoundDown, extraDecimalValue = Div(x, y, 1)
	require.EqualValues(t, 0, afterRoundDown)
	require.EqualValues(t, 3, extraDecimalValue)

	x = int64(20000000000000000)
	y = int64(600000000)
	afterRoundDown, extraDecimalValue = Div(x, y, 1)
	require.EqualValues(t, 3333333333333333, afterRoundDown)
	require.EqualValues(t, 3, extraDecimalValue)
}

func pow(n int) int {
	return Pow(10, n)
}
