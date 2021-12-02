package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func (k Keeper) Distribute(ctx sdk.Context, sideChainId string) {

	// The rewards collected yesterday is decided by the validators the day before yesterday.
	// So this distribution is for the validators bonded 2 days ago
	validators, height, found := k.GetHeightValidatorsByIndex(ctx, 3)
	if !found { // do nothing, if there is no validators to be rewarded.
		return
	}

	bondDenom := k.BondDenom(ctx)
	var toPublish []types.DistributionData
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
			rewards = allocate(simDelsToSharers(delegations), remainReward)
			if commission.RawInt() > 0 { // assign rewards to self-delegator
				if _, _, err := k.bankKeeper.AddCoins(ctx, validator.GetFeeAddr(), sdk.Coins{sdk.NewCoin(bondDenom, commission.RawInt())}); err != nil {
					panic(err)
				}
			}
			// assign rewards to delegator
			changedAddrs := make([]sdk.AccAddress, len(rewards)+1)
			for i := range rewards {
				if _, _, err := k.bankKeeper.AddCoins(ctx, rewards[i].AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, rewards[i].Amount)}); err != nil {
					panic(err)
				}
				changedAddrs[i] = rewards[i].AccAddr
			}

			changedAddrs[len(rewards)] = validator.DistributionAddr
			if k.addrPool != nil {
				k.addrPool.AddAddrs(changedAddrs)
			}
		}

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

// DistributeInBreathBlock will 1) calculate rewards as Distribute does, 2) transfer commissions to all validators, and
// 3) save delegator's rewards to reward store for later distribution.
func (k Keeper) DistributeInBreathBlock(ctx sdk.Context, sideChainId string) {
	validators, height, found := k.GetHeightValidatorsByIndex(ctx, 3)
	if !found {
		return
	}

	var toPublish []types.DistributionDataV2         // data to be published in breathe blocks
	var toSaveRewards []types.StoredReward           // rewards to be saved
	var toSaveValDistAddrs []types.StoredValDistAddr // mapping between validator and distribution address, to be saved

	bondDenom := k.BondDenom(ctx)
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

			//distribute commission
			commission = totalRewardDec.Mul(validator.Commission.Rate)
			if commission.RawInt() > 0 {
				if _, _, err := k.bankKeeper.AddCoins(ctx, validator.GetFeeAddr(), sdk.Coins{sdk.NewCoin(bondDenom, commission.RawInt())}); err != nil {
					panic(err)
				}
				if _, _, err := k.bankKeeper.SubtractCoins(ctx, validator.DistributionAddr, sdk.Coins{sdk.NewCoin(bondDenom, commission.RawInt())}); err != nil {
					panic(err)
				}
			}

			//calculate rewards for delegators
			remainReward := totalRewardDec.Sub(commission)
			rewards = allocate(simDelsToSharers(delegations), remainReward)
			for i := range rewards {
				storedReward := types.StoredReward{
					Validator: validator.GetOperator(),
					AccAddr:   rewards[i].AccAddr,
					Shares:    rewards[i].Shares,
					Amount:    rewards[i].Amount,
				}
				toSaveRewards = append(toSaveRewards, storedReward)
			}

			//track validator and distribution address mapping
			toSaveValDistAddrs = append(toSaveValDistAddrs, types.StoredValDistAddr{
				Validator:      validator.OperatorAddr,
				DistributeAddr: validator.DistributionAddr})

			//update address pool
			changedAddrs := [2]sdk.AccAddress{validator.FeeAddr, validator.DistributionAddr}
			if k.addrPool != nil {
				k.addrPool.AddAddrs(changedAddrs[:])
			}
		}

		if ctx.IsDeliverTx() && k.PbsbServer != nil {
			toPublish = append(toPublish, types.DistributionDataV2{
				Validator:      validator.GetOperator(),
				SelfDelegator:  validator.GetFeeAddr(),
				DistributeAddr: validator.DistributionAddr,
				ValShares:      validator.GetDelegatorShares(),
				ValTokens:      validator.GetTokens(),
				TotalReward:    totalRewardDec,
				Commission:     commission,
				Rewards:        nil, //do not publish rewards in breathe blocks
			})
		}
	}

	//1) get batch size from parameters, 2) hard limit to make sure rewards can be distributed in a day
	batchSize := getDistributionBatchSize(k.GetParams(ctx).RewardDistributionBatchSize, int64(len(toSaveRewards)))
	batchCount := int64(len(toSaveRewards)) / batchSize
	if int64(len(toSaveRewards))%batchSize != 0 {
		batchCount = batchCount + 1
	}

	// save rewards
	var batchNo = int64(0)
	for ; batchNo < batchCount-1; batchNo++ {
		k.SetBatchRewards(ctx, int64(batchNo), toSaveRewards[batchSize*batchSize:(batchNo+1)*batchSize])
	}
	k.SetBatchRewards(ctx, int64(batchNo), toSaveRewards[batchNo*batchSize:])

	// save validator <-> distribution address map
	k.SetRewardValDistAddrs(ctx, toSaveValDistAddrs)

	// publish data if needed
	if ctx.IsDeliverTx() && len(toPublish) > 0 && k.PbsbServer != nil {
		event := types.SideDistributionEventV2{
			SideChainId: sideChainId,
			Data:        toPublish,
		}
		k.PbsbServer.Publish(event)
	}

	removeValidatorsAndDelegationsAtHeight(height, k, ctx, validators)
}

