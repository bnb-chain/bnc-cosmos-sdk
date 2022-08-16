package cross_stake

import (
	"math/big"

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

func (app *CrossStakeApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, relayFee int64) sdk.ExecuteResult {
	app.stakeKeeper.Logger(ctx).Info("receive cross stake syn package")
	pack, err := DeserializeCrossStakeSynPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake sync claim error", "err", err.Error(), "claim", string(payload))
		panic("unmarshal cross stake claim error")
	}

	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.CrossStakeDelegateSynPackage:
		result, err = app.handleDelegate(ctx, p, relayFee)
	case *types.CrossStakeUndelegateSynPackage:
		result, err = app.handleUndelegate(ctx, p, relayFee)
	case *types.CrossStakeRedelegateSynPackage:
		result, err = app.handleRedelegate(ctx, p, relayFee)
	default:
		panic("Unknown cross stake syn package type")
	}
	if err != nil {
		panic(err)
	}

	return result
}

func (app *CrossStakeApp) ExecuteAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	if len(payload) == 0 {
		app.stakeKeeper.Logger(ctx).Info("receive cross stake ack package")
		return sdk.ExecuteResult{}
	}

	pack, err := DeserializeCrossStakeRefundPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake refund package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	var result sdk.ExecuteResult
	switch pack.EventType {
	case types.CrossStakeTypeDistributeReward:
		result, err = app.handleDistributeRewardRefund(ctx, pack)
	case types.CrossStakeTypeDistributeUndelegated:
		result, err = app.handleDistributeUndelegatedRefund(ctx, pack)
	default:
		app.stakeKeeper.Logger(ctx).Error("unknown cross stake refund event type", "package", string(payload))
		return sdk.ExecuteResult{}
	}
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("handle cross stake refund package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	return result
}

