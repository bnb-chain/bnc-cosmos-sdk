package stake

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) (validatorUpdates []abci.ValidatorUpdate, completedUbds []types.UnbondingDelegation, endBlockerTags sdk.Tags) {
	endBlockerTags = sdk.EmptyTags()
	_, validatorUpdates, completedUbds, endBlockerTags = handleValidatorAndDelegations(ctx, k)
	return
}

func EndBreatheBlock(ctx sdk.Context, k keeper.Keeper) (validatorUpdates []abci.ValidatorUpdate, completedUbds []types.UnbondingDelegation, endBlockerTags sdk.Tags) {
	endBlockerTags = sdk.EmptyTags()
	_, validatorUpdates, completedUbds, endBlockerTags = handleValidatorAndDelegations(ctx, k)

	if sdk.IsUpgrade(sdk.LaunchBscUpgrade) {
		sideChainIds, storePrefixes := k.GetAllSideChainPrefixes(ctx)
		for i := range storePrefixes {
			sideChainCtx := ctx.WithSideChainKeyPrefix(storePrefixes[i])

			newVals, _, _, _ := handleValidatorAndDelegations(sideChainCtx, k)
			ibcTags :=  saveSideChainValidatorsToIBC(ctx, sideChainIds[i], newVals, k)
			endBlockerTags = endBlockerTags.AppendTags(ibcTags)

			storeValidatorsWithHeight(newVals, k, ctx)
			markBreatheBlock(ctx,k)

			k.Distribute(ctx, true)
		}
		// TODO: may need to change the return values

	}
	return
}

func saveSideChainValidatorsToIBC(ctx sdk.Context, sideChainId string, newVals []types.Validator, k keeper.Keeper) (sdk.Tags) {
	ibcValidatorSet := make(types.IbcValidatorSet, len(newVals))
	for i := range newVals {
		ibcValidatorSet[i] = types.IbcValidator{
			ConsAddr: newVals[i].SideConsAddr,
			FeeAddr:  newVals[i].SideFeeAddr,
			DistAddr: newVals[i].DistributionAddr,
			Power:    newVals[i].GetPower().RawInt(),
		}
	}
	sequence, err := k.SaveValidatorSetToIbc(ctx, sideChainId, ibcValidatorSet)
	if err != nil {
		k.Logger(ctx).Error("save validators to ibc package failed: " + err.Error())
	}
	return sdk.NewTags(tags.SideChainStakingPackageSequence, []byte(strconv.Itoa(int(sequence))))
}

func markBreatheBlock(ctx sdk.Context, k keeper.Keeper) {
	k.SetBreatheBlockHeight(ctx, ctx.BlockHeight(), ctx.BlockHeader().Time)
}

func storeValidatorsWithHeight(validators []types.Validator, k keeper.Keeper, ctx sdk.Context) {
	validatorsByHeight := make([]types.Validator, 0)
	if validators != nil && len(validators) > 0 {
		for _, validator := range validators {
			simplifiedDelegations := k.GetDelegationsSimplifiedByValidator(ctx, validator.OperatorAddr)

			k.SetSimplifiedDelegations(ctx, ctx.BlockHeight(), validator.OperatorAddr, simplifiedDelegations)

			validatorsByHeight = append(validatorsByHeight, validator)
		}
		k.SetValidatorsByHeight(ctx, ctx.BlockHeight(), validatorsByHeight)
	}
}


func handleValidatorAndDelegations(ctx sdk.Context, k keeper.Keeper) ([]types.Validator, []abci.ValidatorUpdate, []types.UnbondingDelegation, sdk.Tags) {
	endBlockerTags := sdk.EmptyTags()

	k.UnbondAllMatureValidatorQueue(ctx)
	completedUbd, tags := handleMatureUnbondingDelegations(k, ctx)
	endBlockerTags.AppendTags(tags)

	tags = handleMatureRedelegations(k, ctx)
	endBlockerTags.AppendTags(tags)

	// reset the intra-transaction counter
	k.SetIntraTxCounter(ctx, 0)

	// calculate validator set changes
	newVals, validatorUpdates := k.ApplyAndReturnValidatorSetUpdates(ctx)
	return newVals, validatorUpdates, completedUbd, endBlockerTags
}

func handleMatureRedelegations(k keeper.Keeper, ctx sdk.Context) (sdk.Tags) {
	endBlockerTags := sdk.EmptyTags()
	matureRedelegations := k.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockHeader().Time)
	for _, dvvTriplet := range matureRedelegations {
		err := k.CompleteRedelegation(ctx, dvvTriplet.DelegatorAddr, dvvTriplet.ValidatorSrcAddr, dvvTriplet.ValidatorDstAddr)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to complete redelegation: %s", err.Error()), "delegator_address", dvvTriplet.DelegatorAddr.String(), "source_validator_address", dvvTriplet.ValidatorSrcAddr.String(), "source_validator_address", dvvTriplet.ValidatorDstAddr.String())
			continue
		}
		endBlockerTags.AppendTags(sdk.NewTags(
			tags.Action, tags.ActionCompleteRedelegation,
			tags.Delegator, []byte(dvvTriplet.DelegatorAddr.String()),
			tags.SrcValidator, []byte(dvvTriplet.ValidatorSrcAddr.String()),
			tags.DstValidator, []byte(dvvTriplet.ValidatorDstAddr.String()),
		))
	}
	return endBlockerTags
}

func handleMatureUnbondingDelegations(k keeper.Keeper, ctx sdk.Context) ([]types.UnbondingDelegation, sdk.Tags) {
	logger := k.Logger(ctx)
	matureUnbonds := k.DequeueAllMatureUnbondingQueue(ctx, ctx.BlockHeader().Time)
	completed := make([]types.UnbondingDelegation, len(matureUnbonds))
	var endBlockerTags sdk.Tags
	for _, dvPair := range matureUnbonds {
		ubd, found := k.GetUnbondingDelegation(ctx, dvPair.DelegatorAddr, dvPair.ValidatorAddr)
		if !found {
			logger.Error("Failed to get unbonding delegation", "delegator_address", dvPair.DelegatorAddr.String(), "validator_address", dvPair.ValidatorAddr.String())
			continue
		}
		err := k.CompleteUnbonding(ctx, dvPair.DelegatorAddr, dvPair.ValidatorAddr)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to complete unbonding delegation: %s", err.Error()), "delegator_address", dvPair.DelegatorAddr.String(), "validator_address", dvPair.ValidatorAddr.String())
			continue
		}
		completed = append(completed, ubd)
		endBlockerTags.AppendTags(sdk.NewTags(
			tags.Action, ActionCompleteUnbonding,
			tags.Delegator, []byte(dvPair.DelegatorAddr.String()),
			tags.SrcValidator, []byte(dvPair.ValidatorAddr.String()),
		))
	}
	return completed, endBlockerTags
}
