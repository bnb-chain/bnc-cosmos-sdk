package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/abci/server"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// nolint - Mostly for testing
func (app *BaseApp) Check(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(sdk.RunTxModeCheck, nil, tx)
}

// nolint - full tx execution
func (app *BaseApp) Simulate(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(sdk.RunTxModeSimulate, nil, tx)
}

// nolint
func (app *BaseApp) Deliver(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(sdk.RunTxModeDeliver, nil, tx)
}

// RunForever - BasecoinApp execution and cleanup
func RunForever(app abci.Application) {

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:26658", "socket", app)
	if err != nil {
		cmn.Exit(err.Error())
		return
	}
	err = srv.Start()
	if err != nil {
		cmn.Exit(err.Error())
		return
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		err := srv.Stop()
		if err != nil {
			cmn.Exit(err.Error())
		}
	})
}

func collectInvolvedAddresses(ctx sdk.Context, msg sdk.Msg) sdk.Context {
	return addInvolvedAddressesToCtx(ctx, msg.GetInvolvedAddresses()...)
}

func addInvolvedAddressesToCtx(ctx sdk.Context, addresses ...sdk.AccAddress) (newCtx sdk.Context) {
	var newAddress []string
	if existingAddresses, ok := ctx.Value(InvolvedAddressKey).([]string); ok {
		newAddress = existingAddresses
	} else {
		newAddress = make([]string, 0)
	}
	for _, address := range addresses {
		newAddress = append(newAddress, string(address.Bytes()))
	}
	newCtx = ctx.WithValue(InvolvedAddressKey, newAddress)
	return
}
