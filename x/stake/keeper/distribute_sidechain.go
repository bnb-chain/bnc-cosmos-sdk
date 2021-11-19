package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"time"
)

func (k Keeper) DistributeBatchRewards(ctx sdk.Context) {
	if sdk.IsUpgrade(sdk.LaunchBscUpgrade) && k.ScKeeper != nil {
		sideChainId := "bsc"
		bondDenom := k.BondDenom(ctx)

		distributeStart := time.Now()
		fmt.Println("PERF_STAKING start distribute batch: ", distributeStart.Format("20060102150405"))
		rewards := k.GetRewards(ctx, sideChainId, 0)
		if rewards != nil {
			for i := range rewards {
				if _, _, err := k.bankKeeper.AddCoins(ctx, rewards[i].AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, rewards[i].Amount)}); err != nil {
					panic(err)
				}
			}
			distributeElapsed := time.Since(distributeStart)
			fmt.Println("PERF_STAKING delegation rewards batch distribute: ", distributeElapsed)
			fmt.Println("PERF_STAKING delegation rewards batch size: ", len(rewards))
		}
	}
}

func (k Keeper) Distribute(ctx sdk.Context, sideChainId string) {

	// The rewards collected yesterday is decided by the validators the day before yesterday.
	// So this distribution is for the validators bonded 2 days ago
	validators, height, found := k.GetHeightValidatorsByIndex(ctx, 3)
	if !found { // do nothing, if there is no validators to be rewarded.
		return
	}

	bondDenom := k.BondDenom(ctx)
	var toPublish []types.DistributionData

	//------------------------------
	start := time.Now()
	fmt.Println("PERF_STAKING start distribute: ", start.Format("20060102150405"))
	//------------------------------

	for _, validator := range validators {
		distAccCoins := k.bankKeeper.GetCoins(ctx, validator.DistributionAddr)
		totalReward := distAccCoins.AmountOf(bondDenom)
		totalRewardDec := sdk.ZeroDec()
		commission := sdk.ZeroDec()
		rewards := make([]types.Reward, 0)
		if totalReward > 0 {
			delegations, found := k.GetSimplifiedDelegations(ctx, height, validator.OperatorAddr)
			if !found {
				panic(fmt.Sprintf("no delegations found with height=%d, validator=%s", height, validator.OperatorAddr))
			}
			totalRewardDec = sdk.NewDec(totalReward)
			commission = totalRewardDec.Mul(validator.Commission.Rate)
			remainReward := totalRewardDec.Sub(commission)
			// remove all balance of bondDenom from Distribution account
			distAccCoins = distAccCoins.Minus(sdk.Coins{sdk.NewCoin(bondDenom, totalReward)})
			if err := k.bankKeeper.SetCoins(ctx, validator.DistributionAddr, distAccCoins); err != nil {
				panic(err)
			}
			calStart := time.Now()
			rewardsThisValidator := allocate(simDelsToSharers(delegations), remainReward)
			calElapsed := time.Since(calStart)
			fmt.Println("PERF_STAKING cal rewards: ", calElapsed)

			commissionStart := time.Now()
			if commission.RawInt() > 0 { // assign rewards to self-delegator
				if _, _, err := k.bankKeeper.AddCoins(ctx, validator.GetFeeAddr(), sdk.Coins{sdk.NewCoin(bondDenom, commission.RawInt())}); err != nil {
					panic(err)
				}
			}
			commissionElapsed := time.Since(commissionStart)
			fmt.Println("PERF_STAKING commission distribute: ", commissionElapsed)

			rewards = append(rewards, rewardsThisValidator...)

			// assign rewards to delegator
			//changedAddrs := make([]sdk.AccAddress, len(rewards)+1)
			//for i := range rewards {
			//	if _, _, err := k.bankKeeper.AddCoins(ctx, rewards[i].AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, rewards[i].Amount)}); err != nil {
			//		panic(err)
			//	}
			//	changedAddrs[i] = rewards[i].AccAddr
			//}

			//changedAddrs[len(rewards)] = validator.DistributionAddr
			//if k.addrPool != nil {
			//	k.addrPool.AddAddrs(changedAddrs)
			//}
		}

		storeStart := time.Now()
		//todo: not totally correct, refine later
		batchSize := 1000
		batchCount := len(rewards) / batchSize
		if len(rewards)%batchSize != 0 {
			batchCount = batchCount + 1
		}

		for i := 0; i < batchCount-1; i++ {
			k.SetRewards(ctx, sideChainId, int64(i), rewards[i*batchSize:(i+1)*batchSize])
		}
		k.SetRewards(ctx, sideChainId, int64(batchCount), rewards[batchCount*batchSize:])

		storeElapsed := time.Since(storeStart)
		fmt.Println("PERF_STAKING delegation rewards store: ", storeElapsed)
		fmt.Println("PERF_STAKING delegation rewards total size: ", len(rewards))

		//------------------------------
		totalElapsed := time.Since(start)
		fmt.Println("PERF_STAKING total: ", totalElapsed)
		//------------------------------

		if ctx.IsDeliverTx() && k.PbsbServer != nil {
			toPublish = append(toPublish, types.DistributionData{
				Validator:      validator.GetOperator(),
				SelfDelegator:  validator.GetFeeAddr(),
				DistributeAddr: validator.DistributionAddr,
				ValShares:      validator.GetDelegatorShares(),
				ValTokens:      validator.GetTokens(),
				TotalReward:    totalRewardDec,
				Commission:     commission,
				Rewards:        rewards,
			})

		}
	}

	if ctx.IsDeliverTx() && len(toPublish) > 0 && k.PbsbServer != nil {
		event := types.SideDistributionEvent{
			SideChainId: sideChainId,
			Data:        toPublish,
		}
		k.PbsbServer.Publish(event)
	}

	removeValidatorsAndDelegationsAtHeight(height, k, ctx, validators)
}

func simDelsToSharers(simDels []types.SimplifiedDelegation) []types.Sharer {
	sharers := make([]types.Sharer, len(simDels))
	for i, del := range simDels {
		sharers[i] = types.Sharer{AccAddr: del.DelegatorAddr, Shares: del.Shares}
	}
	return sharers
}

func removeValidatorsAndDelegationsAtHeight(height int64, k Keeper, ctx sdk.Context, validators []types.Validator) {
	for _, validator := range validators {
		k.RemoveSimplifiedDelegations(ctx, height, validator.OperatorAddr)
	}
	k.RemoveValidatorsByHeight(ctx, height)
}
