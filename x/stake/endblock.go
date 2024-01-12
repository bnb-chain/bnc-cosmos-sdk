package stake

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) (validatorUpdates []abci.ValidatorUpdate, completedUbds []types.UnbondingDelegation) {
	// only change validator set in breath block after BEP159
	var (
		events   sdk.Events
		csEvents sdk.Events
	)

	if !sdk.IsUpgrade(sdk.BEP159) {
		_, validatorUpdates, completedUbds, _, events = handleValidatorAndDelegations(ctx, k)
	} else {
		k.DistributeInBlock(ctx, types.ChainIDForBeaconChain)
		validatorUpdates = k.PopPendingABCIValidatorUpdate(ctx)
	}

	var (
		sideChainIds  []string
		storePrefixes [][]byte
	)
	if sdk.IsUpgrade(sdk.BEP128) {
		sideChainIds, storePrefixes = k.ScKeeper.GetAllSideChainPrefixes(ctx)
		if len(sideChainIds) == len(storePrefixes) {
			for i := range storePrefixes {
				sideChainCtx := ctx.WithSideChainKeyPrefix(storePrefixes[i])
				csEvents = k.DistributeInBlock(sideChainCtx, sideChainIds[i])
			}
		} else {
			panic("sideChainIds does not equal to sideChainStores")
		}
	}

	if len(storePrefixes) > 0 && sdk.IsUpgrade(sdk.SecondSunsetFork) {
		for i := range storePrefixes {
			events.AppendEvents(handleRefundStake(ctx, storePrefixes[i], k))
		}
	}

	if len(storePrefixes) > 0 && sdk.IsUpgrade(sdk.FirstSunsetFork) {
		for i := range storePrefixes {
			sideChainCtx := ctx.WithSideChainKeyPrefix(storePrefixes[i])
			_, unBoundedEvents := handleMatureUnbondingDelegations(k, sideChainCtx)

			events = append(events, unBoundedEvents...)
		}
	}

	if sdk.IsUpgrade(sdk.BEP153) {
		events = events.AppendEvents(csEvents)
	}

	ctx.EventManager().EmitEvents(events)
	return
}

func EndBreatheBlock(ctx sdk.Context, k keeper.Keeper) (validatorUpdates []abci.ValidatorUpdate, completedUbds []types.UnbondingDelegation) {
	var events sdk.Events
	var newVals []types.Validator
	var completedREDs []types.DVVTriplet
	newVals, validatorUpdates, completedUbds, completedREDs, events = handleValidatorAndDelegations(ctx, k)
	ctx.Logger().Debug("EndBreatheBlock", "newValsLen", len(newVals), "newVals", newVals)
	publishCompletedUBD(k, completedUbds, ChainIDForBeaconChain, ctx.BlockHeight())
	publishCompletedRED(k, completedREDs, ChainIDForBeaconChain)
	if k.PbsbServer != nil {
		sideValidatorsEvent := types.ElectedValidatorsEvent{
			Validators: newVals,
			ChainId:    ChainIDForBeaconChain,
		}
		k.PbsbServer.Publish(sideValidatorsEvent)
	}
	if sdk.IsUpgrade(sdk.BEP159) {
		storeValidatorsWithHeight(ctx, newVals, k)
	}

	if sdk.IsUpgrade(sdk.LaunchBscUpgrade) && k.ScKeeper != nil {
		// distribute sidechain rewards
		sideChainIds, storePrefixes := k.ScKeeper.GetAllSideChainPrefixes(ctx)
		for i := range storePrefixes {
			sideChainCtx := ctx.WithSideChainKeyPrefix(storePrefixes[i])
			newVals, _, completedUbds, completedREDs, scEvents := handleValidatorAndDelegations(sideChainCtx, k)
			if k.ExistHeightValidators(sideChainCtx) { // will not send ibc package if no snapshot of validators stored ever
				saveSideChainValidatorsToIBC(ctx, sideChainIds[i], newVals, k)
			}
			for j := range scEvents {
				scEvents[j] = scEvents[j].AppendAttributes(sdk.NewAttribute(types.AttributeKeySideChainId, sideChainIds[i]))
			}
			events = events.AppendEvents(scEvents)
			// TODO: need to add UBDs for side chains to the return value

			storeValidatorsWithHeight(sideChainCtx, newVals, k)

			var csEvents sdk.Events
			if sdk.IsUpgrade(sdk.BEP128) {
				csEvents = k.DistributeInBreathBlock(sideChainCtx, sideChainIds[i])
			} else {
				k.Distribute(sideChainCtx, sideChainIds[i])
			}
			if sdk.IsUpgrade(sdk.BEP153) {
				events = events.AppendEvents(csEvents)
			}

			publishCompletedUBD(k, completedUbds, sideChainIds[i], ctx.BlockHeight())
			publishCompletedRED(k, completedREDs, sideChainIds[i])
		}
		if sdk.IsUpgrade(sdk.BEP159) {
			// distribute beacon chain rewards
			k.DistributeInBreathBlock(ctx, types.ChainIDForBeaconChain)
		}
	}
	ctx.EventManager().EmitEvents(events)
	return
}

