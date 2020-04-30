package keeper

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func (k Keeper) SlashSideChain(ctx sdk.Context, sideChainId string, sideConsAddr []byte, slashAmount sdk.Dec) (sdk.Dec, error) {
	logger := ctx.Logger().With("module", "x/stake")

	sideCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId)
	if err != nil {
		return sdk.ZeroDec(), errors.New("invalid side chain id")
	}

	validator, found := k.GetValidatorBySideConsAddr(sideCtx, sideConsAddr)
	if !found {
		// If not found, the validator must have been overslashed and removed - so we don't need to do anything
		// NOTE:  Correctness dependent on invariant that unbonding delegations / redelegations must also have been completely
		//        slashed in this case - which we don't explicitly check, but should be true.
		// Log the slash attempt for future reference (maybe we should tag it too)
		logger.Error(fmt.Sprintf(
			"WARNING: Ignored attempt to slash a nonexistent validator with address %s, we recommend you investigate immediately",
			sdk.HexEncode(sideConsAddr)))
		return sdk.ZeroDec(), nil
	}

	// should not be slashing unbonded
	if validator.IsUnbonded() {
		return sdk.ZeroDec(), errors.New(fmt.Sprintf("should not be slashing unbonded validator: %s", validator.GetOperator()))
	}

	if !validator.Jailed {
		k.JailSideChain(sideCtx, sideConsAddr)
	}

	selfDelegation, found := k.GetDelegation(sideCtx, validator.FeeAddr, validator.OperatorAddr)
	remainingSlashAmount := slashAmount
	if found {
		slashShares := validator.SharesFromTokens(slashAmount)
		slashSelfDelegationShares := sdk.MinDec(slashShares, selfDelegation.Shares)
		if slashSelfDelegationShares.RawInt() > 0 {
			unbondAmount, err := k.unbond(sideCtx, selfDelegation.DelegatorAddr, validator.OperatorAddr, slashSelfDelegationShares)
			if err != nil {
				return sdk.ZeroDec(), errors.New(fmt.Sprintf("error unbonding delegator: %v", err))
			}
			remainingSlashAmount = remainingSlashAmount.Sub(unbondAmount)
		}
	}

	if remainingSlashAmount.RawInt() > 0 {
		ubd, found := k.GetUnbondingDelegation(sideCtx, validator.FeeAddr, validator.OperatorAddr)
		if found {
			slashUnBondingAmount := sdk.MinInt64(remainingSlashAmount.RawInt(), ubd.Balance.Amount)
			ubd.Balance.Amount = ubd.Balance.Amount - slashUnBondingAmount
			k.SetUnbondingDelegation(sideCtx, ubd)
			remainingSlashAmount = remainingSlashAmount.Sub(sdk.NewDec(slashUnBondingAmount))
		}
	}

	slashedAmt := slashAmount.Sub(remainingSlashAmount)

	bondDenom := k.BondDenom(ctx)
	delegationAccBalance := k.bankKeeper.GetCoins(ctx, DelegationAccAddr)
	slashedCoin := sdk.NewCoin(bondDenom, slashedAmt.RawInt())
	if err := k.bankKeeper.SetCoins(ctx, DelegationAccAddr, delegationAccBalance.Minus(sdk.Coins{slashedCoin})); err != nil {
		return slashedAmt, err
	}
	if k.addrPool != nil {
		k.addrPool.AddAddrs([]sdk.AccAddress{DelegationAccAddr})
	}
	if validator.IsBonded() {
		ibcValidator := types.IbcValidator{
			ConsAddr: validator.SideConsAddr,
			FeeAddr:  validator.SideFeeAddr,
			DistAddr: validator.DistributionAddr,
			Power:    validator.GetPower().RawInt(),
		}
		if _, err := k.SaveJailedValidatorToIbc(sideCtx, sideChainId, ibcValidator); err != nil {
			return sdk.ZeroDec(), errors.New(err.Error())
		}
	}

	return slashedAmt, nil

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

func (k Keeper) AllocateSlashAmtToValidators(ctx sdk.Context, slashedConsAddr []byte, amount sdk.Dec) (bool, error) {
	// allocate remaining rewards to validators who are going to be distributed next time.
	validators, found := k.GetEarliestValidatorsWithHeight(ctx)
	if !found {
		return found, nil
	}
	// remove bad validator if it exists in the eligible validators
	for i := 0; i < len(validators); i++ {
		if bytes.Compare(validators[i].SideConsAddr, slashedConsAddr) == 0 {
			if i == len(validators)-1 {
				validators = validators[:i]
			} else {
				validators = append(validators[:i], validators[i+1:]...)
			}
			break
		}
	}

	bondDenom := k.BondDenom(ctx)
	sharers, totalShares := convertValidators2Shares(validators)
	rewards := allocate(sharers, amount, totalShares)

	changedAddrs := make([]sdk.AccAddress, len(rewards))
	for i := range rewards {
		accBalance := k.bankKeeper.GetCoins(ctx, rewards[i].AccAddr)
		rewardCoin := sdk.Coins{sdk.NewCoin(bondDenom, rewards[i].Amount)}
		accBalance.Plus(rewardCoin)
		if err := k.bankKeeper.SetCoins(ctx, rewards[i].AccAddr, accBalance.Plus(rewardCoin)); err != nil {
			return found, err
		}
		changedAddrs[i] = rewards[i].AccAddr
	}
	if k.addrPool != nil {
		k.addrPool.AddAddrs(changedAddrs)
	}
	return found, nil
}
