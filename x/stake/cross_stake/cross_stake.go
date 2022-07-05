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

func (app *CrossStakeApp) ExecuteAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	app.stakeKeeper.Logger(ctx).Error("received cross stake ack package ")
	return sdk.ExecuteResult{}
}

func (app *CrossStakeApp) ExecuteFailAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	app.stakeKeeper.Logger(ctx).Error("received cross stake fail ack package ")
	return sdk.ExecuteResult{}
}

func (app *CrossStakeApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, relayerFee int64) sdk.ExecuteResult {
	crossStakeSyncPackage, sdkErr := DeserializeCrossStakeSynPackage(payload)
	if sdkErr != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake sync claim error", "err", sdkErr.Error(), "claim", string(payload))
		panic("unmarshal cross stake claim error")
	}

	var result sdk.ExecuteResult
	var err error
	switch crossStakeSyncPackage.PackageType {
	case types.CrossStakeTypeDelegate:
		result, err = app.handleDelegate(ctx, payload, relayerFee)
	case types.CrossStakeTypeUndelegate:
		result, err = app.handleUndelegate(ctx, payload, relayerFee)
	case types.CrossStakeTypeRedelegate:
		result, err = app.handleRedelegate(ctx, payload, relayerFee)
	}
	if err != nil {
		panic(err)
	}

	return result
}

func (app *CrossStakeApp) handleDelegate(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType types.CrossStakePackageType
		DelAddr     types.SmartChainAddress
		Validator   sdk.ValAddress
		Amount      *big.Int
	}

	type AckPackage struct {
		SynPackage
		Err string
	}

	var pack SynPackage
	err := rlp.DecodeBytes(payload, &pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake delegate sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	sideChainId := app.stakeKeeper.ScKeeper.BscSideChainId(ctx)
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	validator, found := app.stakeKeeper.GetValidator(ctx, pack.Validator)
	if !found || validator.Jailed {
		var sdkErr sdk.Error
		if !found {
			sdkErr = types.ErrNoValidatorFound(types.DefaultCodespace)
		} else {
			sdkErr = types.ErrValidatorJailed(types.DefaultCodespace)
		}
		ackPack := &AckPackage{pack, err.Error()}
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
	_, sdkErr := app.stakeKeeper.BankKeeper.SendCoins(ctx, delAddr, sdk.PegAccount, transferAmount)
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
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.DelAddr.String(), "", "", pack.Validator.String(),
			CrossStakeDelegateType, relayerFee, 0)
	}

	return sdk.ExecuteResult{}, nil
}

func (app *CrossStakeApp) handleUndelegate(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType types.CrossStakePackageType
		DelAddr     types.SmartChainAddress
		Validator   sdk.ValAddress
		Amount      *big.Int
	}
	type AckPackage struct {
		SynPackage
		Err string
	}

	var pack SynPackage
	err := rlp.DecodeBytes(payload, &pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake undelegate sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	sideChainId := app.stakeKeeper.ScKeeper.BscSideChainId(ctx)
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.Validator, pack.Amount.Int64())
	if shares.IsZero() && sdkErr != nil {
		ackPack := &AckPackage{pack, err.Error()}
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
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.DelAddr.String(), "", pack.Validator.String(), "",
			CrossStakeUndelegateType, relayerFee, 0)
	}
	return sdk.ExecuteResult{}, nil
}

func (app *CrossStakeApp) handleRedelegate(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType types.CrossStakePackageType
		DelAddr     types.SmartChainAddress
		ValSrc      sdk.ValAddress
		ValDst      sdk.ValAddress
		Amount      *big.Int
	}
	type AckPackage struct {
		SynPackage
		Err string
	}

	var pack SynPackage
	err := rlp.DecodeBytes(payload, &pack)
	if err != nil {
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
		if !found {
			sdkErr = types.ErrNoValidatorFound(types.DefaultCodespace)
		} else {
			sdkErr = types.ErrValidatorJailed(types.DefaultCodespace)
		}
		ackPack := &AckPackage{pack, err.Error()}
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

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.ValSrc, pack.Amount.Int64())
	if shares.IsZero() && sdkErr != nil {
		ackPack := &AckPackage{pack, err.Error()}
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
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.DelAddr.String(), "", pack.ValSrc.String(), pack.ValDst.String(),
			CrossStakeRedelegateType, relayerFee, 0)
	}

	return sdk.ExecuteResult{}, nil
}
