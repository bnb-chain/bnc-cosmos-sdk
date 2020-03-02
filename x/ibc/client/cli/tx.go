package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/ibc/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

const (
	flagDestChainID     = "dest-chain-id"
	flagBEP2TokenSymbol = "bep2-symbol"
	flagBEP2TokenOwner  = "bep2-owner"
	flagContractAddr    = "contract-addr"
	flagTotalSupply     = "total-supply"
	flagPeggyAmount     = "peggy-amount"
	flagRelayReward     = "relay-reward"
	flagRefundAmount    = "refund-amount"
	flagRefundAddr      = "refund-addr"
	flagRecipient       = "recipient"
	flagExpireTime      = "expire-time"
	flagAmount          = "amount"
)

func IBCBindCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "bind",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			destChainID := viper.GetString(flagDestChainID)
			bep2TokenSymbol := viper.GetString(flagBEP2TokenSymbol)
			bep2TokenOwnerStr := viper.GetString(flagBEP2TokenOwner)
			bep2TokenOwner, err := sdk.AccAddressFromBech32(bep2TokenOwnerStr)
			if err != nil {
				return err
			}

			contractAddrStr := viper.GetString(flagContractAddr)
			if !strings.HasPrefix(contractAddrStr, "0x") {
				return fmt.Errorf("contract adderss must be prefix with 0x")
			}
			contractAddr, err := hex.DecodeString(contractAddrStr[2:])
			if err != nil {
				return err
			}

			totalSupply := viper.GetInt64(flagTotalSupply)
			peggyAmount := viper.GetInt64(flagPeggyAmount)
			relayReward := viper.GetInt64(flagRelayReward)

			channelID := ibc.BindChannelID
			packageBytes, err := client.SerializeBindPackage(bep2TokenSymbol, bep2TokenOwner, contractAddr, totalSupply, peggyAmount, relayReward)
			if err != nil {
				return err
			}

			msg := ibc.NewIBCPackage(from, destChainID, channelID, packageBytes)

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg})
			}

			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flagDestChainID, "", "destination chain-id")
	cmd.Flags().String(flagBEP2TokenSymbol, "", "bep2 token symbol")
	cmd.Flags().String(flagBEP2TokenOwner, "", "bep2 token owner")
	cmd.Flags().String(flagContractAddr, "", "ERC20 contract address")
	cmd.Flags().Int64(flagTotalSupply, 0, "bep2 token total supply")
	cmd.Flags().Int64(flagPeggyAmount, 0, "initial flowable amount on destination chain")
	cmd.Flags().Int64(flagRelayReward, 0, "reward for relayer")

	return cmd
}

func IBCTimeoutCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "timeout",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			destChainID := viper.GetString(flagDestChainID)
			refundAmount := viper.GetInt64(flagRefundAmount)
			refundAddrStr := viper.GetString(flagRefundAddr)
			if !strings.HasPrefix(refundAddrStr, "0x") {
				return fmt.Errorf("contract adderss must be prefix with 0x")
			}
			refundAddr, err := hex.DecodeString(refundAddrStr[2:])
			if err != nil {
				return err
			}
			contractAddrStr := viper.GetString(flagContractAddr)
			if !strings.HasPrefix(contractAddrStr, "0x") {
				return fmt.Errorf("contract adderss must be prefix with 0x")
			}
			contractAddr, err := hex.DecodeString(contractAddrStr[2:])
			if err != nil {
				return err
			}
			channelID := ibc.TimeoutChannelID
			packageBytes, err := client.SerializeTimeoutPackage(refundAmount, contractAddr, refundAddr)
			if err != nil {
				return err
			}

			msg := ibc.NewIBCPackage(from, destChainID, channelID, packageBytes)

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg})
			}

			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flagDestChainID, "", "destination chain-id")
	cmd.Flags().Int64(flagRefundAmount, 0, "refund amount")
	cmd.Flags().String(flagContractAddr, "", "ERC20 contract address")
	cmd.Flags().String(flagRefundAddr, "", "refund address")

	return cmd
}

func IBCTransferCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "transfer",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			destChainID := viper.GetString(flagDestChainID)

			bep2TokenSymbol := viper.GetString(flagBEP2TokenSymbol)
			contractAddrStr := viper.GetString(flagContractAddr)
			if !strings.HasPrefix(contractAddrStr, "0x") {
				return fmt.Errorf("contract adderss must be prefix with 0x")
			}
			contractAddr, err := hex.DecodeString(contractAddrStr[2:])
			if err != nil {
				return err
			}

			recipientStr := viper.GetString(flagRecipient)
			if !strings.HasPrefix(recipientStr, "0x") {
				return fmt.Errorf("contract adderss must be prefix with 0x")
			}
			recipient, err := hex.DecodeString(recipientStr[2:])
			if err != nil {
				return err
			}

			amount := viper.GetInt64(flagAmount)
			expireTime := viper.GetInt64(flagExpireTime)
			relayReward := viper.GetInt64(flagRelayReward)

			channelID := ibc.TransferChannelID
			packageBytes, err := client.SerializeTransferPackage(bep2TokenSymbol, contractAddr, from, recipient, amount, expireTime, relayReward)
			if err != nil {
				return err
			}

			msg := ibc.NewIBCPackage(from, destChainID, channelID, packageBytes)

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(txBldr, cliCtx, []sdk.Msg{msg})
			}

			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flagDestChainID, "", "destination chain-id")
	cmd.Flags().String(flagBEP2TokenSymbol, "", "bep2 token symbol")
	cmd.Flags().String(flagContractAddr, "", "ERC20 contract address")
	cmd.Flags().String(flagRecipient, "", "recipient address")
	cmd.Flags().Int64(flagAmount, 0, "transfer amount")
	cmd.Flags().Int64(flagExpireTime, 0, "expire time")
	cmd.Flags().Int64(flagRelayReward, 0, "reward for relayer")

	return cmd
}
