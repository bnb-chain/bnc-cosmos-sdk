package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func (k Keeper) GetWhiteLabelOracleRelayer(ctx sdk.Context) (oracleRelayers []types.OracleRelayer) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(WhiteLabelOracleRelayerKey)
	if bz == nil {
		panic("white label oracle relayer should not be nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &oracleRelayers)
	return oracleRelayers
}

func (k Keeper) SetWhiteLabelOracleRelayer(ctx sdk.Context, oracleRelayers []types.OracleRelayer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(oracleRelayers)
	store.Set(WhiteLabelOracleRelayerKey, b)
}

func (k Keeper) GetOracleRelayersPower(ctx sdk.Context) map[string]int64 {
	if sdk.IsUpgrade(sdk.BEP159Phase2) {
		return k.GetOracleRelayersPowerV1(ctx)
	} else {
		return k.GetOracleRelayersPowerV0(ctx)
	}
}

// get current validators and their vote power as oracle relayers and vote power in oracle module
func (k Keeper) GetOracleRelayersPowerV0(ctx sdk.Context) map[string]int64 {
	res := make(map[string]int64)
	validators := k.GetBondedValidatorsByPower(ctx)
	for _, validator := range validators {
		res[validator.OperatorAddr.String()] = validator.GetPower().RawInt()
	}
	return res
}

func (k Keeper) GetOracleRelayersPowerV1(ctx sdk.Context) map[string]int64 {
	res := make(map[string]int64)
	validators := k.GetWhiteLabelOracleRelayer(ctx)
	for _, validator := range validators {
		res[validator.Address.String()] = validator.Power
	}
	return res
}

func (k Keeper) CheckIsValidOracleRelayer(ctx sdk.Context, validatorAddress sdk.ValAddress) bool {
	if sdk.IsUpgrade(sdk.BEP159Phase2) {
		return k.CheckIsValidOracleRelayerV1(ctx, validatorAddress)
	} else {
		return k.CheckIsValidOracleRelayerV0(ctx, validatorAddress)
	}
}

func (k Keeper) CheckIsValidOracleRelayerV0(ctx sdk.Context, validatorAddress sdk.ValAddress) bool {
	validator, found := k.GetValidator(ctx, validatorAddress)
	if !found {
		return false
	}
	return validator.GetStatus().Equal(sdk.Bonded)
}

func (k Keeper) CheckIsValidOracleRelayerV1(ctx sdk.Context, validatorAddress sdk.ValAddress) bool {
	validators := k.GetWhiteLabelOracleRelayer(ctx)
	for _, validator := range validators {
		if validator.Address.Equals(validatorAddress) {
			return true
		}
	}
	return false
}