func publishCompletedUBD(k keeper.Keeper, completedUbds []types.UnbondingDelegation, sideChainId string, height int64) {
	if k.PbsbServer != nil && len(completedUbds) > 0 {
		compUBDsEvent := types.CompletedUBDEvent{
			CompUBDs: completedUbds,
			ChainId:  sideChainId,
		}
		k.PbsbServer.Publish(compUBDsEvent)
	}
}

func publishCompletedRED(k keeper.Keeper, completedReds []types.DVVTriplet, sideChainId string) {
	if k.PbsbServer != nil && len(completedReds) > 0 {
		compREDsEvent := types.CompletedREDEvent{
			CompREDs: completedReds,
			ChainId:  sideChainId,
		}
		k.PbsbServer.Publish(compREDsEvent)
	}
}

func saveSideChainValidatorsToIBC(ctx sdk.Context, sideChainId string, newVals []types.Validator, k keeper.Keeper) {
	if sdk.IsUpgrade(sdk.BEP126) {
		ibcPackage := types.IbcValidatorWithVoteAddrSetPackage{
			Type:         types.StakePackageType,
			ValidatorSet: make([]types.IbcValidatorWithVoteAddr, len(newVals)),
		}
		for i := range newVals {
			ibcPackage.ValidatorSet[i] = types.IbcValidatorWithVoteAddr{
				ConsAddr: newVals[i].SideConsAddr,
				FeeAddr:  newVals[i].SideFeeAddr,
				DistAddr: newVals[i].DistributionAddr,
				Power:    uint64(newVals[i].GetPower().RawInt()),
				VoteAddr: newVals[i].SideVoteAddr,
			}
		}
		_, err := k.SaveValidatorWithVoteAddrSetToIbc(ctx, sideChainId, ibcPackage)
		if err != nil {
			k.Logger(ctx).Error("save validators to ibc package failed: " + err.Error())
			return
		}
	} else {
		ibcPackage := types.IbcValidatorSetPackage{
			Type:         types.StakePackageType,
			ValidatorSet: make([]types.IbcValidator, len(newVals)),
		}
		for i := range newVals {
			ibcPackage.ValidatorSet[i] = types.IbcValidator{
				ConsAddr: newVals[i].SideConsAddr,
				FeeAddr:  newVals[i].SideFeeAddr,
				DistAddr: newVals[i].DistributionAddr,
				Power:    uint64(newVals[i].GetPower().RawInt()),
			}
		}
		_, err := k.SaveValidatorSetToIbc(ctx, sideChainId, ibcPackage)
		if err != nil {
			k.Logger(ctx).Error("save validators to ibc package failed: " + err.Error())
			return
		}
	}
	if k.PbsbServer != nil {
		sideValidatorsEvent := types.ElectedValidatorsEvent{
			Validators: newVals,
			ChainId:    sideChainId,
		}
		k.PbsbServer.Publish(sideValidatorsEvent)
	}
}

