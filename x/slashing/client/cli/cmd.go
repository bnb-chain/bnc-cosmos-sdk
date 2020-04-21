package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

func AddCommands(root *cobra.Command, cdc *codec.Codec) {
	slashingCmd := &cobra.Command{
		Use:   "slashing",
		Short: "slashing validators",
	}

	slashingCmd.AddCommand(
		client.PostCommands(
			GetCmdBscSubmitEvidence(cdc),
			GetCmdSideChainUnjail(cdc),
		)...)

	slashingCmd.AddCommand(
		client.GetCommands(
			GetCmdQuerySideChainSigningInfo("slashing", "stake", cdc),
		)...)

	root.AddCommand(slashingCmd)
}
