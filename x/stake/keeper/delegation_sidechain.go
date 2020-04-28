package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Perform a delegation, set/update everything necessary within the store.
// Set delegation by the key grouped in the order of validator and delegator
func (k Keeper) DelegateForSideChain(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Coin,
	validator types.Validator, subtractAccount bool) (newShares sdk.Dec, err sdk.Error) {
	newShares, err = k.Delegate(ctx, delAddr, bondAmt, validator, subtractAccount)
	if err != nil {
		return
	}
	k.SyncDelegationByValDel(ctx, validator.OperatorAddr, delAddr)
	return newShares, err
}

// begin unbonding an unbonding record
// Set delegation with the key grouped in the order of validator and delegator
func (k Keeper) BeginUnbondingForSideChain(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (ubd types.UnbondingDelegation, err sdk.Error) {
	ubd, err = k.BeginUnbonding(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return
	}
	k.SyncDelegationByValDel(ctx, valAddr, delAddr)
	return ubd, err
}

func (k Keeper) BeginRedelegationForSideChain(ctx sdk.Context, delAddr sdk.AccAddress,
	valSrcAddr, valDstAddr sdk.ValAddress, sharesAmount sdk.Dec) (types.Redelegation, sdk.Error) {
	red, err := k.BeginRedelegation(ctx, delAddr, valSrcAddr, valDstAddr, sharesAmount)
	if err != nil {
		return red, err
	}
	k.SyncDelegationByValDel(ctx, valSrcAddr, delAddr)
	k.SyncDelegationByValDel(ctx, valDstAddr, delAddr)
	return red, err
}
