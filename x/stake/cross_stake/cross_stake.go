package cross_stake

import (
	"fmt"
	"math/big"
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

func (app *CrossStakeApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, relayerFee int64) sdk.ExecuteResult {
	crossStakeSyncPackage, sdkErr := DeserializeCrossStakeSynPackage(payload)
	if sdkErr != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake sync claim error", "err", sdkErr.Error(), "claim", string(payload))
		panic("unmarshal cross stake claim error")
	}

	var result sdk.ExecuteResult
	var err error
	switch crossStakeSyncPackage.PackageType {
	case CrossStakeTypeDelegate:
		result, err = app.handleDelegate(ctx, payload, relayerFee)
	case CrossStakeTypeUndelegate:
		result, err = app.handleUndelegate(ctx, payload, relayerFee)
	case CrossStakeTypeClaimReward:
		result, err = app.handleClaimReward(ctx, payload, relayerFee)
	case CrossStakeTypeClaimUndelegated:
		result, err = app.handleClaimUndelegated(ctx, payload, relayerFee)
	case CrossStakeTypeReinvest:
		result, err = app.handleReinvest(ctx, payload, relayerFee)
	case CrossStakeTypeRedelegate:
		result, err = app.handleRedelegate(ctx, payload, relayerFee)
	}
	if err != nil {
		panic(err)
	}

	return result
}

func (app *CrossStakeApp) handleDelegate(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType CrossStakePackageType
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

	stakeAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegator")
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
	_, sdkErr := app.stakeKeeper.BankKeeper.SendCoins(ctx, stakeAddr, sdk.PegAccount, transferAmount)
	if sdkErr != nil {
		app.stakeKeeper.Logger(ctx).Error("send coins error", "err", sdkErr.Error())
		return sdk.ExecuteResult{}, sdkErr
	}

	_, err = app.stakeKeeper.Delegate(ctx, stakeAddr, delegation, validator, true)
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
				Delegator: stakeAddr,
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
		PackageType CrossStakePackageType
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

	stakeAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegator")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, stakeAddr, pack.Validator, pack.Amount.Int64())
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

	_, err = app.stakeKeeper.BeginUnbonding(ctx, stakeAddr, pack.Validator, shares)
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
				Delegator: stakeAddr,
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

func (app *CrossStakeApp) handleClaimReward(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType   CrossStakePackageType
		DelAddr       types.SmartChainAddress
		Receiver      types.SmartChainAddress
		BSCRelayerFee *big.Int
	}
	type AckPackage struct {
		PackageType CrossStakePackageType
		DelAddr     types.SmartChainAddress
		Receiver    types.SmartChainAddress
		Amount      *big.Int
		EventCode   uint8
		RewardAddr  types.SmartChainAddress
		Err         string
	}

	var pack SynPackage
	err := rlp.DecodeBytes(payload, &pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake claim reward sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	symbol := app.stakeKeeper.BondDenom(ctx)
	stakeAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegator")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}
	rewardAddr, err := types.GetStakeCAoB(stakeAddr.Bytes(), "Reward")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	balance := app.stakeKeeper.BankKeeper.GetCoins(ctx, rewardAddr).AmountOf(symbol)
	if balance == 0 {
		sdkErr := types.ErrNoBalance("no delegation or no pending reward")
		ackPackage := &AckPackage{
			PackageType: pack.PackageType,
			DelAddr:     pack.DelAddr,
			Receiver:    pack.Receiver,
			EventCode:   uint8(0),
			Err:         sdkErr.Error(),
		}
		copy(ackPackage.RewardAddr[:], rewardAddr)
		ackBytes, err := rlp.EncodeToBytes(ackPackage)
		if err != nil {
			return sdk.ExecuteResult{}, err
		}
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: ackBytes,
		}, nil
	}

	coin := sdk.NewCoin(app.stakeKeeper.BondDenom(ctx), balance)
	bscRelayerFee := sdk.NewCoin(app.stakeKeeper.BondDenom(ctx), pack.BSCRelayerFee.Int64())
	transferAmount := sdk.Coins{coin, bscRelayerFee}

	_, sdkErr := app.stakeKeeper.BankKeeper.SendCoins(ctx, rewardAddr, sdk.PegAccount, transferAmount)
	if sdkErr != nil {
		app.stakeKeeper.Logger(ctx).Error("send coins error", "err", sdkErr.Error())
		return sdk.ExecuteResult{}, sdkErr
	}

	if ctx.IsDeliverTx() {
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, rewardAddr})
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.DelAddr.String(), pack.Receiver.String(), "", "",
			CrossStakeClaimRewardType, relayerFee, bscRelayerFee.Amount)
	}

	bscTransferAmount := bsc.ConvertBCAmountToBSCAmount(balance)
	ackPackage := &AckPackage{
		Receiver:  pack.Receiver,
		Amount:    bscTransferAmount,
		EventCode: uint8(1),
	}
	ackBytes, err := rlp.EncodeToBytes(ackPackage)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	tags := sdk.NewTags(
		fmt.Sprintf(types.TagCrossChainStakeClaimReward, pack.DelAddr.String(), pack.Receiver.String()), []byte(strconv.FormatInt(balance, 10)),
		types.TagRelayerFee, []byte(strconv.FormatInt(pack.BSCRelayerFee.Int64(), 10)),
	)
	return sdk.ExecuteResult{
		Payload: ackBytes,
		Tags:    tags,
	}, nil
}

