package keeper

import (
	"fmt"

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

	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		validator.DistributionAddr = types.GenerateDistributionAddr(validator.OperatorAddr, types.ChainIDForBeaconChain)
		k.SetValidator(ctx, validator)
		delegation, found := k.GetDelegation(ctx, validator.FeeAddr, validator.OperatorAddr)
		if !found {
			panic(fmt.Sprintf("self delegation for %s not found", validator.OperatorAddr))
		}
		delegation.Height = ctx.BlockHeight()
		k.SetDelegation(ctx, delegation)
	}
}

func MigrateWhiteLabelOracleRelayer(ctx sdk.Context, k Keeper) {
	validators, _, found := k.GetHeightValidatorsByIndex(ctx, 1)
	if !found {
		panic("validators snapshot not found, should never happen")
	}
	var oracleRelayers []types.OracleRelayer
	for _, validator := range validators {
		oracleRelayers = append(oracleRelayers, types.OracleRelayer{
			Address: validator.OperatorAddr,
			Power:   validator.GetPower().RawInt(),
		})
	}
	k.SetWhiteLabelOracleRelayer(ctx, oracleRelayers)
}
