package stake

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/bsc"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func handleMsgCreateSideChainValidator(ctx sdk.Context, msg MsgCreateSideChainValidator, k keeper.Keeper) sdk.Result {
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}
	// check to see if BEP126 has upgraded
	if sdk.IsUpgrade(sdk.BEP126) {
		return types.ErrNilValidatorSideVoteAddr(k.Codespace()).Result()
	}

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorOwnerExists(k.Codespace()).Result()
	}

	_, found = k.GetValidatorBySideConsAddr(ctx, msg.SideConsAddr)
	if found {
		return ErrValidatorSideConsAddrExist(k.Codespace()).Result()
	}

	minSelfDelegation := k.MinSelfDelegation(ctx)
	if msg.Delegation.Amount < minSelfDelegation {
		return ErrBadDelegationAmount(DefaultCodespace,
			fmt.Sprintf("self delegation must not be less than %d", minSelfDelegation)).Result()
	}
	if msg.Delegation.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}

	// self-delegate address will be used to collect fees.
	feeAddr := msg.DelegatorAddr
	validator := NewSideChainValidator(feeAddr, msg.ValidatorAddr, msg.Description, msg.SideChainId, msg.SideConsAddr, msg.SideFeeAddr, nil)
	commission := NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)
	var err sdk.Error
	validator, err = validator.SetInitialCommission(commission)
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
	if len(msg.SideConsAddr) != 0 {
		if !sdk.IsUpgrade(sdk.BEP159) {
			return ErrEditConsensusKeyBeforeBEP159(k.Codespace()).Result()
		}
	}
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}

	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	// replace all editable fields (clients should autofill existing values)
	if description, err := validator.Description.UpdateDescription(msg.Description); err != nil {
		return err.Result()
	} else {
		validator.Description = description
	}

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return err.Result()
		}
		validator.Commission = commission
		k.OnValidatorModified(ctx, msg.ValidatorAddr)
	}

	if len(msg.SideFeeAddr) != 0 {
		validator.SideFeeAddr = msg.SideFeeAddr
	}

	if len(msg.SideConsAddr) != 0 && sdk.IsUpgrade(sdk.BEP159) {
		_, found = k.GetValidatorBySideConsAddr(ctx, msg.SideConsAddr)
		if found {
			return ErrValidatorSideConsAddrExist(k.Codespace()).Result()
		}
		if sdk.IsUpgrade(sdk.LimitConsAddrUpdateInterval) {
			// check update sideConsAddr interval
			latestUpdateConsAddrTime, err := k.GetValLatestUpdateConsAddrTime(ctx, validator.OperatorAddr)
			if err != nil {
				return sdk.ErrInternal(fmt.Sprintf("failed to get latest update cons addr time: %s", err)).Result()
			}
			if ctx.BlockHeader().Time.Sub(latestUpdateConsAddrTime).Hours() < types.ConsAddrUpdateIntervalInHours {
				return types.ErrConsAddrUpdateTime().Result()
			}
			k.SetValLatestUpdateConsAddrTime(ctx, validator.OperatorAddr, ctx.BlockHeader().Time)
		}
		// here consAddr is the sideConsAddr
		k.UpdateSideValidatorConsAddr(ctx, validator, msg.SideConsAddr)
		validator.SideConsAddr = msg.SideConsAddr
	}

	k.SetValidator(ctx, validator)
	return sdk.Result{
		Tags: sdk.NewTags(
			tags.DstValidator, []byte(msg.ValidatorAddr.String()),
			tags.Moniker, []byte(validator.Description.Moniker),
			tags.Identity, []byte(validator.Description.Identity),
		),
	}
}

func handleMsgCreateSideChainValidatorWithVoteAddr(ctx sdk.Context, msg MsgCreateSideChainValidatorWithVoteAddr, k keeper.Keeper) sdk.Result {
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}

	// check to see if BEP126 has upgraded
	if !sdk.IsUpgrade(sdk.BEP126) {
		return types.ErrBadValidatorSideVoteAddr(k.Codespace()).Result()
	}

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorOwnerExists(k.Codespace()).Result()
	}

	_, found = k.GetValidatorBySideConsAddr(ctx, msg.SideConsAddr)
	if found {
		return ErrValidatorSideConsAddrExist(k.Codespace()).Result()
	}

	msg.SideVoteAddr = msg.SideVoteAddr[:sdk.VoteAddrLen]
	_, found = k.GetValidatorBySideVoteAddr(ctx, msg.SideVoteAddr)
	if found {
		return ErrValidatorSideVoteAddrExist(k.Codespace()).Result()
	}

	minSelfDelegation := k.MinSelfDelegation(ctx)
	if msg.Delegation.Amount < minSelfDelegation {
		return ErrBadDelegationAmount(DefaultCodespace,
			fmt.Sprintf("self delegation must not be less than %d", minSelfDelegation)).Result()
	}
	if msg.Delegation.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadDenom(k.Codespace()).Result()
	}

	// self-delegate address will be used to collect fees.
	feeAddr := msg.DelegatorAddr
	validator := NewSideChainValidator(feeAddr, msg.ValidatorAddr, msg.Description, msg.SideChainId, msg.SideConsAddr, msg.SideFeeAddr, msg.SideVoteAddr)
	commission := NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)
	var err sdk.Error
	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return err.Result()
	}

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator) // here consAddr is the sideConsAddr
	k.SetValidatorBySideVoteAddr(ctx, validator)
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

