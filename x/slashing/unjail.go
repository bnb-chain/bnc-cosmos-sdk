package slashing

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) Unjail(ctx sdk.Context, validatorAddr sdk.ValAddress) sdk.Error {

	validator := k.validatorSet.Validator(ctx, validatorAddr)
	if validator == nil {
		return ErrNoValidatorForAddress(k.codespace)
	}

	// cannot be unjailed if no self-delegation exists
	selfDel := k.validatorSet.Delegation(ctx, sdk.AccAddress(validator.GetFeeAddr()), validatorAddr)
	if selfDel == nil {
		return ErrMissingSelfDelegation(k.codespace)
	}

	if validator.TokensFromShares(selfDel.GetShares()).RawInt() < validator.GetMinSelfDelegation() {
		return ErrSelfDelegationTooLowToUnjail(k.codespace)
	}

	if !validator.GetJailed() {
		return ErrValidatorNotJailed(k.codespace)
	}

	var consAddr []byte
	if validator.IsSideChainValidator() {
		consAddr = validator.GetSideChainConsAddr()
	} else {
		consAddr = validator.GetConsAddr().Bytes()
	}
	info, found := k.getValidatorSigningInfo(ctx, consAddr)
	if found {
		// cannot be unjailed until out of jail
		if ctx.BlockHeader().Time.Before(info.JailedUntil) {
			return ErrValidatorJailed(k.codespace)
		}
	}

	// unjail the validator
	if validator.IsSideChainValidator() {
		k.validatorSet.UnjailSideChain(ctx, consAddr)
	} else {
		k.validatorSet.Unjail(ctx, consAddr)
	}

	return nil
}
