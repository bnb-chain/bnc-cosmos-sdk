package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Implements ValidatorSet
var _ sdk.ValidatorSet = Keeper{}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidatorsBonded(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, LastValidatorPowerKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := AddressFromLastValidatorPowerKey(iterator.Key())
		validator, found := k.GetValidator(ctx, address)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", address))
		}

		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

// get the sdk.validator for a particular address
func (k Keeper) Validator(ctx sdk.Context, address sdk.ValAddress) sdk.Validator {
	val, found := k.GetValidator(ctx, address)
	if !found {
		return nil
	}
	return val
}

// get the sdk.validator for a particular pubkey
func (k Keeper) ValidatorByConsAddr(ctx sdk.Context, addr sdk.ConsAddress) sdk.Validator {
	val, found := k.GetValidatorByConsAddr(ctx, addr)
	if !found {
		return nil
	}
	return val
}

// get the sdk.validator for a particular vote address
func (k Keeper) ValidatorByVoteAddr(ctx sdk.Context, VoteAddress []byte) sdk.Validator {
	val, found := k.GetValidatorBySideVoteAddr(ctx, VoteAddress)
	if !found {
		return nil
	}
	return val
}

// get the sdk.validator for a particular consensus address
func (k Keeper) ValidatorBySideChainConsAddr(ctx sdk.Context, sideChainConsAddr []byte) sdk.Validator {
	val, found := k.GetValidatorBySideConsAddr(ctx, sideChainConsAddr)
	if !found {
		return nil
	}
	return val
}

// total power from the bond (not last, but current)
func (k Keeper) TotalPower(ctx sdk.Context) sdk.Dec {
	pool := k.GetPool(ctx)
	return pool.BondedTokens
}

// total power from the bond
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	pool := k.GetPool(ctx)
	return pool.BondedRatio()
}

// when minting new tokens
func (k Keeper) InflateSupply(ctx sdk.Context, newTokens sdk.Dec) {
	pool := k.GetPool(ctx)
	pool.LooseTokens = pool.LooseTokens.Add(newTokens)
	k.SetPool(ctx, pool)
}

//__________________________________________________________________________

// Implements DelegationSet

var _ sdk.DelegationSet = Keeper{}

// Returns self as it is both a validatorset and delegationset
func (k Keeper) GetValidatorSet() sdk.ValidatorSet {
	return k
}

// get the delegation for a particular set of delegator and validator addresses
func (k Keeper) Delegation(ctx sdk.Context, addrDel sdk.AccAddress, addrVal sdk.ValAddress) sdk.Delegation {
	bond, ok := k.GetDelegation(ctx, addrDel, addrVal)
	if !ok {
		return nil
	}

	return bond
}

// iterate through all of the delegations from a delegator
func (k Keeper) IterateDelegations(ctx sdk.Context, delAddr sdk.AccAddress,
	fn func(index int64, del sdk.Delegation) (stop bool)) {

	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delAddr)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest
	defer iterator.Close()
	for i := int64(0); iterator.Valid(); iterator.Next() {
		del := types.MustUnmarshalDelegation(k.cdc, iterator.Key(), iterator.Value())
		stop := fn(i, del)
		if stop {
			break
		}
		i++
	}
}

// iterate through all of the delegations to a validator
func (k Keeper) IterateDelegationsToValidator(ctx sdk.Context, valAddr sdk.ValAddress,
	fn func(del sdk.Delegation) (stop bool)) {

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegationKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		del := types.MustUnmarshalDelegation(k.cdc, iterator.Key(), iterator.Value())
		if !del.ValidatorAddr.Equals(valAddr) {
			continue
		}
		stop := fn(del)
		if stop {
			break
		}
	}
}