func handleMsgEditSideChainValidatorWithVoteAddr(ctx sdk.Context, msg MsgEditSideChainValidatorWithVoteAddr, k keeper.Keeper) sdk.Result {
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}

	// check to see if BEP126 has upgraded
	if !sdk.IsUpgrade(sdk.BEP126) {
		return types.ErrBadValidatorSideVoteAddr(k.Codespace()).Result()
	}

	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	// replace all editable fields (clients should autofill existing values)
	if description, err := validator.Description.UpdateDescription(msg.Description); err != nil {
		return err.Result()
	} else {
		validator.Description = description
	}

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return err.Result()
		}
		validator.Commission = commission
		k.OnValidatorModified(ctx, msg.ValidatorAddr)
	}

	if len(msg.SideFeeAddr) != 0 {
		validator.SideFeeAddr = msg.SideFeeAddr
	}

	if len(msg.SideConsAddr) != 0 && sdk.IsUpgrade(sdk.BEP159) {
		_, found = k.GetValidatorBySideConsAddr(ctx, msg.SideConsAddr)
		if found {
			return ErrValidatorSideConsAddrExist(k.Codespace()).Result()
		}
		if sdk.IsUpgrade(sdk.LimitConsAddrUpdateInterval) {
			// check update sideConsAddr interval
			latestUpdateConsAddrTime, err := k.GetValLatestUpdateConsAddrTime(ctx, validator.OperatorAddr)
			if err != nil {
				return sdk.ErrInternal(fmt.Sprintf("failed to get latest update cons addr time: %s", err)).Result()
			}
			if ctx.BlockHeader().Time.Sub(latestUpdateConsAddrTime).Hours() < types.ConsAddrUpdateIntervalInHours {
				return types.ErrConsAddrUpdateTime().Result()
			}
			k.SetValLatestUpdateConsAddrTime(ctx, validator.OperatorAddr, ctx.BlockHeader().Time)
		}
		// here consAddr is the sideConsAddr
		k.UpdateSideValidatorConsAddr(ctx, validator, msg.SideConsAddr)
		validator.SideConsAddr = msg.SideConsAddr
	}

	if len(msg.SideVoteAddr) != 0 {
		msg.SideVoteAddr = msg.SideVoteAddr[:sdk.VoteAddrLen]
		_, found = k.GetValidatorBySideVoteAddr(ctx, msg.SideVoteAddr)
		if found {
			return ErrValidatorSideVoteAddrExist(k.Codespace()).Result()
		}
		validator.SideVoteAddr = msg.SideVoteAddr
		k.SetValidatorBySideVoteAddr(ctx, validator)
	}

	k.SetValidator(ctx, validator)
	return sdk.Result{
		Tags: sdk.NewTags(
			tags.DstValidator, []byte(msg.ValidatorAddr.String()),
			tags.Moniker, []byte(validator.Description.Moniker),
			tags.Identity, []byte(validator.Description.Identity),
		),
	}
}

