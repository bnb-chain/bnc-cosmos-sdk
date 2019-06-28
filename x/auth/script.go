package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Script func(ctx sdk.Context, tx sdk.Msg) sdk.Error

var scriptsHub = map[string][]Script{}

func RegisterScripts(msgType string, scripts ...Script) {
	scriptsHub[msgType] = append(scriptsHub[msgType], scripts...)
}

func GetRegisteredScripts(msgType string) []Script {
	return scriptsHub[msgType]
}
