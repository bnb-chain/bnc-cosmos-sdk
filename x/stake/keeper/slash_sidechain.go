package keeper

import (
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func (k Keeper) SlashForSideChain(ctx sdk.Context, sideChainId string, consAddr sdk.ConsAddress, slashAmount sdk.Dec, submitterReward sdk.Dec, submitter sdk.AccAddress) error {
	logger := ctx.Logger().With("module", "x/stake")

	storePrefix := k.GetSideChainStorePrefix(ctx, sideChainId)
	if storePrefix == nil {
		return errors.New("invalid side chain id")
	}

	// add store prefix to ctx for side chain use
	ctx = ctx.WithSideChainKeyPrefix(storePrefix)

	validator, found := k.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		// If not found, the validator must have been overslashed and removed - so we don't need to do anything
		// NOTE:  Correctness dependent on invariant that unbonding delegations / redelegations must also have been completely
		//        slashed in this case - which we don't explicitly check, but should be true.
		// Log the slash attempt for future reference (maybe we should tag it too)
		logger.Error(fmt.Sprintf(
			"WARNING: Ignored attempt to slash a nonexistent validator with address %s, we recommend you investigate immediately",
			consAddr))
		return nil
	}

	// should not be slashing unbonded
	if validator.Status == sdk.Unbonded {
		return errors.New(fmt.Sprintf("should not be slashing unbonded validator: %s", validator.GetOperator()))
	}

	bondDenom := k.BondDenom(ctx)

	if !validator.Jailed {
		k.Jail(ctx, consAddr)
	}

	selfDelegation, found := k.GetDelegation(ctx, validator.FeeAddr, validator.OperatorAddr)
	remainingSlashAmount := slashAmount
	if found {
		slashSelfDelegationShares := sdk.MinDec(slashAmount, selfDelegation.Shares)
		_, err := k.unbond(ctx, selfDelegation.DelegatorAddr, validator.OperatorAddr, slashSelfDelegationShares)
		if err != nil {
			return errors.New(fmt.Sprintf("error unbonding delegator: %v", err))
		}
		remainingSlashAmount = remainingSlashAmount.Sub(slashSelfDelegationShares)
	}

	if !remainingSlashAmount.IsZero() {
		ubd, found := k.GetUnbondingDelegation(ctx, validator.FeeAddr, validator.OperatorAddr)
		if found {
			slashUnBondingAmount := sdk.MinInt64(remainingSlashAmount.RawInt(), ubd.Balance.Amount)
			ubd.Balance.Amount = ubd.Balance.Amount - slashUnBondingAmount
			k.SetUnbondingDelegation(ctx, ubd)
			remainingSlashAmount = remainingSlashAmount.Sub(sdk.NewDec(slashUnBondingAmount))
		}
		if !remainingSlashAmount.IsZero() {
			reds := k.GetRedelegationsFromValidator(ctx, validator.OperatorAddr)
			if reds != nil && len(reds) > 0 {
				for _, red := range reds {
					if red.DelegatorAddr.Equals(selfDelegation.DelegatorAddr) {
						slashReDelegationAmount := sdk.MinInt64(remainingSlashAmount.RawInt(), red.Balance.Amount)
						if slashReDelegationAmount > 0 {
							red.Balance.Amount = red.Balance.Amount - slashReDelegationAmount
							red.SharesSrc = red.SharesSrc.Sub(sdk.NewDec(slashReDelegationAmount))
							red.SharesDst = red.SharesDst.Sub(sdk.NewDec(slashReDelegationAmount))
							k.SetRedelegation(ctx, red)
						}

						delegation, found := k.GetDelegation(ctx, red.DelegatorAddr, red.ValidatorDstAddr)
						if !found {
							// If deleted, delegation has zero shares, and we can't unbond any more
							continue
						}
						slashDelegationShares := sdk.MinDec(sdk.NewDec(slashReDelegationAmount), delegation.Shares)
						_, err := k.unbond(ctx, red.DelegatorAddr, red.ValidatorDstAddr, slashDelegationShares)
						if err != nil {
							panic(fmt.Errorf("error unbonding delegator: %v", err))
						}
						remainingSlashAmount = remainingSlashAmount.Sub(slashDelegationShares)
						if remainingSlashAmount.IsZero() {
							break
						}
					}
				}
			}
		}
	}

	slashedAmt := slashAmount.Sub(remainingSlashAmount)
	submitterReward = sdk.MinDec(submitterReward, slashedAmt)
	remainingReward := slashedAmt.Sub(submitterReward)
	if _, err := k.bankKeeper.SendCoins(ctx, DelegationAccAddr, submitter, sdk.Coins{sdk.NewCoin(bondDenom, submitterReward.RawInt())}); err != nil {
		return err
	}
	// allocate remaining rewards to other validators
	height, found := k.GetBreatheBlockHeight(ctx, 1)
	if !found {
		return errors.New("can not found breathe block height of current day")
	}
	validators, found := k.GetValidatorsByHeight(ctx, height)
	if !found {
		return errors.New("can not found validators of current day")
	}

	if !remainingReward.IsZero() {
		for i := 0; i < len(validators); i++ {
			if validators[i].OperatorAddr.Equals(validator.OperatorAddr) {
				validators = append(validators[:i], validators[i+1:]...)
				break
			}
		}

		sharers, totalShares := convertValidators2Shares(validators)
		shouldCarry, shouldNotCarry, _ := allocate(sharers, remainingReward, totalShares, 0)
		rewards := append(shouldCarry, shouldNotCarry...)
		for _, eachReward := range rewards {
			if _, err := k.bankKeeper.SendCoins(ctx, DelegationAccAddr, eachReward.AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, eachReward.Reward)}); err != nil {
				return err
			}
		}

	}

	if validator.Status == sdk.Bonded { // todo send package to BSC

	}

	return nil

}

func convertValidators2Shares(validators []types.Validator) (sharers []Sharer, totalShares sdk.Dec) {
	sharers = make([]Sharer, len(validators))
	for i, val := range validators {
		sharers[i] = Sharer{AccAddr: val.DistributionAddr, Shares: val.DelegatorShares}
		totalShares = totalShares.Add(val.DelegatorShares)
	}
	return sharers, totalShares
}
