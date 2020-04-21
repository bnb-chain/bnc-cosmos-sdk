package cli

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strconv"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sidechain"
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
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
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
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
			if err != nil {
				return err
			}

			delegationKey := stake.GetDelegationKey(delAddr, valAddr)
			key := append(sideChainStorePrefix, delegationKey...)
			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the delegation
			delegation, err := types.UnmarshalDelegation(cdc, delegationKey, res)
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
		Use:   "side-delegations [delegator-addr]",
		Short: "Query all delegations made from one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
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
				k := kv.Key[len(sideChainStorePrefix):] // remove side chain prefix bytes
				delegation := types.MustUnmarshalDelegation(cdc, k, kv.Value)
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
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
			if err != nil {
				return err
			}

			ubdKey := stake.GetUBDKey(delAddr, valAddr)
			key := append(sideChainStorePrefix, ubdKey...)
			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			ubd := types.MustUnmarshalUBD(cdc, ubdKey, res)

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
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
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
				k := kv.Key[len(sideChainStorePrefix):] // remove side chain prefix bytes
				ubd := types.MustUnmarshalUBD(cdc, k, kv.Value)
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
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
			if err != nil {
				return err
			}

			redKey := stake.GetREDKey(delAddr, valSrcAddr, valDstAddr)
			key := append(sideChainStorePrefix, redKey...)
			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			red := types.MustUnmarshalRED(cdc, redKey, res)

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
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
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
				k := kv.Key[len(sideChainStorePrefix):]
				red := types.MustUnmarshalRED(cdc, k, kv.Value)
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

func GetCmdQuerySideChainUnbondingDelegationsByValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-val-unbonding-delegations [validator-addr]",
		Short: "Query all unbonding-delegations records for one validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainId, err := getSideChainId()
			if err != nil {
				return err
			}
			if err = checkSideChainId(cliCtx, sideChainId); err != nil {
				return err
			}

			params := stake.QueryValidatorParams{
				ValidatorAddr: valAddr,
				SideChainId:   sideChainId,
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			response, err := cliCtx.QueryWithData("custom/stake/validatorUnbondingDelegations", bz)
			if err != nil {
				return err
			}

			fmt.Println(string(response))
			return nil
		},
	}
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdQuerySideChainReDelegationsByValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-val-redelegations [validator-addr]",
		Short: "Query all redelegations records for one validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainId, err := getSideChainId()
			if err != nil {
				return err
			}
			if err = checkSideChainId(cliCtx, sideChainId); err != nil {
				return err
			}
			params := stake.QueryValidatorParams{
				ValidatorAddr: valAddr,
				SideChainId:   sideChainId,
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			response, err := cliCtx.QueryWithData("custom/stake/validatorRedelegations", bz)
			if err != nil {
				return err
			}

			fmt.Println(string(response))
			return nil
		},
	}
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdQuerySideChainPool(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-pool",
		Short: "Query the current staking pool values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx)
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

func GetCmdQuerySideChainTopValidators(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "side-top-validators",
		Short: "Query top N validators at current time",
		RunE: func(cmd *cobra.Command, args []string) error {
			topS := viper.GetString("top")
			top, err := strconv.Atoi(topS)
			if err != nil {
				return err
			}
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			if top > 100 || top < 1 {
				return errors.New("top must be between 1 and 100")
			}
			sideChainId, err := getSideChainId()
			if err != nil {
				return err
			}
			if err = checkSideChainId(cliCtx, sideChainId); err != nil {
				return err
			}
			params := stake.QuerySideTopValidatorsParams{
				Top:         top,
				SideChainId: sideChainId,
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			response, err := cliCtx.QueryWithData("custom/stake/sideTopValidators", bz)
			if err != nil {
				return err
			}

			fmt.Println(string(response))
			return nil
		},
	}
	cmd.Flags().AddFlagSet(fsSideChainId)
	cmd.Flags().String("top", "21", "")
	return cmd
}

func getSideChainStorePrefix(cliCtx context.CLIContext) ([]byte, error) {
	sideChainId, err := getSideChainId()
	if err != nil {
		return nil, err
	}

	res, err := cliCtx.QueryStore(sidechain.GetSideChainStorePrefixKey(sideChainId), scStoreKey)
	if err != nil {
		return nil, err
	} else if len(res) == 0 {
		return nil, fmt.Errorf("Invalid side-chain-id %s ", sideChainId)
	}
	return res, err
}

func checkSideChainId(cliCtx context.CLIContext, sideChainId string) error {
	res, err := cliCtx.QueryStore(sidechain.GetSideChainStorePrefixKey(sideChainId), scStoreKey)
	if err != nil {
		return err
	} else if len(res) == 0 {
		return fmt.Errorf("Invalid side-chain-id %s ", sideChainId)
	}
	return nil
}