func storeValidatorsWithHeight(ctx sdk.Context, validators []types.Validator, k keeper.Keeper) {
	blockHeight := ctx.BlockHeight()
	for _, validator := range validators {
		simplifiedDelegations := k.GetSimplifiedDelegationsByValidator(ctx, validator.OperatorAddr)
		k.SetSimplifiedDelegations(ctx, blockHeight, validator.OperatorAddr, simplifiedDelegations)
	}
	k.SetValidatorsByHeight(ctx, blockHeight, validators)
}

func handleValidatorAndDelegations(ctx sdk.Context, k keeper.Keeper) ([]types.Validator, []abci.ValidatorUpdate, []types.UnbondingDelegation, []types.DVVTriplet, sdk.Events) {
	// calculate validator set changes
	var newVals []types.Validator
	var validatorUpdates []abci.ValidatorUpdate
	ctx.Logger().Debug("handleValidatorAndDelegations", "height", ctx.BlockHeader().Height, "addSnapshot", sdk.IsUpgrade(sdk.BEP159) && ctx.SideChainKeyPrefix() == nil)
	if sdk.IsUpgrade(sdk.BEP159) && ctx.SideChainKeyPrefix() == nil {
		validatorUpdatesOfEditValidators := k.PopPendingABCIValidatorUpdate(ctx)
		var validatorUpdatesElect []abci.ValidatorUpdate
		newVals, validatorUpdatesElect = k.UpdateAndElectValidators(ctx)
		// remove the duplicates
		validatorUpdateMap := make(map[string]int)
		combinedSlice := append(validatorUpdatesOfEditValidators[:], validatorUpdatesElect...)
		for _, v := range combinedSlice {
			if index, ok := validatorUpdateMap[v.PubKey.String()]; ok {
				validatorUpdates[index] = v
			} else {
				validatorUpdateMap[v.PubKey.String()] = len(validatorUpdates)
				validatorUpdates = append(validatorUpdates, v)
			}
		}
		ctx.Logger().Debug("handleValidatorAndDelegations", "height", ctx.BlockHeight(), "validatorUpdates", validatorUpdates, "validatorUpdatesOfEditValidators", validatorUpdatesOfEditValidators)
	} else {
		newVals, validatorUpdates = k.ApplyAndReturnValidatorSetUpdates(ctx)
	}

	k.UnbondAllMatureValidatorQueue(ctx)
	completedUbd, events := handleMatureUnbondingDelegations(k, ctx)

	completedREDs, redEvents := handleMatureRedelegations(k, ctx)
	events = events.AppendEvents(redEvents)

	// reset the intra-transaction counter
	k.SetIntraTxCounter(ctx, 0)
	return newVals, validatorUpdates, completedUbd, completedREDs, events
}

func handleMatureRedelegations(k keeper.Keeper, ctx sdk.Context) ([]types.DVVTriplet, sdk.Events) {
	matureRedelegations := k.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockHeader().Time)
	events := make(sdk.Events, 0, len(matureRedelegations))
	for _, dvvTriplet := range matureRedelegations {
		err := k.CompleteRedelegation(ctx, dvvTriplet.DelegatorAddr, dvvTriplet.ValidatorSrcAddr, dvvTriplet.ValidatorDstAddr)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to complete redelegation: %s", err.Error()), "delegator_address", dvvTriplet.DelegatorAddr.String(), "source_validator_address", dvvTriplet.ValidatorSrcAddr.String(), "source_validator_address", dvvTriplet.ValidatorDstAddr.String())
			continue
		}
		events = events.AppendEvent(sdk.NewEvent(
			types.EventTypeCompleteRedelegation,
			sdk.NewAttribute(types.AttributeKeyDelegator, dvvTriplet.DelegatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeySrcValidator, dvvTriplet.ValidatorSrcAddr.String()),
			sdk.NewAttribute(types.AttributeKeyDstValidator, dvvTriplet.ValidatorDstAddr.String()),
		))
	}
	return matureRedelegations, events
}

