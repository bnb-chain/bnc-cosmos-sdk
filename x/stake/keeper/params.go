package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto"
)

// Default parameter namespace
const (
	DefaultParamspace = "stake"
)

var (
	DelegationAccAddr = sdk.AccAddress(crypto.AddressHash([]byte("BinanceChainStakeDelegation")))
)

// ParamTable for stake module
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable().RegisterParamSet(&types.Params{})
}

// TODO: 1. need to distinguish params for different chains.
// TODO: 2. SetParams for side chain in the BeginBlocker of the upgrade height

// UnbondingTime
func (k Keeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyUnbondingTime, &res)
	return
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyBondDenom, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (res types.Params) {
	res.UnbondingTime = k.UnbondingTime(ctx)
	res.MaxValidators = k.MaxValidators(ctx)
	res.BondDenom = k.BondDenom(ctx)
	return
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
