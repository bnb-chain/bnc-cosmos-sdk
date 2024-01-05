package stake_migration

import (
	"github.com/cosmos/cosmos-sdk/bsc"
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

type StakeMigrationApp struct {
	stakeKeeper Keeper
}

func NewStakeMigrationApp(stakeKeeper Keeper) *StakeMigrationApp {
	return &StakeMigrationApp{
		stakeKeeper: stakeKeeper,
	}
}

func (app *StakeMigrationApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, relayFee int64) sdk.ExecuteResult {
	panic("receive unexpected syn package")
}

func (app *StakeMigrationApp) ExecuteAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	if len(payload) == 0 {
		app.stakeKeeper.Logger(ctx).Error("receive empty stake migration ack package")
		return sdk.ExecuteResult{}
	}

	pack, err := DeserializeStakeMigrationRefundPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal stake migration refund package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	result, err := app.handleRefund(ctx, pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("handle stake migration refund package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	return result
}

func (app *StakeMigrationApp) ExecuteFailAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	if len(payload) == 0 {
		app.stakeKeeper.Logger(ctx).Error("receive empty stake migration fail ack package")
		return sdk.ExecuteResult{}
	}

	pack, err := DeserializeStakeMigrationRefundPackage(payload)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("unmarshal stake migration refund package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	result, err := app.handleRefund(ctx, pack)
	if err != nil {
		app.stakeKeeper.Logger(ctx).Error("handle stake migration refund package error", "err", err.Error(), "package", string(payload))
		return sdk.ExecuteResult{}
	}

	return result
}

func (app *StakeMigrationApp) handleRefund(ctx sdk.Context, pack *types.StakeMigrationSynPackage) (sdk.ExecuteResult, error) {
	symbol := app.stakeKeeper.BondDenom(ctx)
	amount := bsc.ConvertBSCAmountToBCAmount(pack.Amount)
	coins := sdk.Coins{sdk.NewCoin(symbol, amount)}
	_, err := app.stakeKeeper.BankKeeper.SendCoins(ctx, sdk.PegAccount, pack.RefundAddress, coins)
	if err != nil {
		return sdk.ExecuteResult{}, err
	}

	// publish event
	if app.stakeKeeper.PbsbServer != nil && ctx.IsDeliverTx() {
		app.stakeKeeper.AddrPool.AddAddrs([]sdk.AccAddress{sdk.PegAccount, pack.RefundAddress})
		PublishStakeMigrationEvent(ctx, app.stakeKeeper, sdk.PegAccount.String(), []pubsub.CrossReceiver{{pack.RefundAddress.String(), pack.Amount.Int64()}},
			app.stakeKeeper.BondDenom(ctx), types.TransferInType, 0)
	}

	return sdk.ExecuteResult{
		Tags: sdk.Tags{sdk.GetPegOutTag(symbol, amount)},
	}, nil
}
