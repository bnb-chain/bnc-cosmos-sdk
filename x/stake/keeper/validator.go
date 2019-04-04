package keeper

import (
	"container/list"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Cache the amino decoding of validators, as it can be the case that repeated slashing calls
// cause many calls to GetValidator, which were shown to throttle the state machine in our
// simulation. Note this is quite biased though, as the simulator does more slashes than a
// live chain should.
type cachedValidator struct {
	val        types.Validator
	marshalled string // marshalled amino bytes for the validator object (not operator address)
}

// validatorCache-key: validator amino bytes
var validatorCache = make(map[string]cachedValidator, 500)
var validatorCacheList = list.New()

func (k Keeper) FixValidatorFeeAddr(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()

	allValidators := make([]types.Validator, 0, 11)
	for ; iterator.Valid(); iterator.Next() {
		addr := iterator.Key()[1:]
		// use the old logic to unmarshal it
		if validator, err := types.UnmarshalValidatorDeprecated(k.cdc, addr, iterator.Value()); err != nil {
			panic(err)
		} else {
			allValidators = append(allValidators, validator)
		}
	}

	// "tbnb13nj6strryvnqud5tqchkltwaatqr9awrxdlk8q",
	// "tbnb1lyjatx8jed40afe75t744hkvj0559xrvv4rle3",
	// "tbnb1snyg4ttdyckluwzphm4eh43uv5sw5ys5x9gxuj",
	// "tbnb1kem52fk9w43hemgqgjft8q76xesv0uypzcncjx",
	// "tbnb1hvcnlrflp2sgzvyrzqgtpsvqxrahhtrlsa4r4p",
	// "tbnb10da4lqdtp8yeahr2n2s88unmc2qn4czgnlr9u7",
	// "tbnb1xmdk54cuytnfgv6krzj2fnr9t55l94hnxdwe72",
	// "tbnb1f02de8sxjcznu5qejuzll0cmuxxzwq4yd4lghw",
	// "tbnb18arl8klkum0fpke8hyuucxu44wrhwdyg23nlsc",
	// "tbnb1zwuz8qklekm4vfwtkswu6nhhs2ghevn2dzqq85",
	// "tbnb1q7cc8e2nn39frppdzkqnzdy2f0pf664deg4zkq",
	feeAddressesHex := []string{
		"8CE5A82C6323260E368B062F6FADDDEAC032F5C3",
		"F925D598F2CB6AFEA73EA2FD5ADECC93E942986C",
		"84C88AAD6D262DFE3841BEEB9BD63C6520EA1214",
		"B6774526C575637CED004492B383DA3660C7F081",
		"BB313F8D3F0AA08130831010B0C18030FB7BAC7F",
		"7B7B5F81AB09C99EDC6A9AA073F27BC2813AE048",
		"36DB6A571C22E694335618A4A4CC655D29F2D6F3",
		"4BD4DC9E0696053E50199705FFBF1BE18C2702A4",
		"3F47F3DBF6E6DE90DB27B939CC1B95AB87773488",
		"13B82382DFCDB75625CBB41DCD4EF782917CB26A",
		"07B183E5539C4A91842D158131348A4BC29D6AAD",
	}

	for i, val := range allValidators {
		if i < 11 {
			aa, err := sdk.AccAddressFromHex(feeAddressesHex[i])
			if err != nil {
				panic(err)
			}
			// create the fee account
			if sdkErr := k.bankKeeper.SetCoins(ctx, aa, nil); sdkErr != nil {
				panic(sdkErr)
			}
			val.FeeAddr = aa
		} else {
			val.FeeAddr = sdk.AccAddress(val.OperatorAddr)
		}
		// use the new logic to Marshal and store it.
		k.SetValidator(ctx, val)
	}
}

// get a single validator
func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	value := store.Get(GetValidatorKey(addr))
	if value == nil {
		return validator, false
	}

	// If these amino encoded bytes are in the cache, return the cached validator
	strValue := string(value)
	if val, ok := validatorCache[strValue]; ok {
		valToReturn := val.val
		// Doesn't mutate the cache's value
		valToReturn.OperatorAddr = addr
		return valToReturn, true
	}

	// amino bytes weren't found in cache, so amino unmarshal and add it to the cache
	validator = types.MustUnmarshalValidator(k.cdc, addr, value)
	cachedVal := cachedValidator{validator, strValue}
	validatorCache[strValue] = cachedValidator{validator, strValue}
	validatorCacheList.PushBack(cachedVal)

	// if the cache is too big, pop off the last element from it
	if validatorCacheList.Len() > 500 {
		valToRemove := validatorCacheList.Remove(validatorCacheList.Front()).(cachedValidator)
		delete(validatorCache, valToRemove.marshalled)
	}

	validator = types.MustUnmarshalValidator(k.cdc, addr, value)
	return validator, true
}

