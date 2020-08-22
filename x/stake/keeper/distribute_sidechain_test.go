package keeper

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/require"
)

func TestDistribute(t *testing.T) {
	ctx, am, k := CreateTestInput(t, false, 0)
	k.addrPool = new(sdk.Pool)
	bondDenom := k.BondDenom(ctx)

	height := int64(1000)
	height2 := int64(2000)
	height3 := int64(3000)

	minDelShares := 1
	maxDelShares := 100000

	minDelNum := 10
	maxDelNum := 500

	minCollectedFee := 1
	maxCollectedFee := 10000

	validators := make([]types.Validator, 21)
	delegators := make([][]sdk.AccAddress, 21)
	rewards := make([]int64, 21)
	rand.Seed(time.Now().UnixNano())
	var totalDelNum int
	for i := 0; i < 21; i++ {
		valPubKey := PKs[i]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
		validator := types.NewValidator(valAddr, valPubKey, types.Description{})

		delNum := minDelNum + rand.Intn(maxDelNum-minDelNum+1)
		var totalShares int64
		simDels := make([]types.SimplifiedDelegation, delNum)
		delsForVal := make([]sdk.AccAddress, 0)
		totalDelNum += delNum
		for j := 0; j < delNum; j++ {
			delAddr := CreateTestAddr()
			if j == 0 {
				validator.FeeAddr = delAddr
			}
			shares := int64((minDelShares + rand.Intn(maxDelShares-minDelShares+1)) * 100000000)
			totalShares += shares
			simDel := types.SimplifiedDelegation{
				DelegatorAddr: delAddr,
				Shares:        sdk.NewDec(shares),
			}
			simDels[j] = simDel
			delsForVal = append(delsForVal, delAddr)
		}
		delegators[i] = delsForVal
		k.SetSimplifiedDelegations(ctx, height, validator.OperatorAddr, simDels)

		validator.DelegatorShares = sdk.NewDec(totalShares)
		validator.Tokens = sdk.NewDec(totalShares)
		validator.DistributionAddr = Addrs[499-i]
		validator, setCommErr := validator.SetInitialCommission(types.Commission{Rate: sdk.NewDecWithPrec(40, 2), MaxRate: sdk.NewDecWithPrec(90, 2)})
		require.NoError(t, setCommErr)
		validators[i] = validator

		// simulate distribute account
		distrAcc := am.NewAccountWithAddress(ctx, validator.DistributionAddr)
		randCollectedFee := int64((minCollectedFee + rand.Intn(maxCollectedFee-minCollectedFee+1)) * 100000000)
		err := distrAcc.SetCoins(sdk.Coins{sdk.NewCoin(bondDenom, randCollectedFee)})
		require.NoError(t, err)
		rewards[i] = randCollectedFee
		am.SetAccount(ctx, distrAcc)
	}
	k.SetValidatorsByHeight(ctx, height, validators)
	k.SetValidatorsByHeight(ctx, height2, make([]types.Validator, 0))
	k.SetValidatorsByHeight(ctx, height3, make([]types.Validator, 0))

	k.Distribute(ctx, "")

	for i, validator := range validators {
		_, found := k.GetSimplifiedDelegations(ctx, height, validator.OperatorAddr)
		require.False(t, found)

		distrAcc := am.GetAccount(ctx, validator.DistributionAddr)
		balance := distrAcc.GetCoins().AmountOf(bondDenom)
		require.Equal(t, int64(0), balance)

		var amountOfAllAccount int64
		for _, delAddr := range delegators[i] {
			delAcc := am.GetAccount(ctx, delAddr)
			amountOfAllAccount += delAcc.GetCoins().AmountOf(bondDenom)
		}
		require.Equal(t, rewards[i], amountOfAllAccount)
	}

	_, found := k.GetValidatorsByHeight(ctx, height)
	require.False(t, found)

	require.EqualValues(t, len(k.addrPool.TxRelatedAddrs()), totalDelNum+21) // add 21 distribution addresses

}