func handleMsgSideChainDelegate(ctx sdk.Context, msg MsgSideChainDelegate, k keeper.Keeper) sdk.Result {
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}

	// we need this lower limit to prevent too many delegation records.
	minDelegationChange := k.MinDelegationChange(ctx)
	if msg.Delegation.Amount < minDelegationChange {
		return ErrBadDelegationAmount(DefaultCodespace, fmt.Sprintf("delegation must not be less than %d", minDelegationChange)).Result()
	}

	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorFound(k.Codespace()).Result()
	}

	if msg.Delegation.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	if err := checkOperatorAsDelegator(k, msg.DelegatorAddr, validator); err != nil {
		return err.Result()
	}

	// if the validator is jailed, only the self-delegator can delegate to itself
	if validator.Jailed && !bytes.Equal(validator.FeeAddr, msg.DelegatorAddr) {
		return ErrValidatorJailed(k.Codespace()).Result()
	}

	_, err := k.Delegate(ctx, msg.DelegatorAddr, msg.Delegation, validator, true)
	if err != nil {
		return err.Result()
	}

	// publish delegate event
	if k.PbsbServer != nil && ctx.IsDeliverTx() {
		event := types.ChainDelegateEvent{
			DelegateEvent: types.DelegateEvent{
				StakeEvent: types.StakeEvent{
					IsFromTx: true,
				},
				Delegator: msg.DelegatorAddr,
				Validator: msg.ValidatorAddr,
				Amount:    msg.Delegation.Amount,
				Denom:     msg.Delegation.Denom,
				TxHash:    ctx.Value(baseapp.TxHashKey).(string),
			},
			ChainId: msg.SideChainId,
		}
		k.PbsbServer.Publish(event)
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.Delegator, []byte(msg.DelegatorAddr.String()),
			tags.DstValidator, []byte(msg.ValidatorAddr.String()),
		),
	}
}

func handleMsgSideChainRedelegate(ctx sdk.Context, msg MsgSideChainRedelegate, k keeper.Keeper) sdk.Result {
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}

	if msg.Amount.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	dstValidator, found := k.GetValidator(ctx, msg.ValidatorDstAddr)
	if !found {
		return types.ErrBadRedelegationDst(k.Codespace()).Result()
	}

	if err := checkOperatorAsDelegator(k, msg.DelegatorAddr, dstValidator); err != nil {
		return err.Result()
	}

	shares, err := k.ValidateUnbondAmount(ctx, msg.DelegatorAddr, msg.ValidatorSrcAddr, msg.Amount.Amount)
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

	// publish redelegate event
	if k.PbsbServer != nil && ctx.IsDeliverTx() {
		event := types.ChainRedelegateEvent{
			RedelegateEvent: types.RedelegateEvent{
				StakeEvent: types.StakeEvent{
					IsFromTx: true,
				},
				Delegator:    msg.DelegatorAddr,
				SrcValidator: msg.ValidatorSrcAddr,
				DstValidator: msg.ValidatorDstAddr,
				Amount:       msg.Amount.Amount,
				Denom:        msg.Amount.Denom,
				TxHash:       ctx.Value(baseapp.TxHashKey).(string),
			},
			ChainId: msg.SideChainId,
		}
		k.PbsbServer.Publish(event)
	}

	return sdk.Result{Data: finishTime, Tags: tags}
}

func handleMsgSideChainUndelegate(ctx sdk.Context, msg MsgSideChainUndelegate, k keeper.Keeper) sdk.Result {
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}

	if msg.Amount.Denom != k.BondDenom(ctx) {
		return ErrBadDenom(k.Codespace()).Result()
	}

	shares, err := k.ValidateUnbondAmount(ctx, msg.DelegatorAddr, msg.ValidatorAddr, msg.Amount.Amount)
	if err != nil {
		return err.Result()
	}

	var (
		ubd    types.UnbondingDelegation
		events sdk.Events
	)

	ubd, err = k.BeginUnbonding(ctx, msg.DelegatorAddr, msg.ValidatorAddr, shares, true)
	if err != nil {
		return err.Result()
	}

	finishTime := types.MsgCdc.MustMarshalBinaryLengthPrefixed(ubd.MinTime)

	tags := sdk.NewTags(
		tags.Delegator, []byte(msg.DelegatorAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorAddr.String()),
		tags.EndTime, finishTime,
	)

	// publish undelegate event
	if k.PbsbServer != nil && ctx.IsDeliverTx() {
		txHash, isFromTx := ctx.Value(baseapp.TxHashKey).(string)
		event := types.ChainUndelegateEvent{
			UndelegateEvent: types.UndelegateEvent{
				StakeEvent: types.StakeEvent{
					IsFromTx: isFromTx,
				},
				Delegator: msg.DelegatorAddr,
				Validator: msg.ValidatorAddr,
				Amount:    msg.Amount.Amount,
				Denom:     msg.Amount.Denom,
				TxHash:    txHash,
			},
			ChainId: msg.SideChainId,
		}
		k.PbsbServer.Publish(event)
	}

	return sdk.Result{Data: finishTime, Tags: tags, Events: events}
}