func (k Keeper) mustGetValidator(ctx sdk.Context, addr sdk.ValAddress) types.Validator {
	validator, found := k.GetValidator(ctx, addr)
	if !found {
		panic(fmt.Sprintf("validator record not found for address: %X\n", addr))
	}
	return validator
}

// get a single validator by consensus address
func (k Keeper) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	opAddr := store.Get(GetValidatorByConsAddrKey(consAddr))
	if opAddr == nil {
		return validator, false
	}
	return k.GetValidator(ctx, opAddr)
}

func (k Keeper) mustGetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) types.Validator {
	validator, found := k.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		panic(fmt.Errorf("validator with consensus-Address %s not found", consAddr))
	}
	return validator
}

//___________________________________________________________________________

// set the main record holding validator details
func (k Keeper) SetValidator(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalValidator(k.cdc, validator)
	store.Set(GetValidatorKey(validator.OperatorAddr), bz)
}

// validator index
func (k Keeper) SetValidatorByConsAddr(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	consAddr := sdk.ConsAddress(validator.ConsPubKey.Address())
	store.Set(GetValidatorByConsAddrKey(consAddr), validator.OperatorAddr)
}

// validator index
func (k Keeper) SetValidatorByPowerIndex(ctx sdk.Context, validator types.Validator, pool types.Pool) {
	// jailed validators are not kept in the power index
	if validator.Jailed {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsByPowerIndexKey(validator, pool), validator.OperatorAddr)
}

// validator index
func (k Keeper) DeleteValidatorByPowerIndex(ctx sdk.Context, validator types.Validator, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetValidatorsByPowerIndexKey(validator, pool))
}

// validator index
func (k Keeper) SetNewValidatorByPowerIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	store.Set(GetValidatorsByPowerIndexKey(validator, pool), validator.OperatorAddr)
}

//___________________________________________________________________________

// Update the tokens of an existing validator, update the validators power index key
func (k Keeper) AddValidatorTokensAndShares(ctx sdk.Context, validator types.Validator,
	tokensToAdd int64) (valOut types.Validator, addedShares sdk.Dec) {

	pool := k.GetPool(ctx)
	k.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool, addedShares = validator.AddTokensFromDel(pool, tokensToAdd)
	// increment the intra-tx counter
	// in case of a conflict, the validator which least recently changed power takes precedence
	counter := k.GetIntraTxCounter(ctx)
	validator.BondIntraTxCounter = counter
	k.SetIntraTxCounter(ctx, counter+1)
	k.SetValidator(ctx, validator)
	k.SetPool(ctx, pool)
	k.SetValidatorByPowerIndex(ctx, validator, pool)
	return validator, addedShares
}

// Update the tokens of an existing validator, update the validators power index key
func (k Keeper) RemoveValidatorTokensAndShares(ctx sdk.Context, validator types.Validator,
	sharesToRemove sdk.Dec) (valOut types.Validator, removedTokens sdk.Dec) {

	pool := k.GetPool(ctx)
	k.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool, removedTokens = validator.RemoveDelShares(pool, sharesToRemove)
	k.SetValidator(ctx, validator)
	k.SetPool(ctx, pool)
	k.SetValidatorByPowerIndex(ctx, validator, pool)
	return validator, removedTokens
}

// Update the tokens of an existing validator, update the validators power index key
func (k Keeper) RemoveValidatorTokens(ctx sdk.Context, validator types.Validator, tokensToRemove sdk.Dec) types.Validator {
	pool := k.GetPool(ctx)
	k.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool = validator.RemoveTokens(pool, tokensToRemove)
	k.SetValidator(ctx, validator)
	k.SetPool(ctx, pool)
	k.SetValidatorByPowerIndex(ctx, validator, pool)
	return validator
}

// UpdateValidatorCommission attempts to update a validator's commission rate.
// An error is returned if the new commission rate is invalid.
func (k Keeper) UpdateValidatorCommission(ctx sdk.Context, validator types.Validator, newRate sdk.Dec) (types.Commission, sdk.Error) {
	commission := validator.Commission
	blockTime := ctx.BlockHeader().Time

	if err := commission.ValidateNewRate(newRate, blockTime); err != nil {
		return commission, err
	}

	commission.Rate = newRate
	commission.UpdateTime = blockTime

	return commission, nil
}

// remove the validator record and associated indexes
// except for the bonded validator index which is only handled in ApplyAndReturnTendermintUpdates
func (k Keeper) RemoveValidator(ctx sdk.Context, address sdk.ValAddress) {

	// first retrieve the old validator record
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return
	}

	// delete the old validator record
	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	store.Delete(GetValidatorKey(address))
	store.Delete(GetValidatorByConsAddrKey(sdk.ConsAddress(validator.ConsPubKey.Address())))
	store.Delete(GetValidatorsByPowerIndexKey(validator, pool))

}

