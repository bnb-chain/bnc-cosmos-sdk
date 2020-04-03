package slashing

import (
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// TODO remove dependencies on staking (should only refer to validator set type from sdk)

var (
	pks = []crypto.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	}
	addrs = []sdk.ValAddress{
		sdk.ValAddress(pks[0].Address()),
		sdk.ValAddress(pks[1].Address()),
		sdk.ValAddress(pks[2].Address()),
	}
	initCoins = sdk.NewDecWithoutFra(200).RawInt()
)

func createTestCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	stake.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

func getAccountCache(cdc *codec.Codec, ms sdk.MultiStore, accountKey *sdk.KVStoreKey) sdk.AccountCache {
	accountStore := ms.GetKVStore(accountKey)
	accountStoreCache := auth.NewAccountStoreCache(cdc, accountStore, 10)
	return auth.NewAccountCache(accountStoreCache)
}

func createTestInput(t *testing.T, defaults Params) (sdk.Context, bank.Keeper, stake.Keeper, params.Subspace, Keeper) {
	keyAcc := sdk.NewKVStoreKey("acc")
	keyStake := sdk.NewKVStoreKey("stake")
	tkeyStake := sdk.NewTransientStoreKey("transient_stake")
	keySlashing := sdk.NewKVStoreKey("slashing")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")
	keyIbc := sdk.NewKVStoreKey("ibc")
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStake, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySlashing, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyIbc, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)
	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(0, 0)}, sdk.RunTxModeDeliver, log.NewTMLogger(os.Stdout))
	cdc := createTestCodec()
	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, auth.ProtoBaseAccount)
	accountCache := getAccountCache(cdc, ms, keyAcc)
	ctx = ctx.WithAccountCache(accountCache)

	ck := bank.NewBaseKeeper(accountKeeper)
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams)
	ibcKeeper := ibc.NewKeeper(keyIbc, ibc.DefaultCodespace)
	sk := stake.NewKeeper(cdc, keyStake, tkeyStake, ck, ibcKeeper, nil, paramsKeeper.Subspace(stake.DefaultParamspace), stake.DefaultCodespace)
	genesis := stake.DefaultGenesisState()

	genesis.Pool.LooseTokens = sdk.NewDec(initCoins * (int64(len(addrs))))

	_, err = stake.InitGenesis(ctx, sk, genesis)
	require.Nil(t, err)

	for _, addr := range addrs {
		_, _, err = ck.AddCoins(ctx, sdk.AccAddress(addr), sdk.Coins{
			{sk.GetParams(ctx).BondDenom, initCoins},
		})
	}
	require.Nil(t, err)
	paramstore := paramsKeeper.Subspace(DefaultParamspace)
	keeper := NewKeeper(cdc, keySlashing, sk, paramstore, DefaultCodespace)
	sk = sk.WithHooks(keeper.Hooks())

	require.NotPanics(t, func() {
		InitGenesis(ctx, keeper, GenesisState{defaults}, genesis)
	})

	return ctx, ck, sk, paramstore, keeper
}

func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

func testAddr(addr string) sdk.AccAddress {
	res := []byte(addr)
	return res
}

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey crypto.PubKey, amt int64) stake.MsgCreateValidator {
	commission := stake.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	return stake.MsgCreateValidator{
		Description:   stake.Description{},
		Commission:    commission,
		DelegatorAddr: sdk.AccAddress(address),
		ValidatorAddr: address,
		PubKey:        pubKey,
		Delegation:    sdk.NewCoin("steak", amt),
	}
}

func newTestMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, delAmount int64) stake.MsgDelegate {
	return stake.MsgDelegate{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Delegation:    sdk.NewCoin("steak", delAmount),
	}
}
