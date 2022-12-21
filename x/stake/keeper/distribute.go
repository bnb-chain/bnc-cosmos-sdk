package keeper

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"github.com/cosmos/cosmos-sdk/bsc"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/fees"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	// for getting the snapshot of validators in the specific days ago
	daysBackwardForValidatorSnapshot = 3
	// there is no cross-chain in Beacon Chain, only backward to validator snapshot of yesterday
	daysBackwardForValidatorSnapshotBeaconChain = 2
	// the count of blocks to distribute a day's rewards should be smaller than this value
	boundOfRewardDistributionBlockCount = int64(10000)
)

func (k Keeper) Distribute(ctx sdk.Context, sideChainId string) {

	// The rewards collected yesterday is decided by the validators the day before yesterday.
	// So this distribution is for the validators bonded 2 days ago
	validators, height, found := k.GetHeightValidatorsByIndex(ctx, daysBackwardForValidatorSnapshot)
	// be noted: if len(validators) == 0, it still needs to call removeValidatorsAndDelegationsAtHeight
	if !found { // do nothing, if there is no validators to be rewarded.
		return
	}

	bondDenom := k.BondDenom(ctx)
	var toPublish []types.DistributionData
	for _, validator := range validators {
		distAccCoins := k.BankKeeper.GetCoins(ctx, validator.DistributionAddr)
		totalReward := distAccCoins.AmountOf(bondDenom)
		totalRewardDec := sdk.ZeroDec()
		commission := sdk.ZeroDec()
		rewards := make([]types.PreReward, 0)
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
			if err := k.BankKeeper.SetCoins(ctx, validator.DistributionAddr, distAccCoins); err != nil {
				panic(err)
			}
			rewards = allocate(simDelsToSharers(delegations), remainReward)
			if commission.RawInt() > 0 { // assign rewards to self-delegator
				if _, _, err := k.BankKeeper.AddCoins(ctx, validator.GetFeeAddr(), sdk.Coins{sdk.NewCoin(bondDenom, commission.RawInt())}); err != nil {
					panic(err)
				}
			}
			// assign rewards to delegator
			changedAddrs := make([]sdk.AccAddress, len(rewards)+1)
			for i := range rewards {
				if _, _, err := k.BankKeeper.AddCoins(ctx, rewards[i].AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, rewards[i].Amount)}); err != nil {
					panic(err)
				}
				changedAddrs[i] = rewards[i].AccAddr
			}

			changedAddrs[len(rewards)] = validator.DistributionAddr
			if k.AddrPool != nil {
				k.AddrPool.AddAddrs(changedAddrs)
			}
		}

		if ctx.IsDeliverTx() && k.PbsbServer != nil {
			var toPublishRewards []types.Reward
			for i := range rewards {
				tokens, err := sdk.MulQuoDec(validator.GetTokens(), rewards[i].Shares, validator.GetDelegatorShares())
				if err != nil {
					panic(err)
				}
				toPublishReward := types.Reward{
					ValAddr: validator.GetOperator(),
					AccAddr: rewards[i].AccAddr,
					Tokens:  tokens,
					Amount:  rewards[i].Amount,
				}
				toPublishRewards = append(toPublishRewards, toPublishReward)
			}

			toPublish = append(toPublish, types.DistributionData{
				Validator:      validator.GetOperator(),
				SelfDelegator:  validator.GetFeeAddr(),
				DistributeAddr: validator.DistributionAddr,
				ValShares:      validator.GetDelegatorShares(),
				ValTokens:      validator.GetTokens(),
				TotalReward:    totalRewardDec,
				Commission:     commission,
				Rewards:        toPublishRewards,
			})

		}
	}

	if ctx.IsDeliverTx() && len(toPublish) > 0 && k.PbsbServer != nil {
		event := types.DistributionEvent{
			ChainId: sideChainId,
			Data:    toPublish,
		}
		k.PbsbServer.Publish(event)
	}

	removeValidatorsAndDelegationsAtHeight(height, k, ctx, validators)
}

