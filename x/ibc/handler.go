package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCPackageMsg:
			return handleIBCMsg(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized IBC Msg type: " + msg.Type()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleIBCMsg(ctx sdk.Context, keeper Keeper, msg IBCPackageMsg) sdk.Result {
	err := keeper.CreateIBCPackage(ctx, msg.DestChainID, msg.ChannelID, msg.Package)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
