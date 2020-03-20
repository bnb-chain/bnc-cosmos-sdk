package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) (validatorUpdates []abci.ValidatorUpdate, completedUbds []types.UnbondingDelegation, endBlockerTags sdk.Tags) {
	endBlockerTags = sdk.EmptyTags()
	logger := ctx.Logger().With("module", "stake")

	validatorUpdates, completedUbds, endBlockerTags = handleValidatorAndDelegations(ctx, k, logger)

	if sdk.IsUpgrade(sdk.SideChainStakingUpgrade) {
		sideChainId := k.GetSideChainId(ctx)
		sideChainCtx := ctx.WithSideChainKeyPrefix(k.GetSideChainStoreKeyPrefix(ctx, sideChainId))
		handleValidatorAndDelegations(sideChainCtx, k, logger)

		// TODO: save new validator set to ibc store
		// TODO: may need to change the return values
	}
	return
}

func handleValidatorAndDelegations(ctx sdk.Context, k keeper.Keeper, logger log.Logger) ([]abci.ValidatorUpdate, []types.UnbondingDelegation, sdk.Tags){
	endBlockerTags := sdk.EmptyTags()

	k.UnbondAllMatureValidatorQueue(ctx)
	completedUbd, tags := handleMatureUnbondingDelegations(k, ctx, logger)
	endBlockerTags.AppendTags(tags)

	tags = handleMatureRedelegations(k, ctx, logger)
	endBlockerTags.AppendTags(tags)

	// reset the intra-transaction counter
	k.SetIntraTxCounter(ctx, 0)

	// calculate validator set changes
	validatorUpdates := k.ApplyAndReturnValidatorSetUpdates(ctx)
	return validatorUpdates, completedUbd, endBlockerTags
}

func handleMatureRedelegations(k keeper.Keeper, ctx sdk.Context, logger log.Logger) (sdk.Tags) {
	endBlockerTags := sdk.EmptyTags()
	matureRedelegations := k.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockHeader().Time)
	for _, dvvTriplet := range matureRedelegations {
		err := k.CompleteRedelegation(ctx, dvvTriplet.DelegatorAddr, dvvTriplet.ValidatorSrcAddr, dvvTriplet.ValidatorDstAddr)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to complete redelegation: %s", err.Error()), "delegator_address", dvvTriplet.DelegatorAddr.String(), "source_validator_address", dvvTriplet.ValidatorSrcAddr.String(), "source_validator_address", dvvTriplet.ValidatorDstAddr.String())
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

func handleMatureUnbondingDelegations(k keeper.Keeper, ctx sdk.Context, logger log.Logger) ([]types.UnbondingDelegation, sdk.Tags) {
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