// DistributeInBreathBlock will 1) calculate rewards as Distribute does, 2) transfer commissions to all validators, and
// 3) save delegator's rewards to reward store for later distribution.
func (k Keeper) DistributeInBreathBlock(ctx sdk.Context, sideChainId string) sdk.Events {
	ctx.Logger().Info("FeeCalculation", "currentHeight", ctx.BlockHeight(), "sideChainId", sideChainId)
	// if there are left reward distribution batches in the previous day, will distribute all of them here
	// this is only a safe guard to make sure that all the previous day's rewards are distributed
	// because this case should happen in very very special case (e.g., bc maintenance for a long time), so there is no much optimization here
	var events sdk.Events
	for k.hasNextBatchRewards(ctx) {
		singleBatchEvents := k.distributeSingleBatch(ctx, sideChainId)
		events = events.AppendEvents(singleBatchEvents)
	}

	var daysBackward int
	if sideChainId != types.ChainIDForBeaconChain {
		daysBackward = daysBackwardForValidatorSnapshot
	} else {
		daysBackward = daysBackwardForValidatorSnapshotBeaconChain
	}
	validators, height, found := k.GetHeightValidatorsByIndex(ctx, daysBackward)
	if !found {
		return events
	}

	var toPublish []types.DistributionData           // data to be published in breathe blocks
	var toSaveRewards []types.Reward                 // rewards to be saved
	var toSaveValDistAddrs []types.StoredValDistAddr // mapping between validator and distribution address, to be saved
	var rewardSum int64

	bondDenom := k.BondDenom(ctx)
	// force getting FeeFromBscToBcRatio from bc context
	feeFromBscToBcRatio := k.FeeFromBscToBcRatio(ctx.WithSideChainKeyPrefix(nil))
	avgFeeForBcVals := sdk.ZeroDec()
	if sdk.IsUpgrade(sdk.BEP159) && sideChainId == types.ChainIDForBeaconChain {
		feeForAllBcValsCoins := k.BankKeeper.GetCoins(ctx, FeeForAllBcValsAccAddr)
		feeForAllBcVals := feeForAllBcValsCoins.AmountOf(bondDenom)
		avgFeeForBcVals = sdk.NewDec(feeForAllBcVals / int64(len(validators)))
		ctx.Logger().Info("FeeCalculation", "avgFeeForBcVals", avgFeeForBcVals, "feeForAllBcVals", feeForAllBcVals, "len(validators)", len(validators))
	}

	for _, validator := range validators {
		distAccCoins := k.BankKeeper.GetCoins(ctx, validator.DistributionAddr)
		totalReward := distAccCoins.AmountOf(bondDenom)
		totalRewardDec := sdk.NewDec(totalReward)
		ctx.Logger().Info("FeeCalculation validator", "DistributionAddr", validator.DistributionAddr, "totalReward", totalReward, "height", height, "validator", validator)
		if sdk.IsUpgrade(sdk.BEP159) {
			if sideChainId != types.ChainIDForBeaconChain {
				// split a portion of fees to BC validators
				feeToBC := totalRewardDec.Mul(feeFromBscToBcRatio)
				if feeToBC.RawInt() > 0 {
					_, err := k.BankKeeper.SendCoins(ctx, validator.DistributionAddr, FeeForAllBcValsAccAddr, sdk.Coins{sdk.NewCoin(bondDenom, feeToBC.RawInt())})
					if err != nil {
						panic(err)
					}
					totalRewardDec = totalRewardDec.Sub(feeToBC)
					totalReward = totalRewardDec.RawInt()
					ctx.Logger().Info("FeeCalculation send to FeeForAllBcValsAccAddr", "feeToBC", feeToBC.RawInt(), "new totalReward", totalReward)
				}
			} else {
				// for beacon chain, split the fees accumulated in FeeForAllBcValsAccAddr
				if avgFeeForBcVals.RawInt() > 0 {
					_, err := k.BankKeeper.SendCoins(ctx, FeeForAllBcValsAccAddr, validator.DistributionAddr, sdk.Coins{sdk.NewCoin(bondDenom, avgFeeForBcVals.RawInt())})
					if err != nil {
						panic(err)
					}
					totalRewardDec = totalRewardDec.Add(avgFeeForBcVals)
					totalReward = totalRewardDec.RawInt()
					ctx.Logger().Info("FeeCalculation receive avgFeeForBcVals", "avgFeeForBcVals", avgFeeForBcVals.RawInt(), "new totalReward", totalReward)
				}
			}
		}
		commission := sdk.ZeroDec()
		rewards := make([]types.PreReward, 0)
		crossStake := make(map[string]bool)
		if totalReward > 0 {
			delegations, found := k.GetSimplifiedDelegations(ctx, height, validator.OperatorAddr)
			if !found {
				panic(fmt.Sprintf("no delegations found with height=%d, validator=%s", height, validator.OperatorAddr))
			}
			for _, del := range delegations {
				if del.CrossStake {
					crossStake[del.DelegatorAddr.String()] = true
				}
			}
			totalRewardDec = sdk.NewDec(totalReward)

			//distribute commission
			commission = totalRewardDec.Mul(validator.Commission.Rate)
			if commission.RawInt() > 0 {
				if _, _, err := k.BankKeeper.AddCoins(ctx, validator.GetFeeAddr(), sdk.Coins{sdk.NewCoin(bondDenom, commission.RawInt())}); err != nil {
					panic(err)
				}
				if _, _, err := k.BankKeeper.SubtractCoins(ctx, validator.DistributionAddr, sdk.Coins{sdk.NewCoin(bondDenom, commission.RawInt())}); err != nil {
					panic(err)
				}
			}

			//calculate rewards for delegators
			remainReward := totalRewardDec.Sub(commission)
			ctx.Logger().Info("FeeCalculation commission", "rate", validator.Commission.Rate, "commission", commission, "remainReward", remainReward, "delegations", delegations)
			rewards = allocate(simDelsToSharers(delegations), remainReward)
			for i := range rewards {
				// previous tokens calculation is in `node` repo, move it to here
				tokens, err := sdk.MulQuoDec(validator.GetTokens(), rewards[i].Shares, validator.GetDelegatorShares())
				if err != nil {
					panic(err)
				}
				toSaveReward := types.Reward{
					ValAddr:    validator.GetOperator(),
					AccAddr:    rewards[i].AccAddr,
					Tokens:     tokens,
					Amount:     rewards[i].Amount,
					CrossStake: crossStake[rewards[i].AccAddr.String()],
				}
				toSaveRewards = append(toSaveRewards, toSaveReward)
			}

			//track validator and distribution address mapping
			toSaveValDistAddrs = append(toSaveValDistAddrs, types.StoredValDistAddr{
				Validator:      validator.OperatorAddr,
				DistributeAddr: validator.DistributionAddr})

			//update address pool
			changedAddrs := [2]sdk.AccAddress{validator.FeeAddr, validator.DistributionAddr}
			if k.AddrPool != nil {
				k.AddrPool.AddAddrs(changedAddrs[:])
			}
			rewardSum += remainReward.RawInt()
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

	ctx.Logger().Info("FeeCalculation DistributeInBreathBlock", "toSaveRewards", toSaveRewards)
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

	if rewardSum > 0 {
		events = events.AppendEvent(sdk.Event{
			Type:       types.EventTypeTotalDistribution,
			Attributes: sdk.NewTags(types.AttributeKeyRewardSum, []byte(strconv.FormatInt(rewardSum, 10))),
		})
	}

	// publish data if needed
	if ctx.IsDeliverTx() && len(toPublish) > 0 && k.PbsbServer != nil {
		event := types.DistributionEvent{
			ChainId: sideChainId,
			Data:    toPublish,
		}
		k.PbsbServer.Publish(event)
	}

	removeValidatorsAndDelegationsAtHeight(height, k, ctx, validators)
	return events
}

// DistributeInBlock will 1) actually distribute rewards to delegators, using reward store, 2) clear reward store if needed
func (k Keeper) DistributeInBlock(ctx sdk.Context, sideChainId string) sdk.Events {
	if hasNext := k.hasNextBatchRewards(ctx); !hasNext { // already done the distribution of rewards
		return sdk.Events{}
	}

	return k.distributeSingleBatch(ctx, sideChainId)
}

// distributeSingleBatch will distribute an single batch of rewards if there is any
func (k Keeper) distributeSingleBatch(ctx sdk.Context, sideChainId string) sdk.Events {
	// get batch rewards and validator <-> distribution address mapping
	rewards, key := k.getNextBatchRewards(ctx)
	valDistAddrs, found := k.getRewardValDistAddrs(ctx)
	if !found {
		panic("cannot find required mapping")
	}

	valDistAddrMap := make(map[string]sdk.AccAddress)
	for _, valDist := range valDistAddrs {
		valDistAddrMap[valDist.Validator.String()] = valDist.DistributeAddr
	}

	distAddrBalanceMap := make(map[string]int64)   // track distribute address balance changes
	crossStakeAddrSet := make([]sdk.AccAddress, 0) // record cross stake address
	var toPublish []types.DistributionData         // data to be published in blocks
	var toPublishRewards []types.Reward            // rewards to be published in blocks

	var changedAddrs []sdk.AccAddress //changed addresses

	bondDenom := k.BondDenom(ctx)
	var events sdk.Events
	for _, reward := range rewards {
		distAddr := valDistAddrMap[reward.ValAddr.String()]
		if value, ok := distAddrBalanceMap[distAddr.String()]; ok {
			distAddrBalanceMap[distAddr.String()] = reward.Amount + value
		} else {
			distAddrBalanceMap[distAddr.String()] = reward.Amount
		}

		if reward.CrossStake && sdk.IsUpgrade(sdk.BEP153) {
			rewardCAoB := types.GetStakeCAoB(reward.AccAddr.Bytes(), types.RewardCAoBSalt)
			crossStakeAddrSet = append(crossStakeAddrSet, rewardCAoB)
			reward.AccAddr = rewardCAoB
		}

		if _, _, err := k.BankKeeper.AddCoins(ctx, reward.AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, reward.Amount)}); err != nil {
			panic(err)
		}

		toPublishRewards = append(toPublishRewards, reward)
		changedAddrs = append(changedAddrs, reward.AccAddr)
	}

	for addr, value := range distAddrBalanceMap {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		if _, _, err := k.BankKeeper.SubtractCoins(ctx, accAddr, sdk.Coins{sdk.NewCoin(bondDenom, value)}); err != nil {
			panic(err)
		}
		changedAddrs = append(changedAddrs, accAddr)
	}

	// cross distribute reward
	for _, addr := range crossStakeAddrSet {
		balance := k.BankKeeper.GetCoins(ctx, addr).AmountOf(bondDenom)
		if balance >= types.MinRewardThreshold {
			event, err := crossDistributeReward(k, ctx, addr, balance)
			if err != nil {
				panic(err)
			}
			events = events.AppendEvents(event)
			changedAddrs = append(changedAddrs, sdk.PegAccount)
		}
	}

	// delete the batch in store
	k.removeBatchRewards(ctx, key)

	// check whether this batch is the last one
	if hasNext := k.hasNextBatchRewards(ctx); !hasNext {
		k.removeRewardValDistAddrs(ctx)
	}

	//update address pool
	if k.AddrPool != nil {
		k.AddrPool.AddAddrs(changedAddrs[:])
	}

	// publish data if needed
	if ctx.IsDeliverTx() && len(toPublishRewards) > 0 && k.PbsbServer != nil {
		toPublish = append(toPublish, types.DistributionData{
			Validator:      nil,
			SelfDelegator:  nil,
			DistributeAddr: nil,
			ValShares:      sdk.Dec{},
			ValTokens:      sdk.Dec{},
			TotalReward:    sdk.Dec{},
			Commission:     sdk.Dec{},
			Rewards:        toPublishRewards, // only publish rewards in normal block
		})
		event := types.DistributionEvent{
			ChainId: sideChainId,
			Data:    toPublish,
		}
		k.PbsbServer.Publish(event)
	}
	return events
}

