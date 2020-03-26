package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func handleMsgCreateSideChainValidator(ctx sdk.Context, msg MsgCreateSideChainValidator, k keeper.Keeper) sdk.Result{
	storePrefix := k.GetSideChainStorePrefix(ctx, msg.SideChainId)
	if storePrefix == nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	}

	// add store prefix to ctx for side chain use
	ctx = ctx.WithSideChainKeyPrefix(storePrefix)

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorOwnerExists(k.Codespace()).Result()
	}

	_, found = k.GetValidatorBySideConsAddr(ctx, msg.SideConsAddr)
	if found {
		return ErrValidatorSideConsAddrExist(k.Codespace()).Result()
	}

	if msg.Delegation.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}

	// self-delegate address will be used to collect fees.
	feeAddr := msg.DelegatorAddr
	validator := NewSideChainValidator(feeAddr, msg.ValidatorAddr, msg.Description, msg.SideChainId, msg.SideConsAddr, msg.SideFeeAddr)
	commission := NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)
	validator, err := validator.SetInitialCommission(commission)
	if err != nil {
		return err.Result()
	}

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator) // here consAddr is the sideConsAddr
	k.SetNewValidatorByPowerIndex(ctx, validator)

	k.OnValidatorCreated(ctx, validator.OperatorAddr)

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	_, err = k.Delegate(ctx, msg.DelegatorAddr, msg.Delegation, validator, true)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.DstValidator, []byte(msg.ValidatorAddr.String()),
			tags.Moniker, []byte(msg.Description.Moniker),
			tags.Identity, []byte(msg.Description.Identity),
		),
	}
}

func handleMsgEditSideChainValidator(ctx sdk.Context, msg MsgEditSideChainValidator, k keeper.Keeper) sdk.Result {
	storePrefix := k.GetSideChainStorePrefix(ctx, msg.SideChainId)
	if storePrefix == nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	}

	// add store prefix to ctx for side chain use
	ctx = ctx.WithSideChainKeyPrefix(storePrefix)
	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return err.Result()
	}

	validator.Description = description

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return err.Result()
		}
		validator.Commission = commission
		k.OnValidatorModified(ctx, msg.ValidatorAddr)
	}

	if len(msg.SideConsAddr) != 0 {
		validator.SideConsAddr = msg.SideConsAddr
	}

	if len(msg.SideFeeAddr) != 0 {
		validator.SideFeeAddr = msg.SideFeeAddr
	}

	k.SetValidator(ctx, validator)
	return sdk.Result{
		Tags: sdk.NewTags(
			tags.DstValidator, []byte(msg.ValidatorAddr.String()),
			tags.Moniker, []byte(description.Moniker),
			tags.Identity, []byte(description.Identity),
		),
	}
}

func handleMsgSideChainDelegate(ctx sdk.Context, msg MsgSideChainDelegate, k keeper.Keeper) sdk.Result {
	storePrefix := k.GetSideChainStorePrefix(ctx, msg.SideChainId)
	if storePrefix == nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	}

	// add store prefix to ctx for side chain use
	ctx = ctx.WithSideChainKeyPrefix(storePrefix)
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	if msg.Delegation.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	// TODO: remove this check
	if validator.Jailed && !bytes.Equal(validator.OperatorAddr, msg.DelegatorAddr) {
		return ErrValidatorJailed(k.Codespace()).Result()
	}

	_, err := k.Delegate(ctx, msg.DelegatorAddr, msg.Delegation, validator, true)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.Delegator, []byte(msg.DelegatorAddr.String()),
			tags.DstValidator, []byte(msg.ValidatorAddr.String()),
		),
	}
}

func handleMsgSideChainRedelegate(ctx sdk.Context, msg MsgSideChainBeginRedelegate, k keeper.Keeper) sdk.Result {
	storePrefix := k.GetSideChainStorePrefix(ctx, msg.SideChainId)
	if storePrefix == nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	}

	// add store prefix to ctx for side chain use
	ctx = ctx.WithSideChainKeyPrefix(storePrefix)
	if msg.Amount.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	shares ,err := k.ValidateUnbondAmount(ctx, msg.DelegatorAddr, msg.ValidatorSrcAddr, msg.Amount.Amount)
	if err != nil {
		return err.Result()
	}

	red, err := k.BeginRedelegation(ctx, msg.DelegatorAddr, msg.ValidatorSrcAddr,
		msg.ValidatorDstAddr, shares)
	if err != nil {
		return err.Result()
	}

	finishTime := types.MsgCdc.MustMarshalBinaryLengthPrefixed(red.MinTime)

	tags := sdk.NewTags(
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorSrcAddr.String()),
		tags.DstValidator, []byte(msg.ValidatorDstAddr.String()),
		tags.EndTime, finishTime,
	)
	return sdk.Result{Data: finishTime, Tags: tags}
}

func handleMsgSideChainUndelegate(ctx sdk.Context, msg MsgSideChainUndelegate, k keeper.Keeper) sdk.Result {
	storePrefix := k.GetSideChainStorePrefix(ctx, msg.SideChainId)
	if storePrefix == nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	}

	// add store prefix to ctx for side chain use
	ctx = ctx.WithSideChainKeyPrefix(storePrefix)
	if msg.Amount.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	shares ,err := k.ValidateUnbondAmount(ctx, msg.DelegatorAddr, msg.ValidatorAddr, msg.Amount.Amount)
	if err != nil {
		return err.Result()
	}

	ubd, err := k.BeginUnbonding(ctx, msg.DelegatorAddr, msg.ValidatorAddr, shares)
	if err != nil {
		return err.Result()
	}

	finishTime := types.MsgCdc.MustMarshalBinaryLengthPrefixed(ubd.MinTime)

	tags := sdk.NewTags(
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorAddr.String()),
		tags.EndTime, finishTime,
	)
	return sdk.Result{Data: finishTime, Tags: tags}
}