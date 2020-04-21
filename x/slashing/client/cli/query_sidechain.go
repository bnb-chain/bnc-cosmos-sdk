package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sidechain"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

// GetCmdQuerySideChainSigningInfo implements the command to query signing info.
func GetCmdQuerySideChainSigningInfo(storeName string, sideChainPrefixStoreName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bsc-signing-info [validator-sideConsAddr]",
		Short: "Query a validator's signing information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sideConsAddr, err := sdk.HexDecode(args[0])
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			sideChainStorePrefix, err := getSideChainStorePrefix(cliCtx, sideChainPrefixStoreName)
			if err != nil {
				return err
			}

			key := append(sideChainStorePrefix, slashing.GetValidatorSigningInfoKey(sdk.ConsAddress(sideConsAddr))...)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			signingInfo := new(slashing.ValidatorSigningInfo)
			cdc.MustUnmarshalBinaryLengthPrefixed(res, signingInfo)

			switch viper.Get(cli.OutputFlag) {

			case "text":
				human := signingInfo.HumanReadableString()
				fmt.Println(human)

			case "json":
				// parse out the signing info
				output, err := codec.MarshalJSONIndent(cdc, signingInfo)
				if err != nil {
					return err
				}
				fmt.Println(string(output))
			}

			return nil
		},
	}

	return cmd
}

func getSideChainStorePrefix(cliCtx context.CLIContext, storeName string) ([]byte, error) {
	sideChainId, err := getSideChainId()
	if err != nil {
		return nil, err
	}

	res, err := cliCtx.QueryStore(sidechain.GetSideChainStorePrefixKey(sideChainId), storeName)
	if err != nil {
		return nil, err
	} else if len(res) == 0 {
		return nil, fmt.Errorf("Invalid side-chain-id %s ", sideChainId)
	}
	return res, err
}
