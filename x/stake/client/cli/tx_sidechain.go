package cli

import (
	stdctx "context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/bsc"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/stake"
	sTypes "github.com/cosmos/cosmos-sdk/x/stake/types"

	"github.com/prysmaticlabs/prysm/v4/crypto/bls"
	"github.com/prysmaticlabs/prysm/v4/crypto/bls/common"
	validatorpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1/validator-client"
	"github.com/prysmaticlabs/prysm/v4/validator/accounts/iface"
	"github.com/prysmaticlabs/prysm/v4/validator/accounts/wallet"
	"github.com/prysmaticlabs/prysm/v4/validator/keymanager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetCmdCreateSideChainValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bsc-create-validator",
		Short: "create new validator for side chain initialized with a self-delegation to it",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
		cliCtx := context.NewCLIContext().
			WithCodec(cdc).
			WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

		amountStr := viper.GetString(FlagAmount)
		if amountStr == "" {
			return fmt.Errorf("Must specify amount to stake using --amount")
		}
		amount, err := sdk.ParseCoin(amountStr)
		if err != nil {
			return err
		}

		valAddr, err := cliCtx.GetFromAddress()
		if err != nil {
			return err
		}

		if viper.GetString(FlagMoniker) == "" {
			return fmt.Errorf("please enter a moniker for the validator using --moniker")
		}

		description := stake.Description{
			Moniker:  viper.GetString(FlagMoniker),
			Identity: viper.GetString(FlagIdentity),
			Website:  viper.GetString(FlagWebsite),
			Details:  viper.GetString(FlagDetails),
		}

		// get the initial validator commission parameters
		rateStr := viper.GetString(FlagCommissionRate)
		maxRateStr := viper.GetString(FlagCommissionMaxRate)
		maxChangeRateStr := viper.GetString(FlagCommissionMaxChangeRate)
		commissionMsg, err := buildCommissionMsg(rateStr, maxRateStr, maxChangeRateStr)
		if err != nil {
			return err
		}

		sideChainId, sideConsAddr, sideFeeAddr, sideVoteAddr, err := getSideChainInfo(true, true)
		if err != nil {
			return err
		}

		var msg sdk.Msg
		if sideVoteAddr != nil {
			if viper.GetString(FlagAddressDelegator) != "" {
				delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
				if err != nil {
					return err
				}

				msg = stake.NewMsgCreateSideChainValidatorWithVoteAddrOnBehalfOf(delAddr, sdk.ValAddress(valAddr), amount, description,
					commissionMsg, sideChainId, sideConsAddr, sideFeeAddr, sideVoteAddr)
			} else {
				msg = stake.NewMsgCreateSideChainValidatorWithVoteAddr(
					sdk.ValAddress(valAddr), amount, description, commissionMsg, sideChainId, sideConsAddr, sideFeeAddr, sideVoteAddr)
			}
		} else {
			if viper.GetString(FlagAddressDelegator) != "" {
				delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
				if err != nil {
					return err
				}

				msg = stake.NewMsgCreateSideChainValidatorOnBehalfOf(delAddr, sdk.ValAddress(valAddr), amount, description,
					commissionMsg, sideChainId, sideConsAddr, sideFeeAddr)
			} else {
				msg = stake.NewMsgCreateSideChainValidator(
					sdk.ValAddress(valAddr), amount, description, commissionMsg, sideChainId, sideConsAddr, sideFeeAddr)
			}
		}

		return utils.GenerateOrBroadcastMsgs(txBldr, cliCtx, []sdk.Msg{msg})
	}

	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsDescriptionCreate)
	cmd.Flags().AddFlagSet(fsCommissionCreate)
	cmd.Flags().AddFlagSet(fsDelegator)
	cmd.Flags().AddFlagSet(fsSideChainFull)
	cmd.MarkFlagRequired(client.FlagFrom)
	return cmd
}

func GetCmdEditSideChainValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bsc-edit-validator",
		Short: "edit an existing side chain validator",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
		cliCtx := context.NewCLIContext().
			WithCodec(cdc).
			WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

		valAddr, err := cliCtx.GetFromAddress()
		if err != nil {
			return err
		}

		description := stake.Description{
			Moniker:  viper.GetString(FlagMoniker),
			Identity: viper.GetString(FlagIdentity),
			Website:  viper.GetString(FlagWebsite),
			Details:  viper.GetString(FlagDetails),
		}

		var newRate *sdk.Dec
		commissionRate := viper.GetString(FlagCommissionRate)
		if commissionRate != "" {
			rate, err := sdk.NewDecFromStr(commissionRate)
			if err != nil {
				return fmt.Errorf("invalid new commission rate: %v", err)
			}

			newRate = &rate
		}

		sideChainId, sideConsAddr, sideFeeAddr, sideVoteAddr, err := getSideChainInfo(false, false)
		if err != nil {
			return err
		}

		var msg sdk.Msg
		if sideVoteAddr != nil {
			msg = stake.NewMsgEditSideChainValidatorWithVoteAddr(sideChainId, sdk.ValAddress(valAddr), description, newRate, sideFeeAddr, sideConsAddr, sideVoteAddr)
		} else {
			msg = stake.NewMsgEditSideChainValidator(sideChainId, sdk.ValAddress(valAddr), description, newRate, sideFeeAddr, sideConsAddr)
		}
		return utils.GenerateOrBroadcastMsgs(txBldr, cliCtx, []sdk.Msg{msg})
	}

	cmd.Flags().AddFlagSet(fsDescriptionEdit)
	cmd.Flags().AddFlagSet(fsCommissionUpdate)
	cmd.Flags().AddFlagSet(fsSideChainEdit)
	return cmd
}

