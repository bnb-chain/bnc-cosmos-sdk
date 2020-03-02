package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/ibc/client"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

// nolint: unparam
func GetIBCPackageCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "query [source-chain-id] [dest-chain-id] [channel-id] [sequence]",
		Short: "Get IBC package",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceChainID := args[0]
			destChainID := args[1]
			channelID, err := ibc.NameToChannelID(args[2])
			if err != nil {
				return err
			}

			sequence, err := strconv.Atoi(args[3])
			if err != nil {
				return err
			}

			key := ibc.BuildIBCPackageKey(sourceChainID, destChainID, channelID, int64(sequence))

			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

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
