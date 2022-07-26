package cross_stake

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/bsc"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

type CrossStakeApp struct {
	stakeKeeper Keeper
}

func NewCrossStakeApp(stakeKeeper Keeper) *CrossStakeApp {
	return &CrossStakeApp{
		stakeKeeper: stakeKeeper,
	}
}

func (app *CrossStakeApp) ExecuteAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	if len(payload) == 0 {
		app.stakeKeeper.Logger(ctx).Info("receive cross stake ack package")
		return sdk.ExecuteResult{}
	}

	eventCode, pack, err := DeserializeCrossStakeSynPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake ack package error", "err", err.Error(), "package", string(payload))
		panic("unmarshal cross stake ack package error")
	}

	var result sdk.ExecuteResult
	switch eventCode {
	case types.CrossStakeTypeDistributeReward:
		result, err = app.handleDistributeRewardRefund(ctx, pack.(types.CrossStakeDistributeRewardSynPackage))
	case types.CrossStakeTypeDistributeUndelegated:
		result, err = app.handleDistributeUndelegatedRefund(ctx, pack.(types.CrossStakeDistributeUndelegatedSynPackage))
	}
	if err != nil {
		panic(err)
	}

	return result
}

func (app *CrossStakeApp) ExecuteFailAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	app.stakeKeeper.Logger(ctx).Error("received cross stake fail ack package ")
	return sdk.ExecuteResult{}
}

func (app *CrossStakeApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, relayFee int64) sdk.ExecuteResult {
	app.stakeKeeper.Logger(ctx).Info("receive cross stake syn package")
	eventCode, pack, err := DeserializeCrossStakeSynPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake sync claim error", "err", err.Error(), "claim", string(payload))
		panic("unmarshal cross stake claim error")
	}

	var result sdk.ExecuteResult
	switch eventCode {
	case types.CrossStakeTypeDelegate:
		result, err = app.handleDelegate(ctx, pack.(types.CrossStakeDelegateSynPackage), relayFee)
	case types.CrossStakeTypeUndelegate:
		result, err = app.handleUndelegate(ctx, pack.(types.CrossStakeUndelegateSynPackage), relayFee)
	case types.CrossStakeTypeRedelegate:
		result, err = app.handleRedelegate(ctx, pack.(types.CrossStakeRedelegateSynPackage), relayFee)
	}
	if err != nil {
		panic(err)
	}

	return result
}

