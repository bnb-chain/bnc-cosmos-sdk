package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func AddCommands(cmd *cobra.Command, cdc *codec.Codec) {
	ibcCmd := &cobra.Command{
		Use:   "ibc",
		Short: "ibc commands",
	}

	ibcCmd.AddCommand(
		client.PostCommands(
			IBCBindCmd(cdc),
			IBCTimeoutCmd(cdc),
			IBCTransferCmd(cdc),
		)...,
	)

	ibcCmd.AddCommand(client.LineBreak)

	ibcCmd.AddCommand(
		client.GetCommands(
			GetIBCPackageCmd(cdc),
		)...,
	)
	cmd.AddCommand(ibcCmd)
}