func (app *CrossStakeApp) handleClaimUndelegated(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType   CrossStakePackageType
		DelAddr       types.SmartChainAddress
		Receiver      types.SmartChainAddress
		BSCRelayerFee *big.Int
	}
	type AckPackage struct {
		PackageType CrossStakePackageType
		DelAddr     types.SmartChainAddress
		Receiver    types.SmartChainAddress
		Amount      *big.Int
		EventCode   uint8
		StakeAddr   types.SmartChainAddress
		Err         string
	}

	var pack SynPackage
	err := rlp.DecodeBytes(payload, &pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake claim undelegated sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	symbol := app.stakeKeeper.BondDenom(ctx)
	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegator")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	balance := app.stakeKeeper.BankKeeper.GetCoins(ctx, delAddr).AmountOf(symbol)
	if balance == 0 {
		sdkErr := types.ErrNoBalance("no delegation or no pending undelegated")
		ackPackage := &AckPackage{
			PackageType: pack.PackageType,
			DelAddr:     pack.DelAddr,
			Receiver:    pack.Receiver,
			EventCode:   uint8(0),
			Err:         sdkErr.Error(),
		}
		copy(ackPackage.StakeAddr[:], delAddr)
		ackBytes, err := rlp.EncodeToBytes(ackPackage)
		if err != nil {
			return sdk.ExecuteResult{}, err
		}
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: ackBytes,
		}, nil
	}

	coin := sdk.NewCoin(app.stakeKeeper.BondDenom(ctx), balance)
	bscRelayerFee := sdk.NewCoin(app.stakeKeeper.BondDenom(ctx), pack.BSCRelayerFee.Int64())
	transferAmount := sdk.Coins{coin, bscRelayerFee}

	_, sdkErr := app.stakeKeeper.BankKeeper.SendCoins(ctx, delAddr, sdk.PegAccount, transferAmount)
	if sdkErr != nil {
		app.stakeKeeper.Logger(ctx).Error("send coins error", "err", sdkErr.Error())
		return sdk.ExecuteResult{}, sdkErr
	}

	if ctx.IsDeliverTx() {
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, delAddr})
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.DelAddr.String(), pack.Receiver.String(), "", "",
			CrossStakeClaimUndelegatedType, relayerFee, bscRelayerFee.Amount)
	}

	bscTransferAmount := bsc.ConvertBCAmountToBSCAmount(balance)
	ackPackage := &AckPackage{
		Receiver:  pack.Receiver,
		Amount:    bscTransferAmount,
		EventCode: uint8(1),
	}
	ackBytes, err := rlp.EncodeToBytes(ackPackage)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	tags := sdk.NewTags(
		fmt.Sprintf(types.TagCrossChainStakeClaimUnstake, pack.DelAddr.String(), pack.Receiver.String()), []byte(strconv.FormatInt(balance, 10)),
		types.TagRelayerFee, []byte(strconv.FormatInt(pack.BSCRelayerFee.Int64(), 10)),
	)
	return sdk.ExecuteResult{
		Payload: ackBytes,
		Tags:    tags,
	}, nil
}

func (app *CrossStakeApp) handleReinvest(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType CrossStakePackageType
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
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake reinvest sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	sideChainId := app.stakeKeeper.ScKeeper.BscSideChainId(ctx)
	if scCtx, err := app.stakeKeeper.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return sdk.ExecuteResult{}, err
	} else {
		ctx = scCtx
	}

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegator")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}
	rewardAddr, err := types.GetStakeCAoB(delAddr.Bytes(), "Reward")
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

	var amount int64
	balance := app.stakeKeeper.BankKeeper.GetCoins(ctx, rewardAddr).AmountOf(app.stakeKeeper.BondDenom(ctx))
	if pack.Amount.Int64() == 0 {
		amount = balance
	} else if pack.Amount.Int64() <= balance {
		amount = balance
	} else {
		sdkErr := types.ErrNotEnoughBalance("reinvest amount is greater than pending reward")
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

	minDelegationChange := app.stakeKeeper.MinDelegationChange(ctx)
	if amount < minDelegationChange {
		sdkErr := types.ErrBadDelegationAmount(types.DefaultCodespace, fmt.Sprintf("reinvest must not be less than %d", minDelegationChange))
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

	delegation := sdk.NewCoin(app.stakeKeeper.BondDenom(ctx), amount)
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
			CrossStakeReinvestType, relayerFee, 0)
	}

	return sdk.ExecuteResult{}, nil
}

func (app *CrossStakeApp) handleRedelegate(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType CrossStakePackageType
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

	delAddr, err := types.GetStakeCAoB(pack.DelAddr[:], "Delegator")
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
