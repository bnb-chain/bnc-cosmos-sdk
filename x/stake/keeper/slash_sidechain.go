package keeper

import (
	"encoding/hex"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func (k Keeper) SlashSideChain(ctx sdk.Context, sideChainId string, sideConsAddr []byte, slashAmount sdk.Dec, submitterReward sdk.Dec, submitter sdk.AccAddress, feePoolAdd sdk.Dec) (sdk.Dec, sdk.Coin, error) {
	logger := ctx.Logger().With("module", "x/stake")

	storePrefix := k.ScKeeper.GetSideChainStorePrefix(ctx, sideChainId)
	if storePrefix == nil {
		return sdk.ZeroDec(), sdk.Coin{}, errors.New("invalid side chain id")
	}

	// add store prefix to ctx for side chain use
	sideCtx := ctx.WithSideChainKeyPrefix(storePrefix)

	validator, found := k.GetValidatorBySideConsAddr(sideCtx, sideConsAddr)
	if !found {
		// If not found, the validator must have been overslashed and removed - so we don't need to do anything
		// NOTE:  Correctness dependent on invariant that unbonding delegations / redelegations must also have been completely
		//        slashed in this case - which we don't explicitly check, but should be true.
		// Log the slash attempt for future reference (maybe we should tag it too)
		logger.Error(fmt.Sprintf(
			"WARNING: Ignored attempt to slash a nonexistent validator with address %s, we recommend you investigate immediately",
			sdk.HexEncode(sideConsAddr)))
		return sdk.ZeroDec(), sdk.Coin{}, nil
	}

	// should not be slashing unbonded
	if validator.Status == sdk.Unbonded {
		return sdk.ZeroDec(), sdk.Coin{}, errors.New(fmt.Sprintf("should not be slashing unbonded validator: %s", validator.GetOperator()))
	}

	bondDenom := k.BondDenom(sideCtx)

	if !validator.Jailed {
		k.JailSideChain(sideCtx, sideConsAddr)
	}

	selfDelegation, found := k.GetDelegation(sideCtx, validator.FeeAddr, validator.OperatorAddr)
	remainingSlashAmount := slashAmount
	if found {
		slashSelfDelegationShares := sdk.MinDec(slashAmount, selfDelegation.Shares)
		_, err := k.unbond(sideCtx, selfDelegation.DelegatorAddr, validator.OperatorAddr, slashSelfDelegationShares)
		if err != nil {
			return sdk.ZeroDec(), sdk.Coin{}, errors.New(fmt.Sprintf("error unbonding delegator: %v", err))
		}
		remainingSlashAmount = remainingSlashAmount.Sub(slashSelfDelegationShares)
	}

	if !remainingSlashAmount.IsZero() {
		ubd, found := k.GetUnbondingDelegation(sideCtx, validator.FeeAddr, validator.OperatorAddr)
		if found {
			slashUnBondingAmount := sdk.MinInt64(remainingSlashAmount.RawInt(), ubd.Balance.Amount)
			ubd.Balance.Amount = ubd.Balance.Amount - slashUnBondingAmount
			k.SetUnbondingDelegation(sideCtx, ubd)
			remainingSlashAmount = remainingSlashAmount.Sub(sdk.NewDec(slashUnBondingAmount))
		}
	}

	slashedAmt := slashAmount.Sub(remainingSlashAmount)
	submitterReward = sdk.MinDec(submitterReward, slashedAmt)
	if !submitterReward.IsZero() && submitter != nil {
		if _, err := k.bankKeeper.SendCoins(sideCtx, DelegationAccAddr, submitter, sdk.Coins{sdk.NewCoin(bondDenom, submitterReward.RawInt())}); err != nil {
			return sdk.ZeroDec(), sdk.Coin{}, err
		}
	}

	remainingReward := slashedAmt.Sub(submitterReward)

	var feeCoinAdd sdk.Coin
	if !remainingReward.IsZero() {
		if !feePoolAdd.IsZero() {
			feePoolAdd = sdk.MinDec(feePoolAdd, remainingReward)
			feeCoinAdd = sdk.NewCoin(k.BondDenom(sideCtx), feePoolAdd.RawInt())
			remainingReward = remainingReward.Sub(feePoolAdd)
		}

		if !remainingReward.IsZero() {
			// allocate remaining rewards to other validators
			validators, _, found := k.GetHeightValidatorsByIndex(sideCtx, 1)
			if !found {
				return sdk.ZeroDec(), sdk.Coin{}, errors.New("can not found validators of current day")
			}

			// remove bad validator if it exists in the eligible validators for current day
			for i := 0; i < len(validators); i++ {
				if validators[i].OperatorAddr.Equals(validator.OperatorAddr) {
					if i == len(validators)-1 {
						validators = validators[:i]
					} else {
						validators = append(validators[:i], validators[i+1:]...)
					}
					break
				}
			}

			sharers, totalShares := convertValidators2Shares(validators)
			rewards := allocate(sharers, remainingReward, totalShares)
			for i := range rewards {
				if _, err := k.bankKeeper.SendCoins(sideCtx, DelegationAccAddr, rewards[i].AccAddr, sdk.Coins{sdk.NewCoin(bondDenom, rewards[i].Amount)}); err != nil {
					return sdk.ZeroDec(), sdk.Coin{}, err
				}
			}
		}
	}

	if validator.Status == sdk.Bonded {
		ibcValidator := types.IbcValidator{
			ConsAddr: validator.SideConsAddr,
			FeeAddr:  validator.SideFeeAddr,
			DistAddr: validator.DistributionAddr,
			Power:    validator.GetPower().RawInt(),
		}
		if _, err := k.SaveJailedValidatorToIbc(sideCtx, sideChainId, ibcValidator); err != nil {
			return sdk.ZeroDec(), sdk.Coin{}, errors.New(err.Error())
		}
	}

	return slashedAmt, feeCoinAdd, nil

}

// jail a validator
func (k Keeper) JailSideChain(ctx sdk.Context, consAddr []byte) {
	validator := k.mustGetValidatorBySideConsAddr(ctx, consAddr)
	k.jailValidator(ctx, validator)
	k.Logger(ctx).Info(fmt.Sprintf("validator %s jailed", hex.EncodeToString(consAddr)))
	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// unjail a validator
func (k Keeper) UnjailSideChain(ctx sdk.Context, consAddr []byte) {
	validator := k.mustGetValidatorBySideConsAddr(ctx, consAddr)
	k.unjailValidator(ctx, validator)
	k.Logger(ctx).Info(fmt.Sprintf("validator %s unjailed", hex.EncodeToString(consAddr)))
	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

func convertValidators2Shares(validators []types.Validator) (sharers []Sharer, totalShares sdk.Dec) {
	sharers = make([]Sharer, len(validators))
	for i, val := range validators {
		sharers[i] = Sharer{AccAddr: val.DistributionAddr, Shares: val.DelegatorShares}
		totalShares = totalShares.Add(val.DelegatorShares)
	}
	return sharers, totalShares
}
