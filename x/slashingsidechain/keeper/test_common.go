package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"testing"
)

// create a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(types.MsgSubmitEvidence{}, "test/slashsc/SubmitEvidence", nil)

	codec.RegisterCrypto(cdc)

	return cdc
}


func CreateTestInput(t *testing.T) (sdk.Context, Keeper){
	keySlashing := sdk.NewKVStoreKey("slashingsidechain")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keySlashing, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	mode := sdk.RunTxModeDeliver

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, mode, log.NewNopLogger())
	cdc := MakeTestCodec()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	keeper := NewKeeper(cdc,keySlashing,nil,pk.Subspace(types.DefaultParamspace),types.DefaultCodespace)

	keeper.SetParams(ctx,types.DefaultParams())

	return ctx, keeper
}
