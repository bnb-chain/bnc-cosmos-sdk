package utils

import (
	"bytes"
	"fmt"
	"os"

	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

// CompleteAndBroadcastTxCli implements a utility function that
// facilitates sending a series of messages in a signed
// transaction given a TxBuilder and a QueryContext. It ensures
// that the account exists, has a proper number and sequence
// set. In addition, it builds and signs a transaction with the
// supplied messages.  Finally, it broadcasts the signed
// transaction to a node.
// NOTE: Also see CompleteAndBroadcastTxREST.
func CompleteAndBroadcastTxCli(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) error {
	txBldr, err := prepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return err
	}

	name, err := cliCtx.GetFromName()
	if err != nil {
		return err
	}

	if cliCtx.DryRun {
		return nil
	}

	passphrase, err := keys.GetPassphrase(name)
	if err != nil {
		return err
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(name, passphrase, msgs)
	if err != nil {
		return err
	}
	// broadcast to a Tendermint node
	_, err = cliCtx.BroadcastTx(txBytes)
	return err
}

// PrintUnsignedStdTx builds an unsigned StdTx and prints it to os.Stdout.
// Don't perform online validation or lookups if offline is true.
func PrintUnsignedStdTx(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg, offline bool) (err error) {
	var stdTx auth.StdTx
	if offline {
		stdTx, err = buildUnsignedStdTxOffline(txBldr, msgs)
	} else {
		stdTx, err = buildUnsignedStdTx(txBldr, cliCtx, msgs)
	}
	if err != nil {
		return
	}
	json, err := txBldr.Codec.MarshalJSON(stdTx)
	if err == nil {
		fmt.Printf("%s\n", json)
	}
	return
}

// SignStdTx appends a signature to a StdTx and returns a copy of a it. If appendSig
// is false, it replaces the signatures already attached with the new signature.
// Don't perform online validation or lookups if offline is true.
func SignStdTx(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, name string, stdTx auth.StdTx, appendSig bool, offline bool) (auth.StdTx, error) {
	var signedStdTx auth.StdTx

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return signedStdTx, err
	}
	info, err := keybase.Get(name)
	if err != nil {
		return signedStdTx, err
	}
	addr := info.GetPubKey().Address()

	// Check whether the address is a signer
	if !isTxSigner(sdk.AccAddress(addr), stdTx.GetSigners()) {
		fmt.Fprintf(os.Stderr, "WARNING: The generated transaction's intended signer does not match the given signer: '%v'\n", name)
	}

	if !offline && txBldr.AccountNumber == 0 {
		accNum, err := cliCtx.GetAccountNumber(addr)
		if err != nil {
			return signedStdTx, err
		}
		txBldr = txBldr.WithAccountNumber(accNum)
	}

	if !offline && txBldr.Sequence == 0 {
		accSeq, err := cliCtx.GetAccountSequence(addr)
		if err != nil {
			return signedStdTx, err
		}
		txBldr = txBldr.WithSequence(accSeq)
	}

	passphrase, err := keys.GetPassphrase(name)
	if err != nil {
		return signedStdTx, err
	}
	return txBldr.SignStdTx(name, passphrase, stdTx, appendSig)
}

func parseQueryResponse(cdc *amino.Codec, rawRes []byte) error {
	var simulationResult sdk.Result
	if err := cdc.UnmarshalBinaryLengthPrefixed(rawRes, &simulationResult); err != nil {
		return err
	}
	return nil
}

func prepareTxBuilder(txBldr authtxb.TxBuilder, cliCtx context.CLIContext) (authtxb.TxBuilder, error) {
	if err := cliCtx.EnsureAccountExists(); err != nil {
		return txBldr, err
	}

	from, err := cliCtx.GetFromAddress()
	if err != nil {
		return txBldr, err
	}

	// TODO: (ref #1903) Allow for user supplied account number without
	// automatically doing a manual lookup.
	if txBldr.AccountNumber == 0 {
		accNum, err := cliCtx.GetAccountNumber(from)
		if err != nil {
			return txBldr, err
		}
		txBldr = txBldr.WithAccountNumber(accNum)
	}

	// TODO: (ref #1903) Allow for user supplied account sequence without
	// automatically doing a manual lookup.
	if txBldr.Sequence == 0 {
		accSeq, err := cliCtx.GetAccountSequence(from)
		if err != nil {
			return txBldr, err
		}
		txBldr = txBldr.WithSequence(accSeq)
	}
	return txBldr, nil
}

// buildUnsignedStdTx builds a StdTx as per the parameters passed in the
// contexts.
func buildUnsignedStdTx(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) (stdTx auth.StdTx, err error) {
	txBldr, err = prepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return
	}
	return buildUnsignedStdTxOffline(txBldr, msgs)
}

func buildUnsignedStdTxOffline(txBldr authtxb.TxBuilder, msgs []sdk.Msg) (stdTx auth.StdTx, err error) {
	stdSignMsg, err := txBldr.Build(msgs)
	if err != nil {
		return
	}
	return auth.NewStdTx(stdSignMsg.Msgs, nil, stdSignMsg.Memo, stdSignMsg.Source, nil), nil
}

func isTxSigner(user sdk.AccAddress, signers []sdk.AccAddress) bool {
	for _, s := range signers {
		if bytes.Equal(user.Bytes(), s.Bytes()) {
			return true
		}
	}
	return false
}
