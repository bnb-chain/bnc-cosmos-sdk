package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/sidechain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
)

const (
	flagEvidence     = "evidence"
	flagEvidenceFile = "evidence-file"

	flagSideChainId  = "side-chain-id"
)

// GetCmdSubmitEvidence implements the submit evidence command handler.
func GetCmdSubmitEvidence(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-evidence",
		Short: "submit evidence against the malicious validator on side chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			// get the from/to address
			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			filePath := viper.GetString(flagEvidenceFile)
			evidenceBytes := make([]byte, 0)
			if filePath != "" {
				evidenceBytes, err = ioutil.ReadFile(filePath)
				if err != nil {
					return err
				}
			} else {
				txStr := viper.GetString(flagEvidence)
				if txStr == "" {
					return errors.New(fmt.Sprintf("either %s or %s is required",flagEvidenceFile,flagEvidence))
				}
				evidenceBytes = []byte(txStr)
			}

			headers := [2]*sidechain.Header{}
			err = json.Unmarshal(evidenceBytes, &headers)
			if err != nil {
				return err
			}

			sideChainId,err := getSideChainId()
			if err != nil {
				return err
			}

			msg := slashingsidechain.NewMsgSubmitEvidence(from,sideChainId,headers)
			bytes ,err := json.Marshal(msg)
			fmt.Println(string(bytes))

			return utils.GenerateOrBroadcastMsgs(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagSideChainId,"","Chain-id of the side chain the validator belongs to")
	cmd.Flags().String(flagEvidence,"","Evidence details, including two headers with json format, e.g. [{\"difficulty\":\"0x2\",\"extraData\":\"0xd98301...},{\"difficulty\":\"0x3\",\"extraData\":\"0xd64372...}]")
	cmd.Flags().String(flagEvidenceFile,"","File of evidence details, if evidence-file is not empty, --evidence will be ignored")
	cmd.MarkFlagRequired(flagSideChainId)
	return cmd
}

func getSideChainId() (sideChainId string, err error) {
	sideChainId = viper.GetString(flagSideChainId)
	if len(sideChainId) == 0 {
		err = fmt.Errorf("%s is required", flagSideChainId)
	}
	return
}
