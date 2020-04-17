package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) (validatorUpdates []abci.ValidatorUpdate, completedUbds []types.UnbondingDelegation) {
	var events sdk.Events
	_, validatorUpdates, completedUbds, events = handleValidatorAndDelegations(ctx, k)
	ctx.EventManager().EmitEvents(events)
	return
}

func EndBreatheBlock(ctx sdk.Context, k keeper.Keeper) (validatorUpdates []abci.ValidatorUpdate, completedUbds []types.UnbondingDelegation) {
	var events sdk.Events
	_, validatorUpdates, completedUbds, events = handleValidatorAndDelegations(ctx, k)

	if sdk.IsUpgrade(sdk.LaunchBscUpgrade) && k.ScKeeper != nil {
		sideChainIds, storePrefixes := k.ScKeeper.GetAllSideChainPrefixes(ctx)
		for i := range storePrefixes {
			sideChainCtx := ctx.WithSideChainKeyPrefix(storePrefixes[i])
			newVals, _, _, scEvents := handleValidatorAndDelegations(sideChainCtx, k)
			saveSideChainValidatorsToIBC(ctx, sideChainIds[i], newVals, k)
			for j := range scEvents {
				scEvents[j] = scEvents[j].AppendAttributes(sdk.NewAttribute(types.AttributeKeySideChainId, sideChainIds[i]))
			}
			events = events.AppendEvents(scEvents)
			// TODO: need to add UBDs for side chains to the return value
		}
	}
	ctx.EventManager().EmitEvents(events)
	return
}

func saveSideChainValidatorsToIBC(ctx sdk.Context, sideChainId string, newVals []types.Validator, k keeper.Keeper) {
	ibcValidatorSet := make(types.IbcValidatorSet, len(newVals))
	for i := range newVals {
		ibcValidatorSet[i] = types.IbcValidator{
			ConsAddr: newVals[i].SideConsAddr,
			FeeAddr:  newVals[i].SideFeeAddr,
			DistAddr: newVals[i].DistributionAddr,
			Power:    newVals[i].GetPower().RawInt(),
		}
	}
	_, err := k.SaveValidatorSetToIbc(ctx, sideChainId, ibcValidatorSet)
	if err != nil {
		k.Logger(ctx).Error("save validators to ibc package failed: " + err.Error())
		return
	}
}

func handleValidatorAndDelegations(ctx sdk.Context, k keeper.Keeper) ([]types.Validator, []abci.ValidatorUpdate, []types.UnbondingDelegation, sdk.Events) {
	// calculate validator set changes
	newVals, validatorUpdates := k.ApplyAndReturnValidatorSetUpdates(ctx)

	k.UnbondAllMatureValidatorQueue(ctx)
	completedUbd, events := handleMatureUnbondingDelegations(k, ctx)

	redEvents := handleMatureRedelegations(k, ctx)
	events = events.AppendEvents(redEvents)

	// reset the intra-transaction counter
	k.SetIntraTxCounter(ctx, 0)
	return newVals, validatorUpdates, completedUbd, events
}

func handleMatureRedelegations(k keeper.Keeper, ctx sdk.Context) sdk.Events {
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
	return events
}

func handleMatureUnbondingDelegations(k keeper.Keeper, ctx sdk.Context) ([]types.UnbondingDelegation, sdk.Events) {
	logger := k.Logger(ctx)
	matureUnbonds := k.DequeueAllMatureUnbondingQueue(ctx, ctx.BlockHeader().Time)
	completed := make([]types.UnbondingDelegation, len(matureUnbonds))
	events := make(sdk.Events, 0, len(matureUnbonds))
	for _, dvPair := range matureUnbonds {
		ubd, err := k.CompleteUnbonding(ctx, dvPair.DelegatorAddr, dvPair.ValidatorAddr)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to complete unbonding delegation: %s", err.Error()), "delegator_address", dvPair.DelegatorAddr.String(), "validator_address", dvPair.ValidatorAddr.String())
			continue
		}
		completed = append(completed, ubd)
		events = events.AppendEvent(sdk.NewEvent(
			types.EventTypeCompleteUnbonding,
			sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddr.String()),
		))
	}

	return completed, events
}
