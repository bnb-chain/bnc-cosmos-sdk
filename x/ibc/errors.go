package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC errors reserve 200 ~ 299.
const (
	DefaultCodespace sdk.CodespaceType = 3

	CodeUnsupportedChannel sdk.CodeType = 101
	CodeChainIDTooLong     sdk.CodeType = 102
	CodeEmptyPackage       sdk.CodeType = 103
)

func ErrUnsupportedChannel(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeUnsupportedChannel, msg)
}

func ErrChainIDTooLong(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeChainIDTooLong, msg)
}

func ErrEmptyPackage(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyPackage, msg)
}
