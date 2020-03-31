package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func AddCommands(root *cobra.Command, cdc *codec.Codec) {
	slashingsidechainCmd := &cobra.Command{
		Use:   "slashingsc",
		Short: "slashing side chain commands",
	}

	slashingsidechainCmd.AddCommand(
		client.PostCommands(
			GetCmdSubmitEvidence(cdc),
		)...)

	root.AddCommand(slashingsidechainCmd)
}