func (app *CrossStakeApp) handleDelegate(ctx sdk.Context, pack types.CrossStakeDelegateSynPackage, relayFee int64) (sdk.ExecuteResult, error) {
	sideChainId := app.stakeKeeper.DestChainName
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr := types.GetStakeCAoB(pack.DelAddr[:], "Delegate")
	validator, found := app.stakeKeeper.GetValidator(ctx, pack.Validator)
	if !found || validator.Jailed {
		var sdkErr sdk.Error
		var errCode uint8
		if !found {
			sdkErr = types.ErrNoValidatorFound(types.DefaultCodespace)
			errCode = CrossStakeErrValidatorNotFound
		} else {
			sdkErr = types.ErrValidatorJailed(types.DefaultCodespace)
			errCode = CrossStakeErrValidatorJailed
		}
		ackPack := types.NewCrossStakeDelegationAckPackage(&pack, errCode)
		ackPack.Amount = bsc.ConvertBCAmountToBSCAmount(ackPack.Amount.Int64())
		ackBytes, err := rlp.EncodeToBytes(ackPack)
		if err != nil {
			return sdk.ExecuteResult{}, err
		}
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: ackBytes,
		}, nil
	}

	delegation := sdk.NewCoin(app.stakeKeeper.BondDenom(ctx), pack.Amount.Int64())
	transferAmount := sdk.Coins{delegation}
	_, sdkErr := app.stakeKeeper.BankKeeper.SendCoins(ctx, sdk.PegAccount, delAddr, transferAmount)
	if sdkErr != nil {
		app.stakeKeeper.Logger(ctx).Error("send coins error", "err", sdkErr.Error())
		return sdk.ExecuteResult{}, sdkErr
	}

	_, err := app.stakeKeeper.Delegate(ctx.WithCrossStake(true), delAddr, delegation, validator, true)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish delegate event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		event := types.SideDelegateEvent{
			DelegateEvent: types.DelegateEvent{
				StakeEvent: types.StakeEvent{
					IsFromTx: true,
				},
				Delegator: delAddr,
				Validator: pack.Validator,
				Amount:    pack.Amount.Int64(),
				Denom:     app.stakeKeeper.BondDenom(ctx),
				TxHash:    ctx.Value(baseapp.TxHashKey).(string),
			},
			SideChainId: sideChainId,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
		PublishCrossChainEvent(ctx, app.stakeKeeper, delAddr, sdk.ValAddress{}, pack.Validator, types.CrossStakeDelegateType,
			relayFee)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte(strconv.FormatInt(int64(types.CrossStakeTypeDelegate), 10)),
	)
	resultTags = append(resultTags, sdk.GetPegOutTag(delegation.Denom, delegation.Amount))

	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleUndelegate(ctx sdk.Context, pack types.CrossStakeUndelegateSynPackage, relayFee int64) (sdk.ExecuteResult, error) {
	sideChainId := app.stakeKeeper.DestChainName
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr := types.GetStakeCAoB(pack.DelAddr[:], "Delegate")
	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.Validator, pack.Amount.Int64())
	if sdkErr != nil {
		ackPack := types.NewCrossStakeUndelegateAckPackage(&pack, CrossStakeErrBadDelegation)
		ackPack.Amount = bsc.ConvertBCAmountToBSCAmount(ackPack.Amount.Int64())
		ackBytes, err := rlp.EncodeToBytes(ackPack)
		if err != nil {
			return sdk.ExecuteResult{}, err
		}
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: ackBytes,
		}, nil
	}

	_, err := app.stakeKeeper.BeginUnbonding(ctx.WithCrossStake(true), delAddr, pack.Validator, shares)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish undelegate event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		event := types.SideUnDelegateEvent{
			UndelegateEvent: types.UndelegateEvent{
				StakeEvent: types.StakeEvent{
					IsFromTx: true,
				},
				Delegator: delAddr,
				Validator: pack.Validator,
				Amount:    shares.RawInt(),
				Denom:     app.stakeKeeper.BondDenom(ctx),
				TxHash:    ctx.Value(baseapp.TxHashKey).(string),
			},
			SideChainId: sideChainId,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
		PublishCrossChainEvent(ctx, app.stakeKeeper, delAddr, pack.Validator, sdk.ValAddress{}, types.CrossStakeUndelegateType,
			relayFee)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte(strconv.FormatInt(int64(types.CrossStakeTypeUndelegate), 10)),
	)
	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleRedelegate(ctx sdk.Context, pack types.CrossStakeRedelegateSynPackage, relayFee int64) (sdk.ExecuteResult, error) {
	sideChainId := app.stakeKeeper.DestChainName
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	valDst, found := app.stakeKeeper.GetValidator(ctx, pack.ValDst)
	if !found || valDst.Jailed {
		var sdkErr sdk.Error
		var errCode uint8
		if !found {
			sdkErr = types.ErrNoValidatorFound(types.DefaultCodespace)
			errCode = CrossStakeErrValidatorNotFound
		} else {
			sdkErr = types.ErrValidatorJailed(types.DefaultCodespace)
			errCode = CrossStakeErrValidatorJailed
		}
		ackPack := types.NewCrossStakeRedelegationAckPackage(&pack, errCode)
		ackPack.Amount = bsc.ConvertBCAmountToBSCAmount(ackPack.Amount.Int64())
		ackBytes, err := rlp.EncodeToBytes(ackPack)
		if err != nil {
			return sdk.ExecuteResult{}, err
		}
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: ackBytes,
		}, nil
	}

	delAddr := types.GetStakeCAoB(pack.DelAddr[:], "Delegate")
	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.ValSrc, pack.Amount.Int64())
	if sdkErr != nil {
		ackPack := types.NewCrossStakeRedelegationAckPackage(&pack, CrossStakeErrBadDelegation)
		ackPack.Amount = bsc.ConvertBCAmountToBSCAmount(ackPack.Amount.Int64())
		ackBytes, err := rlp.EncodeToBytes(ackPack)
		if err != nil {
			return sdk.ExecuteResult{}, err
		}
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: ackBytes,
		}, nil
	}

	_, err := app.stakeKeeper.BeginRedelegation(ctx.WithCrossStake(true), delAddr, pack.ValSrc, pack.ValDst, shares)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish redelegate event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		event := types.SideRedelegateEvent{
			RedelegateEvent: types.RedelegateEvent{
				StakeEvent: types.StakeEvent{
					IsFromTx: true,
				},
				Delegator:    delAddr,
				SrcValidator: pack.ValSrc,
				DstValidator: pack.ValDst,
				Amount:       pack.Amount.Int64(),
				Denom:        app.stakeKeeper.BondDenom(ctx),
				TxHash:       ctx.Value(baseapp.TxHashKey).(string),
			},
			SideChainId: sideChainId,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
		PublishCrossChainEvent(ctx, app.stakeKeeper, delAddr, pack.ValSrc, pack.ValDst, types.CrossStakeRedelegateType,
			relayFee)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte(strconv.FormatInt(int64(types.CrossStakeTypeRedelegate), 10)),
	)
	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleDistributeRewardRefund(ctx sdk.Context, pack types.CrossStakeDistributeRewardSynPackage) (sdk.ExecuteResult, error) {
	symbol := app.stakeKeeper.BondDenom(ctx)
	refundAmount := bsc.ConvertBSCAmountToBCAmount(pack.Amount)
	coins := sdk.Coins{sdk.NewCoin(symbol, refundAmount)}
	delAddr := types.GetStakeCAoB(pack.Recipient[:], "Delegate")
	refundAddr := types.GetStakeCAoB(delAddr.Bytes(), "Reward")
	_, err := app.stakeKeeper.BankKeeper.SendCoins(ctx, sdk.PegAccount, refundAddr, coins)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish  event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, refundAddr})
		event := types.RewardRefundEvent{
			RefundAddr: refundAddr,
			Amount:     refundAmount,
			Recipient:  pack.Recipient,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
	}

	return sdk.ExecuteResult{
		Tags: sdk.Tags{sdk.GetPegOutTag(symbol, refundAmount)},
	}, nil
}

func (app *CrossStakeApp) handleDistributeUndelegatedRefund(ctx sdk.Context, pack types.CrossStakeDistributeUndelegatedSynPackage) (sdk.ExecuteResult, error) {
	symbol := app.stakeKeeper.BondDenom(ctx)
	refundAmount := bsc.ConvertBSCAmountToBCAmount(pack.Amount)
	coins := sdk.Coins{sdk.NewCoin(symbol, refundAmount)}
	refundAddr := types.GetStakeCAoB(pack.Recipient[:], "Delegate")
	_, err := app.stakeKeeper.BankKeeper.SendCoins(ctx, sdk.PegAccount, refundAddr, coins)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish  event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, refundAddr})
		event := types.UndelegatedRefundEvent{
			RefundAddr: refundAddr,
			Amount:     refundAmount,
			Recipient:  pack.Recipient,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
	}

	return sdk.ExecuteResult{
		Tags: sdk.Tags{sdk.GetPegOutTag(symbol, refundAmount)},
	}, nil
}