//___________________________________________________________________________
// get groups of validators

// get the set of all validators with no limits, used during genesis dump
func (k Keeper) GetAllValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
		validators = append(validators, validator)
	}
	return validators
}

// return a given amount of all the validators
func (k Keeper) GetValidators(ctx sdk.Context, maxRetrieve uint16) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	validators = make([]types.Validator, maxRetrieve)

	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
		validators[i] = validator
		i++
	}
	return validators[:i] // trim if the array length < maxRetrieve
}

// get the group of the bonded validators
func (k Keeper) GetLastValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.MaxValidators(ctx)
	validators = make([]types.Validator, maxValidators)

	iterator := sdk.KVStorePrefixIterator(store, LastValidatorPowerKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid(); iterator.Next() {

		// sanity check
		if i >= int(maxValidators) {
			panic("more validators than maxValidators found")
		}
		address := AddressFromLastValidatorPowerKey(iterator.Key())
		validator := k.mustGetValidator(ctx, address)

		validators[i] = validator
		i++
	}
	return validators[:i] // trim
}

// get the current group of bonded validators sorted by power-rank
func (k Keeper) GetBondedValidatorsByPower(ctx sdk.Context) []types.Validator {
	store := ctx.KVStore(k.storeKey)
	maxValidators := k.MaxValidators(ctx)
	validators := make([]types.Validator, maxValidators)

	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
		address := iterator.Value()
		validator := k.mustGetValidator(ctx, address)

		if validator.Status == sdk.Bonded {
			validators[i] = validator
			i++
		}
	}
	return validators[:i] // trim
}

// gets a specific validator queue timeslice. A timeslice is a slice of ValAddresses corresponding to unbonding validators
// that expire at a certain time.
func (k Keeper) GetValidatorQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (valAddrs []sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GetValidatorQueueTimeKey(timestamp))
	if bz == nil {
		return []sdk.ValAddress{}
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &valAddrs)
	return valAddrs
}

// Sets a specific validator queue timeslice.
func (k Keeper) SetValidatorQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(keys)
	store.Set(GetValidatorQueueTimeKey(timestamp), bz)
}

// Insert an validator address to the appropriate timeslice in the validator queue
func (k Keeper) InsertValidatorQueue(ctx sdk.Context, val types.Validator) {
	timeSlice := k.GetValidatorQueueTimeSlice(ctx, val.UnbondingMinTime)
	if len(timeSlice) == 0 {
		k.SetValidatorQueueTimeSlice(ctx, val.UnbondingMinTime, []sdk.ValAddress{val.OperatorAddr})
	} else {
		timeSlice = append(timeSlice, val.OperatorAddr)
		k.SetValidatorQueueTimeSlice(ctx, val.UnbondingMinTime, timeSlice)
	}
}

// Returns all the validator queue timeslices from time 0 until endTime
func (k Keeper) ValidatorQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(ValidatorQueueKey, sdk.InclusiveEndBytes(GetValidatorQueueTimeKey(endTime)))
}

// Returns a concatenated list of all the timeslices before currTime, and deletes the timeslices from the queue
func (k Keeper) GetAllMatureValidatorQueue(ctx sdk.Context, currTime time.Time) (matureValsAddrs []sdk.ValAddress) {
	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	validatorTimesliceIterator := k.ValidatorQueueIterator(ctx, ctx.BlockHeader().Time)
	for ; validatorTimesliceIterator.Valid(); validatorTimesliceIterator.Next() {
		timeslice := []sdk.ValAddress{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(validatorTimesliceIterator.Value(), &timeslice)
		matureValsAddrs = append(matureValsAddrs, timeslice...)
	}
	return matureValsAddrs
}

// Unbonds all the unbonding validators that have finished their unbonding period
func (k Keeper) UnbondAllMatureValidatorQueue(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	validatorTimesliceIterator := k.ValidatorQueueIterator(ctx, ctx.BlockHeader().Time)
	for ; validatorTimesliceIterator.Valid(); validatorTimesliceIterator.Next() {
		timeslice := []sdk.ValAddress{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(validatorTimesliceIterator.Value(), &timeslice)
		for _, valAddr := range timeslice {
			val, found := k.GetValidator(ctx, valAddr)
			if !found || val.GetStatus() != sdk.Unbonding {
				continue
			}
			if val.GetDelegatorShares().IsZero() {
				k.RemoveValidator(ctx, val.OperatorAddr)
			} else {
				k.unbondingToUnbonded(ctx, val)
			}
		}
		store.Delete(validatorTimesliceIterator.Key())
	}
}
