package ibc

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout sdk
var msgCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	msgCdc = cdc.Seal()
}

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(IBCPackageMsg{}, "cosmos-sdk/IBCPackageMsg", nil)
}
