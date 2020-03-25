package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func GetCmdQuerySideValidator(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-validator [operator-addr]",
		Short: "Query a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}
			key := append(sideChainStorePrefix, stake.GetValidatorKey(addr)...)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			} else if len(res) == 0 {
				return fmt.Errorf("No validator found with address %s", args[0])
			}

			validator, err := types.UnmarshalValidator(cdc, res)
			if err != nil {
				return err
			}

			switch viper.Get(cli.OutputFlag) {
			case "text":
				human, err := validator.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(human)

			case "json":
				// parse out the validator
				output, err := codec.MarshalJSONIndent(cdc, validator)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
			}

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsSideChainId)

	return cmd
}

func GetCmdQuerySideValidators(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-validators",
		Short: "Query for all validators",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			storeKeyPrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}
			key := append(storeKeyPrefix, stake.ValidatorsKey...)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var validators []stake.Validator
			for _, kv := range resKVs {
				validator, err := types.UnmarshalValidator(cdc, kv.Value)
				if err != nil {
					return err
				}
				validators = append(validators, validator)
			}

			switch viper.Get(cli.OutputFlag) {
			case "text":
				for _, validator := range validators {
					resp, err := validator.HumanReadableString()
					if err != nil {
						return err
					}

					fmt.Println(resp)
				}
			case "json":
				output, err := codec.MarshalJSONIndent(cdc, validators)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdQuerySideChainDelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-delegation",
		Short: "Query a delegation based on address and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, stake.GetDelegationKey(delAddr, valAddr)...)
			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the delegation
			delegation, err := types.UnmarshalDelegation(cdc, key, res)
			if err != nil {
				return err
			}

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := delegation.HumanReadableString()
				if err != nil {
					return err
				}

				fmt.Println(resp)
			case "json":
				output, err := codec.MarshalJSONIndent(cdc, delegation)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdQuerySideChainDelegations(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made from one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, stake.GetDelegationsKey(delegatorAddr)...)
			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var delegations []stake.Delegation
			for _, kv := range resKVs {
				delegation := types.MustUnmarshalDelegation(cdc, kv.Key, kv.Value)
				delegations = append(delegations, delegation)
			}

			output, err := codec.MarshalJSONIndent(cdc, delegations)
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQuerySideChainUnbondingDelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-unbonding-delegation",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, stake.GetUBDKey(delAddr, valAddr)...)
			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			ubd := types.MustUnmarshalUBD(cdc, key, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := ubd.HumanReadableString()
				if err != nil {
					return err
				}

				fmt.Println(resp)
			case "json":
				output, err := codec.MarshalJSONIndent(cdc, ubd)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

// GetCmdQueryUnbondingDelegations implements the command to query all the
// unbonding-delegation records for a delegator.
func GetCmdQuerySideChainUnbondingDelegations(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, stake.GetUBDsKey(delegatorAddr)...)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var ubds []stake.UnbondingDelegation
			for _, kv := range resKVs {
				ubd := types.MustUnmarshalUBD(cdc, kv.Key, kv.Value)
				ubds = append(ubds, ubd)
			}

			output, err := codec.MarshalJSONIndent(cdc, ubds)
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

// GetCmdQueryRedelegation implements the command to query a single
// redelegation record.
func GetCmdQuerySideChainRedelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-redelegation",
		Short: "Query a redelegation record based on delegator and a source and destination validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			valSrcAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidatorSrc))
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidatorDst))
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, stake.GetREDKey(delAddr, valSrcAddr, valDstAddr)...)
			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			red := types.MustUnmarshalRED(cdc, key, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := red.HumanReadableString()
				if err != nil {
					return err
				}

				fmt.Println(resp)
			case "json":
				output, err := codec.MarshalJSONIndent(cdc, red)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsRedelegation)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQuerySideChainRedelegations(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-redelegations [delegator-addr]",
		Short: "Query all redelegations records for one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, stake.GetREDsKey(delegatorAddr)...)
			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var reds []stake.Redelegation
			for _, kv := range resKVs {
				red := types.MustUnmarshalRED(cdc, kv.Key, kv.Value)
				reds = append(reds, red)
			}

			output, err := codec.MarshalJSONIndent(cdc, reds)
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdQuerySideChainPool(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool",
		Short: "Query the current staking pool values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, storeName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, stake.PoolKey...)
			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			pool := types.MustUnmarshalPool(cdc, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				human := pool.HumanReadableString()

				fmt.Println(human)

			case "json":
				// parse out the pool
				output, err := codec.MarshalJSONIndent(cdc, pool)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
			}
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func getSideChainStorePrefix(cliCtx context.CLIContext, storeName string) ([]byte, error) {
	sideChainId, err := getSideChainId()
	if err != nil {
		return nil, err
	}

	res, err := cliCtx.QueryStore(stake.GetSideChainStorePrefixKey(sideChainId), storeName)
	if err != nil {
		return nil, err
	} else if len(res) == 0 {
		return nil, fmt.Errorf("Invalid side-chain-id %s", sideChainId)
	}
	return res, err
}
