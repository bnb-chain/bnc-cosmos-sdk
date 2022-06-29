package cross_stake

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/cosmos/cosmos-sdk/baseapp"
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
		PackageType CrossStakeSynPackage
		delAddr     types.SmartChainAddress
		validator   sdk.ValAddress
		amount      *big.Int
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

	stakeAddr, err := types.GetStakeCAoB(pack.delAddr[:], "stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	validator, found := app.stakeKeeper.GetValidator(ctx, pack.validator)
	if !found {
		return sdk.ExecuteResult{
			Err:     types.ErrNoValidatorFound(types.DefaultCodespace),
			Payload: payload,
		}, nil
	}
	if validator.Jailed {
		return sdk.ExecuteResult{
			Err:     types.ErrValidatorJailed(types.DefaultCodespace),
			Payload: payload,
		}, nil
	}

	delegation := sdk.NewCoin(app.stakeKeeper.BondDenom(ctx), pack.amount.Int64())
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
				Validator: pack.validator,
				Amount:    pack.amount.Int64(),
				Denom:     app.stakeKeeper.BondDenom(ctx),
				TxHash:    ctx.Value(baseapp.TxHashKey).(string),
			},
			SideChainId: sideChainId,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.delAddr.String(), "", "", pack.validator.String(),
			CrossStakeDelegateType, relayerFee, 0)
	}

	return sdk.ExecuteResult{}, nil
}

func (app *CrossStakeApp) handleUndelegate(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType CrossStakeSynPackage
		delAddr     types.SmartChainAddress
		validator   sdk.ValAddress
		amount      *big.Int
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

	stakeAddr, err := types.GetStakeCAoB(pack.delAddr[:], "stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, stakeAddr, pack.validator, pack.amount.Int64())
	if shares.IsZero() && sdkErr != nil {
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: payload,
		}, nil
	}

	_, err = app.stakeKeeper.BeginUnbonding(ctx, stakeAddr, pack.validator, shares)
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
				Validator: pack.validator,
				Amount:    shares.RawInt(),
				Denom:     app.stakeKeeper.BondDenom(ctx),
				TxHash:    ctx.Value(baseapp.TxHashKey).(string),
			},
			SideChainId: sideChainId,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.delAddr.String(), "", pack.validator.String(), "",
			CrossStakeUndelegateType, relayerFee, 0)
	}
	return sdk.ExecuteResult{}, nil
}