func GetCmdSideChainDelegate(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bsc-delegate",
		Short: "delegate liquid tokens to a side chain validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			amount, err := getAmount()
			if err != nil {
				return err
			}

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			valAddr, err := getValidatorAddr(FlagAddressValidator)
			if err != nil {
				return err
			}

			sideChainId, err := getSideChainId()
			if err != nil {
				return err
			}

			msg := stake.NewMsgSideChainDelegate(sideChainId, delAddr, valAddr, amount)
			return utils.GenerateOrBroadcastMsgs(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdSideChainRedelegate(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bsc-redelegate",
		Short: "Redelegate illiquid tokens from one validator to another",

		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			valSrcAddr, err := getValidatorAddr(FlagAddressValidatorSrc)
			if err != nil {
				return err
			}

			valDstAddr, err := getValidatorAddr(FlagAddressValidatorDst)
			if err != nil {
				return err
			}

			amount, err := getAmount()
			if err != nil {
				return err
			}

			sideChainId, err := getSideChainId()
			if err != nil {
				return err
			}

			msg := stake.NewMsgSideChainRedelegate(sideChainId, delAddr, valSrcAddr, valDstAddr, amount)
			return utils.GenerateOrBroadcastMsgs(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsRedelegation)
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdSideChainUnbond(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bsc-unbond",
		Short: "Undelegate illiquid tokens from the validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}
			valAddr, err := getValidatorAddr(FlagAddressValidator)
			if err != nil {
				return err
			}

			amount, err := getAmount()
			if err != nil {
				return err
			}

			sideChainId, err := getSideChainId()
			if err != nil {
				return err
			}

			msg := stake.NewMsgSideChainUndelegate(sideChainId, delAddr, valAddr, amount)
			return utils.GenerateOrBroadcastMsgs(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsSideChainId)
	return cmd
}

func GetCmdSideChainStakeMigration(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bsc-stake-migration",
		Short: "Migrate delegation from Beacon Chain to Smart Chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			valAddr, err := getValidatorAddr(FlagAddressValidator)
			if err != nil {
				return err
			}
			operatorAddr, err := getSmartChainAddr(FlagAddressSmartChainOperator)
			if err != nil {
				return err
			}
			delAddr, err := getSmartChainAddr(FlagAddressSmartChainDelegator)
			if err != nil {
				return err
			}
			refundAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			amount, err := getAmount()
			if err != nil {
				return err
			}

			msg := sTypes.NewMsgSideChainStakeMigration(valAddr, operatorAddr, delAddr, refundAddr, amount)
			return utils.GenerateOrBroadcastMsgs(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsSmartChainOperator)
	cmd.Flags().AddFlagSet(fsSmartChainDelegator)

	return cmd
}

func getSideChainId() (sideChainId string, err error) {
	sideChainId = viper.GetString(FlagSideChainId)
	if len(sideChainId) == 0 {
		err = fmt.Errorf("%s is required", FlagSideChainId)
	}
	return
}

func getSideChainInfo(requireConsAddr, requireFeeAddr bool) (sideChainId string, sideConsAddr, sideFeeAddr, sideVoteAddr []byte, err error) {
	sideChainId, err = getSideChainId()
	if err != nil {
		return
	}

	sideConsAddrStr := viper.GetString(FlagSideConsAddr)
	if len(sideConsAddrStr) == 0 {
		if requireConsAddr {
			err = fmt.Errorf("%s is required", FlagSideConsAddr)
			return
		}
	} else {
		sideConsAddr, err = sdk.HexDecode(sideConsAddrStr)
		if err != nil {
			return
		}
	}

	sideFeeAddrStr := viper.GetString(FlagSideFeeAddr)
	if len(sideFeeAddrStr) == 0 {
		if requireFeeAddr {
			err = fmt.Errorf("%s is required", FlagSideFeeAddr)
			return
		}
	} else {
		sideFeeAddr, err = sdk.HexDecode(sideFeeAddrStr)
		if err != nil {
			return
		}
	}

	sideVoteAddrStr := viper.GetString(FlagSideVoteAddr)
	if len(sideVoteAddrStr) != 0 {
		// check vote addr
		sideVoteAddr, err = sdk.HexDecode(sideVoteAddrStr)
		if err != nil {
			return
		}
		if len(sideVoteAddr) != types.VoteAddrLen {
			err = fmt.Errorf("Expected SideVoteAddr length is 48, got %d ", len(sideVoteAddr))
			return
		}
		var voteKey bls.PublicKey
		voteKey, err = bls.PublicKeyFromBytes(sideVoteAddr)
		if err != nil {
			err = fmt.Errorf("Invalid side vote addr")
			return
		}

		// open bls wallet
		blsWalletDir := viper.GetString(FlagBLSWalletDir)
		if len(blsWalletDir) == 0 {
			err = fmt.Errorf("Path of BLS wallet which containing the given side vote address should be provided")
			return
		}
		var passphrase string
		passphrase, err = getBLSPassword()
		if err != nil {
			return
		}
		var w *wallet.Wallet
		w, err = wallet.OpenWallet(stdctx.Background(), &wallet.Config{
			WalletDir:      blsWalletDir,
			WalletPassword: passphrase,
		})
		if err != nil {
			err = fmt.Errorf("Open BLS wallet failed")
			return
		}

		// generate proof of possession
		var km keymanager.IKeymanager
		km, err = w.InitializeKeymanager(stdctx.Background(), iface.InitKeymanagerConfig{ListenForChanges: false})
		if err != nil {
			err = fmt.Errorf("Initialize key manager failed: %v", err)
			return
		}
		signingRoot := bsc.Keccak256(append(voteKey.Marshal(), []byte(sideChainId)...)) // here sideChainId used as `domain` in bls spec
		voteSignerTimeout := time.Second * 5
		ctx, cancel := stdctx.WithTimeout(stdctx.Background(), voteSignerTimeout)
		defer cancel()
		var signature common.Signature
		signature, err = km.Sign(ctx, &validatorpb.SignRequest{
			PublicKey:   []byte(sideVoteAddr),
			SigningRoot: signingRoot,
		})
		if err != nil {
			return
		}
		sideVoteAddr = append(sideVoteAddr, signature.Marshal()...)
	}
	return
}

func getValidatorAddr(flagName string) (valAddr sdk.ValAddress, err error) {
	valAddrStr := viper.GetString(flagName)
	if len(valAddrStr) == 0 {
		err = fmt.Errorf("%s is required", flagName)
		return
	}
	return sdk.ValAddressFromBech32(valAddrStr)
}

func getSmartChainAddr(flagName string) (addr sdk.SmartChainAddress, err error) {
	addrStr := viper.GetString(flagName)
	if len(addrStr) == 0 {
		err = fmt.Errorf("%s is required", flagName)
		return
	}
	return sdk.NewSmartChainAddress(addrStr)
}

func getBLSPassword() (string, error) {
	blsPassword := viper.GetString(FlagBLSPassword)
	if len(blsPassword) > 0 {
		return blsPassword, nil
	}
	return readPassphraseFromStdin()
}

func readPassphraseFromStdin() (string, error) {
	buf := client.BufferStdin()
	prompt := "Password to open bls wallet:"
	passphrase, err := client.GetPasswordWithoutCheck(prompt, buf)
	if err != nil {
		return passphrase, fmt.Errorf("Error reading passphrase: %v", err)
	}
	return passphrase, nil
}