// getDistributionBatchSize will adjust batch size to make sure all rewards will be distribute in a day (pre-defined block number)
// usually the batch size will not be changed, just for prevention
func getDistributionBatchSize(batchSize, totalRewardLen int64) int64 {
	if totalRewardLen/boundOfRewardDistributionBlockCount >= batchSize {
		batchSize = totalRewardLen / (boundOfRewardDistributionBlockCount / 2)
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

func crossDistributeReward(k Keeper, ctx sdk.Context, rewardCAoB sdk.AccAddress, amount int64) (sdk.Events, error) {
	denom := k.BondDenom(ctx)
	relayFeeCalc := fees.GetCalculator(types.CrossDistributeRewardRelayFee)
	if relayFeeCalc == nil {
		return sdk.Events{}, fmt.Errorf("no fee calculator of transferOutRewards")
	}
	relayFee := relayFeeCalc(nil)
	if relayFee.Tokens.AmountOf(denom) >= amount {
		return sdk.Events{}, sdk.ErrInternal("not enough funds to cover relay fee")
	}
	bscRelayFee := bsc.ConvertBCAmountToBSCAmount(relayFee.Tokens.AmountOf(denom))

	bscTransferAmount := new(big.Int).Sub(bsc.ConvertBCAmountToBSCAmount(amount), bscRelayFee)
	delAddr := types.GetStakeCAoB(rewardCAoB.Bytes(), types.RewardCAoBSalt)
	delBscAddrAcc := types.GetStakeCAoB(delAddr.Bytes(), types.DelegateCAoBSalt)
	delBscAddr := hex.EncodeToString(delBscAddrAcc.Bytes())
	recipient, err := sdk.NewSmartChainAddress(delBscAddr)
	if err != nil {
		return sdk.Events{}, err
	}

	transferPackage := types.CrossStakeDistributeRewardSynPackage{
		EventType: types.CrossStakeTypeDistributeReward,
		Amount:    bscTransferAmount,
		Recipient: recipient,
	}
	encodedPackage, err := rlp.EncodeToBytes(transferPackage)
	if err != nil {
		return sdk.Events{}, err
	}

	sendSeq, sdkErr := k.ibcKeeper.CreateRawIBCPackageByIdWithFee(ctx.DepriveSideChainKeyPrefix(), k.DestChainId, types.CrossStakeChannelID, sdk.SynCrossChainPackageType,
		encodedPackage, *bscRelayFee)
	if sdkErr != nil {
		return sdk.Events{}, sdkErr
	}

	if _, sdkErr := k.BankKeeper.SendCoins(ctx, rewardCAoB, sdk.PegAccount, sdk.Coins{sdk.NewCoin(denom, amount)}); sdkErr != nil {
		return sdk.Events{}, sdkErr
	}

	// publish data if needed
	if ctx.IsDeliverTx() && k.PbsbServer != nil {
		event := pubsub.CrossTransferEvent{
			ChainId:    k.DestChainName,
			RelayerFee: relayFee.Tokens.AmountOf(denom),
			Type:       types.TransferOutType,
			From:       rewardCAoB.String(),
			Denom:      denom,
			To:         []pubsub.CrossReceiver{{sdk.PegAccount.String(), amount}},
		}
		k.PbsbServer.Publish(event)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte{uint8(types.CrossStakeTypeDistributeReward)},
		types.TagCrossStakeChannel, []byte{uint8(types.CrossStakeChannelID)},
		types.TagCrossStakeSendSequence, []byte(strconv.FormatUint(sendSeq, 10)),
	)
	resultTags = append(resultTags, sdk.GetPegInTag(denom, amount))

	events := sdk.Events{sdk.Event{
		Type:       types.EventTypeCrossStake,
		Attributes: resultTags,
	}}
	return events, nil
}

func (k Keeper) GetPrevProposerDistributionAddr(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(PrevProposerDistributionAddrKey)
	return bz
}

func (k Keeper) SetPrevProposerDistributionAddr(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(PrevProposerDistributionAddrKey, addr)
}
