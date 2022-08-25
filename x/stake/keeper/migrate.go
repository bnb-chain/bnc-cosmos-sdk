package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func MigratePowerRankKey(ctx sdk.Context, keeper Keeper) {
	store := ctx.KVStore(keeper.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, ValidatorsByPowerIndexKey)
	defer iterator.Close()

	var validators []types.Validator
	for ; iterator.Valid(); iterator.Next() {
		valAddr := sdk.ValAddress(iterator.Value())
		validator, found := keeper.GetValidator(ctx, valAddr)
		if !found {
			keeper.Logger(ctx).Error("can't load validator", "operator_addr", valAddr.String())
			continue
		}
		validators = append(validators, validator)
		store.Delete(iterator.Key())
	}
	// Rebuild power rank key for validators
	for _, val := range validators {
		keeper.SetNewValidatorByPowerIndex(ctx, val)
	}
}

func MigrateValidators(ctx sdk.Context, k Keeper) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()

	bondDenom := k.BondDenom(ctx)
	var err error
	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		validator.DistributionAddr = types.GenerateDistributionAddr(validator.OperatorAddr, types.MockSideChainIDForBeaconChain)
		k.SetValidator(ctx, validator)
		_, err = k.Delegate(ctx, validator.FeeAddr, sdk.NewCoin(bondDenom, 0), validator, false)
		if err != nil {
			panic(err)
		}
	}
}
