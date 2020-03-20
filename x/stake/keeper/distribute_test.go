package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

func TestDistribute(t *testing.T) {
	ctx, am, k := CreateTestInput(t, false, 0)

	height := int64(1000)
	height2 := int64(2000)
	height3 := int64(3000)
	now := time.Now()
	k.SetBreatheBlockHeight(ctx,height,now.Add( - 2 * 24 * time.Hour))
	k.SetBreatheBlockHeight(ctx,height2,now.Add( - 24 * time.Hour))
	k.SetBreatheBlockHeight(ctx,height3,now)

	minDelShares := 1
	maxDelShares := 100000

	minDelNum := 10
	maxDelNum := 500

	minCollectedFee := 1
	maxCollectedFee := 10000

	validators := make([]types.Validator,21)
	delegators := make([][]sdk.AccAddress,21)
	rewards := make([]int64,21)
	rand.Seed(time.Now().UnixNano())
	for i:=0;i<21;i++ {
		valPubKey := PKs[i]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
		validator := types.NewValidator(valAddr, valPubKey, types.Description{})

		delNum := minDelNum + rand.Intn(maxDelNum - minDelNum + 1)
		var totalShares int64
		simDels := make([]types.SimplifiedDelegation,delNum)
		delsForVal := make([]sdk.AccAddress,0)
		for j:=0;j<delNum;j++ {
			delAddr := CreateTestAddr()
			if j == 0 {
				validator.FeeAddr = delAddr
			}
			shares := int64( (minDelShares + rand.Intn(maxDelShares - minDelShares + 1)) * 100000000 )
			totalShares += shares
			simDel := types.SimplifiedDelegation{
				DelegatorAddr: delAddr,
				Shares:        sdk.NewDec(shares),
			}
			simDels[j] = simDel
			delsForVal = append(delsForVal, delAddr)
		}
		delegators[i] = delsForVal
		k.SetSimplifiedDelegations(ctx,height,validator.OperatorAddr,simDels)

		validator.DelegatorShares = sdk.NewDec(totalShares)
		validator.Tokens = sdk.NewDec(totalShares)
		validator.DistributionAddr = Addrs[499-i]
		validator,setCommErr := validator.SetInitialCommission(types.Commission{Rate: sdk.NewDecWithPrec(60,2),MaxRate: sdk.NewDecWithPrec(90,2)})
		require.NoError(t,setCommErr)
		validators[i] = validator

		// simulate distribute account
		distrAcc := am.NewAccountWithAddress(ctx,validator.DistributionAddr)
		randCollectedFee := int64( (minCollectedFee + rand.Intn(maxCollectedFee - minCollectedFee + 1)) * 100000000 )
		err := distrAcc.SetCoins(sdk.Coins{sdk.NewCoin("BNB",randCollectedFee)})
		require.NoError(t,err)
		rewards[i] = randCollectedFee
		am.SetAccount(ctx,distrAcc)
	}
	k.SetValidatorsByHeight(ctx,height,validators)
	k.Distribute(ctx,true)

	for i,validator := range validators {
		_,found := k.GetSimplifiedDelegations(ctx,height,validator.OperatorAddr)
		require.False(t,found)

		distrAcc := am.GetAccount(ctx,validator.DistributionAddr)
		balanceOfBNB := distrAcc.GetCoins().AmountOf("BNB")
		require.Equal(t,int64(0),balanceOfBNB)

		var amountOfAllAccount int64
		for _, delAddr := range delegators[i] {
			delAcc := am.GetAccount(ctx,delAddr)
			amountOfAllAccount += delAcc.GetCoins().AmountOf("BNB")
		}
		require.Equal(t,rewards[i], amountOfAllAccount)
	}
	_,found := k.GetValidatorsByHeight(ctx,height)
	require.False(t,found)

}

func TestAllocateReward(t *testing.T) {

	simDels := make([]types.SimplifiedDelegation,11)
	// set 5% shares for the previous 10 delegator,
	// set 50% shares for the last delegator
	for i:=0;i<11;i++ {
		delAddr := CreateTestAddr()
		shares := sdk.NewDecWithoutFra(5)
		if i == 10 {
			shares = sdk.NewDecWithoutFra(50)
		}
		simDel := types.SimplifiedDelegation{DelegatorAddr: delAddr,Shares: shares}
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

	shouldCarry,shouldNotCarry,remainIntAfter := allocateReward(simDels,commission,totalShares,remainInt)
	require.Len(t,shouldCarry,10)
	require.Len(t,shouldNotCarry,1)
	require.EqualValues(t,0,remainIntAfter)
	for _,sc := range shouldCarry {
		require.EqualValues(t,1,sc.reward)
	}
	require.EqualValues(t,5,shouldNotCarry[0].reward)

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
	shouldCarry,shouldNotCarry,remainIntAfter = allocateReward(simDels,commission,totalShares,remainInt)
	require.Len(t,shouldCarry,10)
	require.Len(t,shouldNotCarry,1)
	require.EqualValues(t,5,remainIntAfter)
	for _,sc := range shouldCarry {
		require.EqualValues(t,1,sc.reward)
	}
	require.EqualValues(t,5,shouldNotCarry[0].reward)

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
	shouldCarry,shouldNotCarry,remainIntAfter = allocateReward(simDels,commission,totalShares,remainInt)
	require.Len(t,shouldCarry,10)
	require.Len(t,shouldNotCarry,1)
	require.EqualValues(t,0,remainIntAfter)
	for i,sc := range shouldCarry {
		if i == 8 || i == 9 {
			require.EqualValues(t,0,sc.reward)
		} else {
			require.EqualValues(t,1,sc.reward)
		}
	}
	require.EqualValues(t,5,shouldNotCarry[0].reward)

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
	shouldCarry,shouldNotCarry,remainIntAfter = allocateReward(simDels,commission,totalShares,remainInt)
	require.Len(t,shouldCarry,1)
	require.Len(t,shouldNotCarry,10)
	require.EqualValues(t,10,remainIntAfter)
	for _,sc := range shouldCarry {
		require.EqualValues(t,11,sc.reward)
	}
	for _,sc := range shouldNotCarry {
		require.EqualValues(t,1,sc.reward)
	}

}

func TestDiv(t *testing.T) {
	x := int64(1)
	y := int64(3)
	afterRoundDown,extraDecimalValue := Div(x,y,1)
	require.EqualValues(t,33333333,afterRoundDown)
	require.EqualValues(t,3,extraDecimalValue)

	y = int64(300000000)
	afterRoundDown,extraDecimalValue = Div(x,y,1)
	require.EqualValues(t,0,afterRoundDown)
	require.EqualValues(t,3,extraDecimalValue)

	x = int64(20000000000000000)
	y = int64(600000000)
	afterRoundDown,extraDecimalValue = Div(x,y,1)
	require.EqualValues(t,3333333333333333,afterRoundDown)
	require.EqualValues(t,3,extraDecimalValue)
}

func pow(n int) int {
	return Pow(10,n)
}