func handleMatureUnbondingDelegations(k keeper.Keeper, ctx sdk.Context) ([]types.UnbondingDelegation, sdk.Events) {
	logger := k.Logger(ctx)
	matureUnbonds := k.DequeueAllMatureUnbondingQueue(ctx, ctx.BlockHeader().Time)
	completed := make([]types.UnbondingDelegation, len(matureUnbonds))
	events := make(sdk.Events, 0, len(matureUnbonds))
	for i, dvPair := range matureUnbonds {
		ubd, csEvents, err := k.CompleteUnbonding(ctx, dvPair.DelegatorAddr, dvPair.ValidatorAddr)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to complete unbonding delegation: %s", err.Error()), "delegator_address", dvPair.DelegatorAddr.String(), "validator_address", dvPair.ValidatorAddr.String())
			continue
		}
		completed[i] = ubd
		if sdk.IsUpgrade(sdk.BEP153) {
			events = events.AppendEvents(csEvents)
		}
		events = events.AppendEvent(sdk.NewEvent(
			types.EventTypeCompleteUnbonding,
			sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddr.String()),
		))
	}

	return completed, events
}

const (
	maxProcessedRefundCount  = 10
	maxProcessedRefundFailed = 200
)

func handleRefundStake(ctx sdk.Context, sideChainPrefix []byte, k keeper.Keeper) sdk.Events {
	sideChainCtx := ctx.WithSideChainKeyPrefix(sideChainPrefix)
	iterator := k.IteratorAllDelegations(sideChainCtx)
	defer iterator.Close()
	var refundEvents sdk.Events
	succeedCount := 0
	failedCount := 0
	boundDenom := k.BondDenom(sideChainCtx)
	bscSideChainId := k.ScKeeper.BscSideChainId(ctx)

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.CDC(), iterator.Key(), iterator.Value())
		if delegation.CrossStake {
			ctx = ctx.WithCrossStake(true)
		}

		result := handleMsgSideChainUndelegate(ctx, types.MsgSideChainUndelegate{
			DelegatorAddr: delegation.DelegatorAddr,
			ValidatorAddr: delegation.ValidatorAddr,
			Amount:        sdk.NewCoin(boundDenom, delegation.GetShares().RawInt()),
			SideChainId:   bscSideChainId,
		}, k)
		refundEvents = refundEvents.AppendEvents(result.Events)
		if !result.IsOK() {
			ctx.Logger().Debug("handleRefundStake failed",
				"delegator", delegation.DelegatorAddr.String(),
				"validator", delegation.ValidatorAddr.String(),
				"amount", delegation.GetShares().String(),
				"sideChainId", bscSideChainId,
				"result", fmt.Sprintf("%+v", result),
			)
			// this is to prevent too many delegation is in unbounded state
			// if too many delegation is in unbounded state, it will cause too many iteration in the block
			failedCount++
			if failedCount >= maxProcessedRefundFailed {
				break
			}
		}

		if result.IsOK() && delegation.CrossStake {
			k.SetAutoUnDelegate(sideChainCtx, delegation.DelegatorAddr, delegation.ValidatorAddr)
		}

		ctx.Logger().Info("handleRefundStake after SecondSunsetFork",
			"delegator", delegation.DelegatorAddr.String(),
			"validator", delegation.ValidatorAddr.String(),
			"amount", delegation.GetShares().String(),
			"sideChainId", bscSideChainId,
		)
		succeedCount++
		if succeedCount >= maxProcessedRefundCount {
			break
		}
	}

	return refundEvents
}
