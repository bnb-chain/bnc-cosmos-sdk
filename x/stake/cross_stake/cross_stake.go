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
	app.stakeKeeper.Logger(ctx).Error("received cross stake ack package ")
	return sdk.ExecuteResult{}
}

func (app *CrossStakeApp) ExecuteFailAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	app.stakeKeeper.Logger(ctx).Error("received cross stake fail ack package ")
	return sdk.ExecuteResult{}
}

func (app *CrossStakeApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, relayFee int64) sdk.ExecuteResult {
	eventType, err := DeserializeCrossStakeSynPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake sync claim error", "err", err.Error(), "claim", string(payload))
		panic("unmarshal cross stake claim error")
	}

	var result sdk.ExecuteResult
	switch eventType {
	case types.CrossStakeTypeDelegate:
		result, err = app.handleDelegate(ctx, payload, relayFee)
	case types.CrossStakeTypeUndelegate:
		result, err = app.handleUndelegate(ctx, payload, relayFee)
	case types.CrossStakeTypeRedelegate:
		result, err = app.handleRedelegate(ctx, payload, relayFee)
	}
	if err != nil {
		panic(err)
	}

	return result
}

func (app *CrossStakeApp) handleDelegate(ctx sdk.Context, payload []byte, relayFee int64) (sdk.ExecuteResult, error) {
	var pack types.CrossStakeDelegateSynPackage
	if err := rlp.DecodeBytes(payload, &pack); err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake delegate sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	sideChainId := app.stakeKeeper.ScKeeper.BscSideChainId(ctx)
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegate")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

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
		ackPack := &types.CrossStakeDelegationAckPackage{
			CrossStakeDelegateSynPackage: pack,
			ErrorCode:                    errCode,
		}
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

	_, err = app.stakeKeeper.Delegate(ctx, delAddr, delegation, validator, true)
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
		PublishCrossChainEvent(ctx, app.stakeKeeper, delAddr, sdk.ValAddress{}, pack.Validator, CrossStakeDelegateType,
			relayFee)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte(strconv.FormatInt(int64(types.CrossStakeTypeDelegate), 10)),
		sdk.GetPegOutTag(delegation.Denom, delegation.Amount),
	)

	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleUndelegate(ctx sdk.Context, payload []byte, relayFee int64) (sdk.ExecuteResult, error) {
	var pack types.CrossStakeUndelegateSynPackage
	if err := rlp.DecodeBytes(payload, &pack); err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake undelegate sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	sideChainId := app.stakeKeeper.ScKeeper.BscSideChainId(ctx)
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegate")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.Validator, pack.Amount.Int64())
	if shares.IsZero() && sdkErr != nil {
		ackPack := &types.CrossStakeUndelegateAckPackage{
			CrossStakeUndelegateSynPackage: pack,
			ErrorCode:                      CrossStakeErrBadDelegation,
		}
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

	_, err = app.stakeKeeper.BeginUnbonding(ctx, delAddr, pack.Validator, shares)
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
		PublishCrossChainEvent(ctx, app.stakeKeeper, delAddr, pack.Validator, sdk.ValAddress{}, CrossStakeUndelegateType,
			relayFee)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte(strconv.FormatInt(int64(types.CrossStakeTypeUndelegate), 10)),
	)
	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}

func (app *CrossStakeApp) handleRedelegate(ctx sdk.Context, payload []byte, relayFee int64) (sdk.ExecuteResult, error) {
	var pack types.CrossStakeRedelegateSynPackage
	if err := rlp.DecodeBytes(payload, &pack); err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake redelegate sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	sideChainId := app.stakeKeeper.ScKeeper.BscSideChainId(ctx)
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
		ackPack := &types.CrossStakeRedelegateAckPackage{
			CrossStakeRedelegateSynPackage: pack,
			ErrorCode:                      errCode,
		}
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

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegate")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.ValSrc, pack.Amount.Int64())
	if shares.IsZero() && sdkErr != nil {
		ackPack := &types.CrossStakeRedelegateAckPackage{
			CrossStakeRedelegateSynPackage: pack,
			ErrorCode:                      CrossStakeErrBadDelegation,
		}
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

	_, err = app.stakeKeeper.BeginRedelegation(ctx, delAddr, pack.ValSrc, pack.ValDst, shares)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish delegate event
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
		PublishCrossChainEvent(ctx, app.stakeKeeper, delAddr, pack.ValSrc, pack.ValDst, CrossStakeRedelegateType,
			relayFee)
	}

	resultTags := sdk.NewTags(
		types.TagCrossStakePackageType, []byte(strconv.FormatInt(int64(types.CrossStakeTypeRedelegate), 10)),
	)
	return sdk.ExecuteResult{
		Tags: resultTags,
	}, nil
}
