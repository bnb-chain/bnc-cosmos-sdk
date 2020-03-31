package slashingsidechain

import (
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/types"
)

type (
	Keeper = keeper.Keeper
	Params = types.Params
	MsgSubmitEvidence = types.MsgSubmitEvidence
)

var (
	NewKeeper            = keeper.NewKeeper
	DefaultParamspace    = types.DefaultParamspace
	RegisterCodec        = types.RegisterCodec
	NewMsgSubmitEvidence = types.NewMsgSubmitEvidence
)

const DefaultCodespace = types.DefaultCodespace
