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

func (k Keeper) MinSelfDelegation(ctx sdk.Context) (res int64) {
	k.paramstore.GetIfExists(ctx, types.KeyMinSelfDelegation, &res)
	return
}

func (k Keeper) MinDelegationChange(ctx sdk.Context) (res int64) {
	k.paramstore.GetIfExists(ctx, types.KeyMinDelegationChange, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (res types.Params) {
	res.UnbondingTime = k.UnbondingTime(ctx)
	res.MaxValidators = k.MaxValidators(ctx)
	res.BondDenom = k.BondDenom(ctx)
	res.MinSelfDelegation = k.MinSelfDelegation(ctx)
	res.MinDelegationChange = k.MinDelegationChange(ctx)
	return
}

// in order to be compatible with before
type paramBeforeBscUpgrade struct {
	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

// Implements params.ParamSet
func (p *paramBeforeBscUpgrade) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{types.KeyUnbondingTime, &p.UnbondingTime},
		{types.KeyMaxValidators, &p.MaxValidators},
		{types.KeyBondDenom, &p.BondDenom},
	}
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	sdk.Upgrade(sdk.LaunchBscUpgrade, func() {
		var pb paramBeforeBscUpgrade
		pb.UnbondingTime = params.UnbondingTime
		pb.MaxValidators = params.MaxValidators
		pb.BondDenom = params.BondDenom

		k.paramstore.SetParamSet(ctx, &pb)
	}, nil, func() {
		k.paramstore.SetParamSet(ctx, &params)
	})

}