func (app *CrossStakeApp) handleClaimReward(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType   CrossStakeSynPackage
		delAddr       types.SmartChainAddress
		receiver      types.SmartChainAddress
		BSCRelayerFee *big.Int
	}
	//type AckPackage struct {
	//	receiver  types.SmartChainAddress
	//	amount    *big.Int
	//	SyncFee   *big.Int
	//	ErrorCode uint8
	//}

	var pack SynPackage
	err := rlp.DecodeBytes(payload, &pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake claim reward sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	symbol := app.stakeKeeper.BondDenom(ctx)
	stakeAddr, err := types.GetStakeCAoB(pack.delAddr[:], "Stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}
	rewardAddr, err := types.GetStakeCAoB(stakeAddr.Bytes(), "Reward")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	balance := app.stakeKeeper.BankKeeper.GetCoins(ctx, rewardAddr).AmountOf(symbol)
	if balance == 0 {
		return sdk.ExecuteResult{
			Err:     types.ErrNoBalance("no delegation or no pending reward"),
			Payload: payload,
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
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.delAddr.String(), pack.receiver.String(), "", "",
			CrossStakeClaimRewardType, relayerFee, bscRelayerFee.Amount)
	}

	//bscTransferAmount := bsc.ConvertBCAmountToBSCAmount(balance)
	//bscRelayerFeeBSCAmount := bsc.ConvertBCAmountToBSCAmount(pack.BSCRelayerFee.Int64())
	//
	//ackPackage := &AckPackage{
	//	receiver:  pack.receiver,
	//	amount:    bscTransferAmount,
	//	SyncFee:   bscRelayerFeeBSCAmount,
	//	ErrorCode: 0,
	//}
	//
	//encodedBytes, err := rlp.EncodeToBytes(ackPackage)
	//if err != nil {
	//	return sdk.ExecuteResult{}, err
	//}

	tags := sdk.NewTags(
		fmt.Sprintf(types.TagCrossChainStakeClaimReward, pack.delAddr.String(), pack.receiver.String()), []byte(strconv.FormatInt(balance, 10)),
		types.TagRelayerFee, []byte(strconv.FormatInt(pack.BSCRelayerFee.Int64(), 10)),
	)
	return sdk.ExecuteResult{
		Payload: payload,
		Tags:    tags,
	}, nil
}

func (app *CrossStakeApp) handleClaimUndelegated(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType   CrossStakeSynPackage
		delAddr       types.SmartChainAddress
		receiver      types.SmartChainAddress
		BSCRelayerFee *big.Int
	}
	//type AckPackage struct {
	//	receiver  types.SmartChainAddress
	//	amount    *big.Int
	//	SyncFee   *big.Int
	//	ErrorCode uint8
	//}

	var pack SynPackage
	err := rlp.DecodeBytes(payload, &pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal cross stake claim undelegated sync claim error", "err", err.Error(), "claim", string(payload))
		return sdk.ExecuteResult{}, err
	}

	symbol := app.stakeKeeper.BondDenom(ctx)
	delAddr, err := types.GetStakeCAoB(pack.delAddr[:], "Stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	balance := app.stakeKeeper.BankKeeper.GetCoins(ctx, delAddr).AmountOf(symbol)
	if balance == 0 {
		return sdk.ExecuteResult{
			Err:     types.ErrNoBalance("no delegation or no pending undelegated"),
			Payload: payload,
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
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.delAddr.String(), pack.receiver.String(), "", "",
			CrossStakeClaimUndelegatedType, relayerFee, bscRelayerFee.Amount)
	}

	//bscTransferAmount := bsc.ConvertBCAmountToBSCAmount(balance)
	//bscRelayerFeeBSCAmount := bsc.ConvertBCAmountToBSCAmount(pack.BSCRelayerFee.Int64())
	//
	//ackPackage := &AckPackage{
	//	receiver:  pack.receiver,
	//	amount:    bscTransferAmount,
	//	SyncFee:   bscRelayerFeeBSCAmount,
	//	ErrorCode: 0,
	//}
	//
	//encodedBytes, err := rlp.EncodeToBytes(ackPackage)
	//if err != nil {
	//	return sdk.ExecuteResult{}, err
	//}

	tags := sdk.NewTags(
		fmt.Sprintf(types.TagCrossChainStakeClaimUnstake, pack.delAddr.String(), pack.receiver.String()), []byte(strconv.FormatInt(balance, 10)),
		types.TagRelayerFee, []byte(strconv.FormatInt(pack.BSCRelayerFee.Int64(), 10)),
	)
	return sdk.ExecuteResult{
		Payload: payload,
		Tags:    tags,
	}, nil
}

func (app *CrossStakeApp) handleReinvest(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType CrossStakeSynPackage
		delAddr     types.SmartChainAddress
		validator   sdk.ValAddress
		amount      *big.Int
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

	delAddr, err := types.GetStakeCAoB(pack.delAddr[:], "stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}
	rewardAddr, err := types.GetStakeCAoB(delAddr.Bytes(), "Reward")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	validator, found := app.stakeKeeper.GetValidator(ctx, pack.validator)
	if !found {
		return sdk.ExecuteResult{
			Err:     types.ErrNoValidatorFound(types.DefaultCodespace),
			Payload: payload,
		}, nil
	}
	if validator.Jailed {
		return sdk.ExecuteResult{
			Err:     types.ErrValidatorJailed(types.DefaultCodespace),
			Payload: payload,
		}, nil
	}

	var amount int64
	balance := app.stakeKeeper.BankKeeper.GetCoins(ctx, rewardAddr).AmountOf(app.stakeKeeper.BondDenom(ctx))
	if pack.amount.Int64() == 0 {
		amount = balance
	} else if pack.amount.Int64() <= balance {
		amount = balance
	} else {
		return sdk.ExecuteResult{
			Err:     types.ErrNotEnoughBalance("reinvest amount is greater than pending reward"),
			Payload: payload,
		}, nil
	}

	minDelegationChange := app.stakeKeeper.MinDelegationChange(ctx)
	if amount < minDelegationChange {
		return sdk.ExecuteResult{
			Err: types.ErrBadDelegationAmount(types.DefaultCodespace,
				fmt.Sprintf("reinvest must not be less than %d", minDelegationChange)),
			Payload: payload,
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
				Validator: pack.validator,
				Amount:    pack.amount.Int64(),
				Denom:     app.stakeKeeper.BondDenom(ctx),
				TxHash:    ctx.Value(baseapp.TxHashKey).(string),
			},
			SideChainId: sideChainId,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.delAddr.String(), "", "", pack.validator.String(),
			CrossStakeReinvestType, relayerFee, 0)
	}

	return sdk.ExecuteResult{}, nil
}

func (app *CrossStakeApp) handleRedelegate(ctx sdk.Context, payload []byte, relayerFee int64) (sdk.ExecuteResult, error) {
	type SynPackage struct {
		PackageType CrossStakeSynPackage
		delAddr     types.SmartChainAddress
		valSrc      sdk.ValAddress
		valDst      sdk.ValAddress
		amount      *big.Int
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

	valDst, found := app.stakeKeeper.GetValidator(ctx, pack.valDst)
	if !found {
		return sdk.ExecuteResult{
			Err:     types.ErrNoValidatorFound(types.DefaultCodespace),
			Payload: payload,
		}, nil
	}
	if valDst.Jailed {
		return sdk.ExecuteResult{
			Err:     types.ErrValidatorJailed(types.DefaultCodespace),
			Payload: payload,
		}, nil
	}

	delAddr, err := types.GetStakeCAoB(pack.delAddr[:], "stake")
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	shares, sdkErr := app.stakeKeeper.ValidateUnbondAmount(ctx, delAddr, pack.valSrc, pack.amount.Int64())
	if shares.IsZero() && sdkErr != nil {
		return sdk.ExecuteResult{
			Err:     sdkErr,
			Payload: payload,
		}, nil
	}

	_, err = app.stakeKeeper.BeginRedelegation(ctx, delAddr, pack.valSrc, pack.valDst, shares)
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
				SrcValidator: pack.valSrc,
				DstValidator: pack.valDst,
				Amount:       pack.amount.Int64(),
				Denom:        app.stakeKeeper.BondDenom(ctx),
				TxHash:       ctx.Value(baseapp.TxHashKey).(string),
			},
			SideChainId: sideChainId,
		}
		app.stakeKeeper.PbsbServer.Publish(event)
		publishCrossChainEvent(ctx, app.stakeKeeper, pack.delAddr.String(), "", pack.valSrc.String(), pack.valDst.String(),
			CrossStakeRedelegateType, relayerFee, 0)
	}

	return sdk.ExecuteResult{}, nil
}