// DistributeInBlock will 1) actually distribute rewards to delegators, using reward store, 2) clear reward store if needed
func (k Keeper) DistributeInBlock(ctx sdk.Context, sideChainId string) {
	if hasNext := k.HasNextBatchRewards(ctx); hasNext == false { // already done the distribution of rewards
		return
	}

	// get batch rewards and validator <-> distribute address mapping
	rewards, key := k.GetBatchRewards(ctx)
	valDistAddrs, _ := k.GetRewardValDistAddrs(ctx)

	valDistAddrMap := make(map[string]sdk.AccAddress)
	for _, valDist := range valDistAddrs {
		valDistAddrMap[valDist.Validator.String()] = valDist.DistributeAddr
	}

	var distAddrBalanceMap = make(map[string]int64) // track distribute address balance changes
	var toPublish []types.DistributionDataV2        // data to be published in blocks
	var toPublishRewards []types.StoredReward       // rewards to be published in blocks

	bondDenom := k.BondDenom(ctx)
	for _, reward := range rewards {
		distAddr := valDistAddrMap[reward.Validator.String()]
		if value, ok := distAddrBalanceMap[distAddr.String()]; ok {
			distAddrBalanceMap[distAddr.String()] = reward.Amount + value
		} else {
			distAddrBalanceMap[distAddr.String()] = reward.Amount
		}

		if _, _, err := k.bankKeeper.AddCoins(ctx, reward.AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, reward.Amount)}); err != nil {
			panic(err)
		}

		toPublishRewards = append(toPublishRewards, reward)
	}

	for addr, value := range distAddrBalanceMap {
		accAddr, err := sdk.AccAddressFromHex(addr)
		if err != nil {
			panic(err)
		}
		if _, _, err := k.bankKeeper.SubtractCoins(ctx, accAddr, sdk.Coins{sdk.NewCoin(bondDenom, value)}); err != nil {
			panic(err)
		}
	}

	// delete the batch in store
	k.RemoveBatchRewards(ctx, key)

	// check if this batch is the last one
	if hasNext := k.HasNextBatchRewards(ctx); hasNext == false {
		k.RemoveRewardValDistAddrs(ctx)
	}

	// publish data if needed
	if ctx.IsDeliverTx() && len(toPublish) > 0 && k.PbsbServer != nil {
		toPublish = append(toPublish, types.DistributionDataV2{
			Validator:      nil,
			SelfDelegator:  nil,
			DistributeAddr: nil,
			ValShares:      sdk.Dec{},
			ValTokens:      sdk.Dec{},
			TotalReward:    sdk.Dec{},
			Commission:     sdk.Dec{},
			Rewards:        toPublishRewards,
		})
		event := types.SideDistributionEventV2{
			SideChainId: sideChainId,
			Data:        toPublish,
		}
		k.PbsbServer.Publish(event)
	}
}

// getDistributionBatchSize will adjust batch size to make sure all rewards will be distribute in a day (pre-defined block number)
// usually the batch size will not be changed, just for prevention
func getDistributionBatchSize(batchSize, totalRewardLen int64) int64 {
	//TODO: define maxBlockCount somewhere else
	maxBlockCount := int64(1000000)
	if totalRewardLen/maxBlockCount >= batchSize {
		batchSize = totalRewardLen / (maxBlockCount / 2)
	}
	return batchSize
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
