package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

var storeKey = "stake"
var scStoreKey = "sc"

func AddCommands(root *cobra.Command, cdc *codec.Codec) {
	stakingCmd := &cobra.Command{
		Use:   "staking",
		Short: "staking commands",
	}

	stakingCmd.AddCommand(
		client.PostCommands(
			GetCmdCreateValidator(cdc),
			GetCmdRemoveValidator(cdc),
			GetCmdCreateValidatorOpen(cdc),
			GetCmdEditValidator(cdc),
			GetCmdDelegate(cdc),
			GetCmdRedelegate(storeKey, cdc),
			GetCmdUnbond(storeKey, cdc),
		)...,
	)
	stakingCmd.AddCommand(client.LineBreak)

	stakingCmd.AddCommand(
		client.GetCommands(
			GetCmdQueryValidator(storeKey, cdc),
			GetCmdQueryValidators(storeKey, cdc),
			GetCmdQueryParams(storeKey, cdc),
			GetCmdQueryDelegation(storeKey, cdc),
			GetCmdQueryDelegations(storeKey, cdc),
			GetCmdQueryPool(storeKey, cdc),
			GetCmdQueryRedelegation(storeKey, cdc),
			GetCmdQueryRedelegations(storeKey, cdc),
			GetCmdQueryUnbondingDelegation(storeKey, cdc),
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
			GetCmdQuerySideParams(storeKey, cdc),
			GetCmdQuerySideValidator(storeKey, cdc),
			GetCmdQuerySideChainDelegation(storeKey, cdc),
			GetCmdQuerySideChainDelegations(storeKey, cdc),
			GetCmdQuerySideChainRedelegation(storeKey, cdc),
			GetCmdQuerySideChainRedelegations(storeKey, cdc),
			GetCmdQuerySideChainUnbondingDelegation(storeKey, cdc),
			GetCmdQuerySideChainUnbondingDelegations(storeKey, cdc),
			GetCmdQuerySideChainPool(storeKey, cdc),
			GetCmdQuerySideChainUnbondingDelegationsByValidator(cdc),
			GetCmdQuerySideChainReDelegationsByValidator(cdc),
			GetCmdQuerySideChainTopValidators(cdc),
			GetCmdQuerySideAllValidatorsCount(cdc),
			GetCmdQueryCrossStakeInfoByBscAddress(cdc),
		)...,
	)

	root.AddCommand(stakingCmd)
}