func handleMsgSideChainStakeMigration(ctx sdk.Context, msg MsgSideChainStakeMigration, k keeper.Keeper) sdk.Result {
	if scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, k.DestChainName); err != nil {
		return ErrInvalidSideChainId(k.Codespace()).Result()
	} else {
		ctx = scCtx
	}

	denom := k.BondDenom(ctx)
	if msg.Amount.Denom != denom {
		return ErrBadDenom(k.Codespace()).Result()
	}

	shares, sdkErr := k.ValidateUnbondAmount(ctx, msg.RefundAddr, msg.ValidatorSrcAddr, msg.Amount.Amount)
	if sdkErr != nil {
		return sdkErr.Result()
	}

	// unbond immediately
	ubd, events, sdkErr := k.UnboundDelegation(ctx, msg.RefundAddr, msg.ValidatorSrcAddr, shares)
	if sdkErr != nil {
		return sdkErr.Result()
	}

	// send coins to pegAccount
	relayFee := sdk.NewCoin(denom, types.StakeMigrationRelayFee)
	transferAmt := sdk.Coins{ubd.Balance}.Plus(sdk.Coins{relayFee})
	_, sdkErr = k.BankKeeper.SendCoins(ctx, msg.RefundAddr, sdk.PegAccount, transferAmt)
	if sdkErr != nil {
		return sdkErr.Result()
	}

	// send cross-chain package
	bscAmount := bsc.ConvertBCAmountToBSCAmount(ubd.Balance.Amount)
	stakeMigrationSynPackage := types.StakeMigrationSynPackage{
		OperatorAddress:  msg.ValidatorDstAddr,
		DelegatorAddress: msg.DelegatorAddr,
		RefundAddress:    msg.RefundAddr,
		Amount:           bscAmount,
	}

	encodedPackage, err := rlp.EncodeToBytes(stakeMigrationSynPackage)
	if err != nil {
		return sdk.ErrInternal("encode stake migration package error").Result()
	}

	bscRelayFee := bsc.ConvertBCAmountToBSCAmount(relayFee.Amount)
	sendSeq, sdkErr := k.IbcKeeper.CreateRawIBCPackageByIdWithFee(ctx.DepriveSideChainKeyPrefix(), k.DestChainId, types.StakeMigrationChannelID, sdk.SynCrossChainPackageType,
		encodedPackage, *bscRelayFee)
	if sdkErr != nil {
		return sdkErr.Result()
	}

	if k.PbsbServer != nil && ctx.IsDeliverTx() {
		uEvent := types.ChainUndelegateEvent{
			UndelegateEvent: types.UndelegateEvent{
				StakeEvent: types.StakeEvent{
					IsFromTx: true,
				},
				Delegator: msg.RefundAddr,
				Validator: msg.ValidatorSrcAddr,
				Amount:    msg.Amount.Amount,
				Denom:     msg.Amount.Denom,
				TxHash:    ctx.Value(baseapp.TxHashKey).(string),
			},
			ChainId: k.DestChainName,
		}
		k.PbsbServer.Publish(uEvent)

		completedUBDEvent := types.CompletedUBDEvent{
			CompUBDs: []types.UnbondingDelegation{ubd},
			ChainId:  k.DestChainName,
		}
		k.PbsbServer.Publish(completedUBDEvent)

		ctEvent := pubsub.CrossTransferEvent{
			ChainId:    k.DestChainName,
			RelayerFee: types.StakeMigrationRelayFee,
			Type:       types.TransferOutType,
			From:       msg.RefundAddr.String(),
			Denom:      denom,
			To:         []pubsub.CrossReceiver{{sdk.PegAccount.String(), ubd.Balance.Amount}},
		}
		k.PbsbServer.Publish(ctEvent)
	}

	finishTime := types.MsgCdc.MustMarshalBinaryLengthPrefixed(ubd.MinTime)
	txTags := sdk.NewTags(
		tags.Delegator, []byte(msg.RefundAddr.String()),
		tags.SrcValidator, []byte(msg.ValidatorSrcAddr.String()),
		tags.EndTime, finishTime,
	)

	for _, coin := range transferAmt {
		if coin.Amount > 0 {
			txTags = append(txTags, sdk.GetPegInTag(coin.Denom, coin.Amount))
		}
	}
	txTags = append(txTags, sdk.MakeTag(types.TagStakeMigrationSendSequence, []byte(strconv.FormatUint(sendSeq, 10))))

	return sdk.Result{
		Tags:   txTags,
		Events: events,
	}
}

// we allow the self-delegator delegating/redelegating to its validator.
// but the operator is not allowed if it is not a self-delegator
func checkOperatorAsDelegator(k Keeper, delegator sdk.AccAddress, validator Validator) sdk.Error {
	delegatorIsOperator := bytes.Equal(delegator.Bytes(), validator.OperatorAddr.Bytes())
	operatorIsSelfDelegator := validator.IsSelfDelegator(sdk.AccAddress(validator.OperatorAddr))

	if delegatorIsOperator && !operatorIsSelfDelegator {
		return ErrInvalidDelegator(k.Codespace())
	}
	return nil
}
