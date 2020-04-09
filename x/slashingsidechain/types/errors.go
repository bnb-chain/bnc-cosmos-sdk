package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = 9

	CodeExpiredEvidence CodeType = 101

	CodeFailSlash CodeType = 102

	CodeHandledEvidence CodeType = 103
)

func ErrExpiredEvidence(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeExpiredEvidence, "The given evidences are expired")
}

func ErrFailedToSlash(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeFailSlash, fmt.Sprintf("failed to slash, %s", msg))
}

func ErrEvidenceHasBeenHandled(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeHandledEvidence, "The evidence has been handled")
}