func (app *CrossStakeApp) ExecuteFailAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	if len(payload) == 0 {
		app.stakeKeeper.Logger(ctx).Info("receive cross stake fail ack package")
		return sdk.ExecuteResult{}
	}

	pack, err := DeserializeCrossStakeFailAckPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake fail ack package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.CrossStakeDistributeRewardSynPackage:
		bcAmount := bsc.ConvertBSCAmountToBCAmount(p.Amount)
		refundPackage := &types.CrossStakeRefundPackage{
			EventType: types.CrossStakeTypeDistributeReward,
			Amount:    big.NewInt(bcAmount),
			Recipient: p.Recipient,
		}
		result, err = app.handleDistributeRewardRefund(ctx, refundPackage)
	case *types.CrossStakeDistributeUndelegatedSynPackage:
		bcAmount := bsc.ConvertBSCAmountToBCAmount(p.Amount)
		refundPackage := &types.CrossStakeRefundPackage{
			EventType: types.CrossStakeTypeDistributeUndelegated,
			Amount:    big.NewInt(bcAmount),
			Recipient: p.Recipient,
		}
		result, err = app.handleDistributeUndelegatedRefund(ctx, refundPackage)
	default:
		app.stakeKeeper.Logger(ctx).Error("unknown cross stake fail ack event type", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("handle cross stake fail ack package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	return result
}

func (app *CrossStakeApp) handleDelegate(ctx sdk.Context, pack *types.CrossStakeDelegateSynPackage, relayFee int64) (sdk.ExecuteResult, error) {
	sideChainId := app.stakeKeeper.DestChainName
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr := types.GetStakeCAoB(pack.DelAddr[:], types.DelegateCAoBSalt)
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
		ackPack := types.NewCrossStakeDelegationAckPackage(pack, types.CrossStakeTypeDelegate, errCode)
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
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, delAddr})
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
		PublishCrossStakeEvent(ctx, app.stakeKeeper, sdk.PegAccount.String(), []types.CrossReceiver{{delAddr.String(), pack.Amount.Int64()}},
			app.stakeKeeper.BondDenom(ctx), types.CrossStakeDelegateType, relayFee)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte{uint8(types.CrossStakeTypeDelegate)},
	)
	resultTags = append(resultTags, sdk.GetPegOutTag(delegation.Denom, delegation.Amount))

	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleUndelegate(ctx sdk.Context, pack *types.CrossStakeUndelegateSynPackage, relayFee int64) (sdk.ExecuteResult, error) {
	sideChainId := app.stakeKeeper.DestChainName
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr := types.GetStakeCAoB(pack.DelAddr[:], types.DelegateCAoBSalt)
	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.Validator, pack.Amount.Int64())
	if sdkErr != nil {
		ackPack := types.NewCrossStakeUndelegateAckPackage(pack, types.CrossStakeTypeUndelegate, CrossStakeErrBadDelegation)
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
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte{uint8(types.CrossStakeTypeUndelegate)},
	)
	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleRedelegate(ctx sdk.Context, pack *types.CrossStakeRedelegateSynPackage, relayFee int64) (sdk.ExecuteResult, error) {
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
		ackPack := types.NewCrossStakeRedelegationAckPackage(pack, types.CrossStakeTypeRedelegate, errCode)
		ackBytes, err := rlp.EncodeToBytes(ackPack)
		if err != nil {
			return sdk.ExecuteResult{}, err
		}
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: ackBytes,
		}, nil
	}

	delAddr := types.GetStakeCAoB(pack.DelAddr[:], types.DelegateCAoBSalt)
	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.ValSrc, pack.Amount.Int64())
	if sdkErr != nil {
		ackPack := types.NewCrossStakeRedelegationAckPackage(pack, types.CrossStakeTypeRedelegate, CrossStakeErrBadDelegation)
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
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte{uint8(types.CrossStakeTypeRedelegate)},
	)
	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleDistributeRewardRefund(ctx sdk.Context, pack *types.CrossStakeRefundPackage) (sdk.ExecuteResult, error) {
	symbol := app.stakeKeeper.BondDenom(ctx)
	coins := sdk.Coins{sdk.NewCoin(symbol, pack.Amount.Int64())}
	delAddr := types.GetStakeCAoB(pack.Recipient[:], types.DelegateCAoBSalt)
	refundAddr := types.GetStakeCAoB(delAddr.Bytes(), types.RewardCAoBSalt)
	_, err := app.stakeKeeper.BankKeeper.SendCoins(ctx, sdk.PegAccount, refundAddr, coins)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish  event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, refundAddr})
		PublishCrossStakeEvent(ctx, app.stakeKeeper, sdk.PegAccount.String(), []types.CrossReceiver{{refundAddr.String(), pack.Amount.Int64()}},
			app.stakeKeeper.BondDenom(ctx), types.CrossStakeDistributeRewardFailAckRefundType, 0)
	}

	return sdk.ExecuteResult{
		Tags: sdk.Tags{sdk.GetPegOutTag(symbol, pack.Amount.Int64())},
	}, nil
}

func (app *CrossStakeApp) handleDistributeUndelegatedRefund(ctx sdk.Context, pack *types.CrossStakeRefundPackage) (sdk.ExecuteResult, error) {
	symbol := app.stakeKeeper.BondDenom(ctx)
	coins := sdk.Coins{sdk.NewCoin(symbol, pack.Amount.Int64())}
	refundAddr := types.GetStakeCAoB(pack.Recipient[:], types.DelegateCAoBSalt)
	_, err := app.stakeKeeper.BankKeeper.SendCoins(ctx, sdk.PegAccount, refundAddr, coins)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish  event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, refundAddr})
		PublishCrossStakeEvent(ctx, app.stakeKeeper, sdk.PegAccount.String(), []types.CrossReceiver{{refundAddr.String(), pack.Amount.Int64()}},
			app.stakeKeeper.BondDenom(ctx), types.CrossStakeDistributeUndelegatedFailAckRefundType, 0)
	}

	return sdk.ExecuteResult{
		Tags: sdk.Tags{sdk.GetPegOutTag(symbol, pack.Amount.Int64())},
	}, nil
}
