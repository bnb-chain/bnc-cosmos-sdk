package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

var storeKey = "stake"

func AddCommands(root *cobra.Command, cdc *codec.Codec) {
	stakingCmd := &cobra.Command{
		Use:   "staking",
		Short: "staking commands",
	}

	stakingCmd.AddCommand(
		client.PostCommands(
			GetCmdCreateValidator(cdc),
			GetCmdRemoveValidator(cdc),
		)...,
	)
	stakingCmd.AddCommand(client.LineBreak)

	stakingCmd.AddCommand(
		client.GetCommands(
			GetCmdQueryValidator(storeKey, cdc),
			GetCmdQueryValidators(storeKey, cdc),
			GetCmdQueryUnbondingDelegations(storeKey, cdc),
		)...,
	)
	stakingCmd.AddCommand(client.LineBreak)

	stakingCmd.AddCommand(
		client.PostCommands(
			GetCmdCreateSideChainValidator(cdc),
			GetCmdEditSideChainValidator(cdc),
			GetCmdSideChainDelegate(cdc),
			GetCmdSideChainRedelegate(cdc),
			GetCmdSideChainUnbond(cdc),
		)...,
	)
	stakingCmd.AddCommand(client.LineBreak)
	stakingCmd.AddCommand(
		client.GetCommands(
			GetCmdQuerySideValidator(storeKey, cdc),
			GetCmdQuerySideValidators(storeKey, cdc),
			GetCmdQuerySideChainDelegation(storeKey, cdc),
			GetCmdQuerySideChainDelegations(storeKey, cdc),
			GetCmdQuerySideChainRedelegation(storeKey, cdc),
			GetCmdQuerySideChainRedelegations(storeKey, cdc),
			GetCmdQuerySideChainUnbondingDelegation(storeKey, cdc),
			GetCmdQuerySideChainUnbondingDelegations(storeKey, cdc),
		)...,
	)

	root.AddCommand(stakingCmd)
}
