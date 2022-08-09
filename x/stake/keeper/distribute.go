package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func (k Keeper) DistributeBeaconChainInBreathBlock(ctx sdk.Context, accumulatedFeeFromBscToBc int64) {
	for k.hasNextBatchRewards(ctx) {
		k.distributeSingleBatch(ctx, MockSideChainIDForBeaconChain)
	}

	validators, height, found := k.GetHeightValidatorsByIndex(ctx, daysBackwardForValidatorSnapshot)
	if !found {
		return
	}
	avgFee := accumulatedFeeFromBscToBc / int64(len(validators))
	changeFee := accumulatedFeeFromBscToBc - avgFee*int64(len(validators))

	var toPublish []types.DistributionData           // data to be published in breathe blocks
	var toSaveRewards []types.Reward                 // rewards to be saved
	var toSaveValDistAddrs []types.StoredValDistAddr // mapping between validator and distribution address, to be saved

	bondDenom := k.BondDenom(ctx)
	for i, validator := range validators {
		distAccCoins := k.bankKeeper.GetCoins(ctx, validator.DistributionAddr)
		totalReward := distAccCoins.AmountOf(bondDenom) + avgFee
		// give the remaining change to the first validator
		if i == 0 {
			totalReward += changeFee
		}
		totalRewardDec := sdk.NewDec(totalReward)
		commission := sdk.ZeroDec()
		rewards := make([]types.PreReward, 0)
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
				// previous tokens calculation is in `node` repo, move it to here
				tokens, err := sdk.MulQuoDec(validator.GetTokens(), rewards[i].Shares, validator.GetDelegatorShares())
				if err != nil {
					panic(err)
				}
				toSaveReward := types.Reward{
					ValAddr: validator.GetOperator(),
					AccAddr: rewards[i].AccAddr,
					Tokens:  tokens,
					Amount:  rewards[i].Amount,
				}
				toSaveRewards = append(toSaveRewards, toSaveReward)
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
			toPublish = append(toPublish, types.DistributionData{
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

	if len(toSaveRewards) > 0 { //to save rewards
		//1) get batch size from parameters, 2) hard limit to make sure rewards can be distributed in a day
		batchSize := getDistributionBatchSize(k.GetParams(ctx).RewardDistributionBatchSize, int64(len(toSaveRewards)))
		batchCount := int64(len(toSaveRewards)) / batchSize
		if int64(len(toSaveRewards))%batchSize != 0 {
			batchCount = batchCount + 1
		}

		// save rewards
		var batchNo = int64(0)
		for ; batchNo < batchCount-1; batchNo++ {
			k.setBatchRewards(ctx, batchNo, toSaveRewards[batchNo*batchSize:(batchNo+1)*batchSize])
		}
		k.setBatchRewards(ctx, batchNo, toSaveRewards[batchNo*batchSize:])

		// save validator <-> distribution address map
		k.setRewardValDistAddrs(ctx, toSaveValDistAddrs)
	}

	// publish data if needed
	if ctx.IsDeliverTx() && len(toPublish) > 0 && k.PbsbServer != nil {
		event := types.SideDistributionEvent{
			SideChainId: MockSideChainIDForBeaconChain,
			Data:        toPublish,
		}
		k.PbsbServer.Publish(event)
	}

	removeValidatorsAndDelegationsAtHeight(height, k, ctx, validators)
}
