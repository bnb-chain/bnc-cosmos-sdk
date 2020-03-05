package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/ibc/client"
)

// nolint: unparam
func GetIBCPackageCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "query [dest-chain-id] [channel-name] [sequence]",
		Short: "Get IBC package",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			destChainID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			channelID, err := sdk.GetChannelID(args[1])
			if err != nil {
				return err
			}

			sequence, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			key := ibc.BuildIBCPackageKey(sdk.GetSourceChainID(), sdk.CrossChainID(destChainID), channelID, uint64(sequence))

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			path := "/store/ibc/key"
			res, err := client.QueryStore(cliCtx, path, key)
			if err != nil {
				return err
			} else if len(res.Value) == 0 {
				fmt.Println("IBC package doesn't exist")
				return nil
			}

			fmt.Println(hex.EncodeToString(res.Value))
			return nil
		},
	}
}
