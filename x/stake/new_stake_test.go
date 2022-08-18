// test BEPHHH: new stake mechanism
package stake

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/gov"
	stake "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/sidechain"
)

func getBeginBlocker(keeper Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
		sdk.UpgradeMgr.BeginBlocker(ctx)
		return
	}
}

func getNewInitChainer(mapp *mock.App, keeper Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakeGenesis := DefaultGenesisState()
		stakeGenesis.Params.BondDenom = "BNB"
		stakeGenesis.Pool.LooseTokens = sdk.NewDecWithoutFra(100000)

		validators, err := InitGenesis(ctx, keeper, stakeGenesis)
		if err != nil {
			panic(err)
		}

		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

func getNewEndBlocker(keeper Keeper, breatheBlockInterval int) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		var validatorUpdates []abci.ValidatorUpdate
		if ctx.BlockHeader().Height%int64(breatheBlockInterval) != 0 {
			validatorUpdates, _ = EndBlocker(ctx, keeper)
		} else {
			validatorUpdates, _ = EndBreatheBlock(ctx, keeper)
		}

		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	}
}

func getNewStakeMockApp(t *testing.T) (*mock.App, Keeper) {
	mApp := mock.NewApp()

	RegisterCodec(mApp.Cdc)

	keyStake := sdk.NewKVStoreKey("stake")
	keyStakeReward := sdk.NewKVStoreKey("stake_reward")
	tkeyStake := sdk.NewTransientStoreKey("transient_stake")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")
	keyIbc := sdk.NewKVStoreKey("ibc")
	keySideChain := sdk.NewKVStoreKey("sc")
	keyGov := sdk.NewKVStoreKey("gov")

	bankKeeper := bank.NewBaseKeeper(mApp.AccountKeeper)
	paramsKeeper := params.NewKeeper(mApp.Cdc, keyParams, tkeyParams)
	scKeeper := sidechain.NewKeeper(keySideChain, paramsKeeper.Subspace(sidechain.DefaultParamspace), mApp.Cdc)
	ibcKeeper := ibc.NewKeeper(keyIbc, paramsKeeper.Subspace(ibc.DefaultParamspace), ibc.DefaultCodespace, scKeeper)
	keeper := NewKeeper(mApp.Cdc, keyStake, keyStakeReward, tkeyStake, bankKeeper, nil, paramsKeeper.Subspace(DefaultParamspace), mApp.RegisterCodespace(DefaultCodespace))
	govKeeper := gov.NewKeeper(mApp.Cdc, keyGov, paramsKeeper, paramsKeeper.Subspace(gov.DefaultParamSpace), bankKeeper, keeper, mApp.RegisterCodespace(DefaultCodespace), nil)
	keeper.SetupForSideChain(&scKeeper, &ibcKeeper)

	sdk.UpgradeMgr.AddUpgradeHeight(sdk.LaunchBscUpgrade, 6)
	sdk.UpgradeMgr.AddUpgradeHeight(sdk.BEP128, 7)
	sdk.UpgradeMgr.AddUpgradeHeight(sdk.BEPHHH, 8)
	BscChainId := "bsc"
	sdk.UpgradeMgr.RegisterBeginBlocker(sdk.LaunchBscUpgrade, func(ctx sdk.Context) {
		MigratePowerRankKey(ctx, keeper)
		storePrefix := scKeeper.GetSideChainStorePrefix(ctx, BscChainId)
		newCtx := ctx.WithSideChainKeyPrefix(storePrefix)
		keeper.SetParams(newCtx, stake.Params{
			UnbondingTime:       60 * 60 * 24 * 7 * time.Second, // 7 days
			MaxValidators:       21,
			BondDenom:           "BNB",
			MinSelfDelegation:   20000e8,
			MinDelegationChange: 1e8,
		})
		keeper.SetPool(newCtx, stake.Pool{
			LooseTokens: sdk.NewDec(5e15),
		})
	})
	sdk.UpgradeMgr.RegisterBeginBlocker(sdk.BEP128, func(ctx sdk.Context) {
		storePrefix := scKeeper.GetSideChainStorePrefix(ctx, BscChainId)
		// init new param RewardDistributionBatchSize
		newCtx := ctx.WithSideChainKeyPrefix(storePrefix)
		params := keeper.GetParams(newCtx)
		params.RewardDistributionBatchSize = 1000
		keeper.SetParams(newCtx, params)
	})
	sdk.UpgradeMgr.RegisterBeginBlocker(sdk.BEPHHH, func(ctx sdk.Context) {
		storePrefix := scKeeper.GetSideChainStorePrefix(ctx, BscChainId)
		newCtx := ctx.WithSideChainKeyPrefix(storePrefix)
		params := keeper.GetParams(newCtx)
		params.MaxStakeSnapshots = 30
		params.MaxValidators = 11
		keeper.SetParams(ctx, params)
	})
	mApp.Router().AddRoute("stake", NewHandler(keeper, govKeeper))
	mApp.SetBeginBlocker(getBeginBlocker(keeper))
	mApp.SetEndBlocker(getNewEndBlocker(keeper, 5))
	mApp.SetInitChainer(getNewInitChainer(mApp, keeper))

	require.NoError(t, mApp.CompleteSetup(keyStake, keyStakeReward, tkeyStake, keyParams, tkeyParams, keyIbc, keySideChain, keyGov))
	return mApp, keeper
}

type Account struct {
	Priv        crypto.PrivKey
	Address     sdk.AccAddress
	BaseAccount *auth.BaseAccount
}

func GenAccounts(n int) (accounts []Account) {
	for i := 0; i < n; i++ {
		priv := ed25519.GenPrivKey()
		address := sdk.AccAddress(priv.PubKey().Address())
		genCoin := sdk.NewCoin("BNB", sdk.NewDecWithoutFra(12345678).RawInt())
		baseAccount := auth.BaseAccount{
			Address: address,
			Coins:   sdk.Coins{genCoin},
		}
		accounts = append(accounts, Account{
			Priv:        priv,
			Address:     address,
			BaseAccount: &baseAccount,
		})
	}
	return
}

func TestNewStake(t *testing.T) {
	mApp, keeper := getNewStakeMockApp(t)

	genCoin := sdk.NewCoin("BNB", sdk.NewDecWithoutFra(42).RawInt())
	bondCoin := sdk.NewCoin("BNB", sdk.NewDecWithoutFra(10).RawInt())

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []sdk.Account{acc1, acc2}
	accounts := GenAccounts(100)
	fmt.Printf("acc1: %+v, accounts[0]: %+v\n", acc1, accounts[0])
	for _, acc := range accounts {
		accs = append(accs, acc.BaseAccount)
		mApp.Logger.Debug("add genesis account", "account", acc)
	}
	mApp.Logger.Debug("add genesis accounts", "accounts", accs)

	mock.SetGenesis(mApp, accs)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin})
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})

	// create validator
	description := NewDescription("foo_moniker", "", "", "")
	createValidatorMsg := NewMsgCreateValidator(
		sdk.ValAddress(addr1), priv1.PubKey(), bondCoin, description, commissionMsg,
	)

	var height int64 = 1
	txs := mock.GenSimTxs(t, mApp, []sdk.Msg{createValidatorMsg}, true, priv1)
	height = mock.ApplyBlock(t, mApp.BaseApp, height, txs)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin.Minus(bondCoin)})

	validator := checkValidator(t, mApp, keeper, sdk.ValAddress(addr1), true)
	require.Equal(t, sdk.ValAddress(addr1), validator.OperatorAddr)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.DecEq(t, sdk.NewDecWithoutFra(10), validator.BondedTokens()))

	ctx := mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators := keeper.GetLastValidators(ctx)
	fmt.Printf("%+v\n", validators)

	// create validator2
	description2 := NewDescription("foo_moniker", "", "", "")
	createValidatorMsg2 := NewMsgCreateValidator(
		sdk.ValAddress(addr2), priv2.PubKey(), bondCoin, description2, commissionMsg,
	)
	txs = mock.GenSimTxs(t, mApp, []sdk.Msg{createValidatorMsg2}, true, priv2)
	height = mock.ApplyBlock(t, mApp.BaseApp, height, txs)
	ctx = mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators = keeper.GetLastValidators(ctx)
	fmt.Printf("%+v\n", validators)

	// hardfork
	height = mock.ApplyEmptyBlocks(t, mApp.BaseApp, height, 200)
	ctx = mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators = keeper.GetLastValidators(ctx)
	fmt.Printf("%+v\n", validators)

	// fail to create validator after hardfork, self delegation not enough
	acc := accounts[0]
	description3 := NewDescription("validator3", "", "", "")
	createValidatorMsg3 := NewMsgCreateValidator(
		sdk.ValAddress(acc.Address), acc.Priv.PubKey(), bondCoin, description3, commissionMsg,
	)
	txs = mock.GenSimTxs(t, mApp, []sdk.Msg{createValidatorMsg3}, false, acc.Priv)

	// create validators
	var msgs []sdk.Msg
	var privs []crypto.PrivKey
	for i := 0; i < 10; i++ {
		newBondCoin := sdk.NewCoin("BNB", sdk.NewDecWithoutFra(20000+int64(i)).RawInt())
		description := NewDescription(fmt.Sprintf("account%d", i), "", "", "")
		createValidatorMsg := NewMsgCreateValidator(
			sdk.ValAddress(accounts[i].Address), accounts[i].Priv.PubKey(), newBondCoin, description, commissionMsg,
		)
		msgs = append(msgs, createValidatorMsg)
		privs = append(privs, accounts[i].Priv)
	}
	txs = mock.GenSimTxs(t, mApp, msgs, true, privs...)
	height = mock.ApplyBlock(t, mApp.BaseApp, height, txs)
	ctx = mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators = keeper.GetLastValidators(ctx)
	require.Len(t, validators, 2)

	// new validators elected
	height = mock.ApplyEmptyBlocks(t, mApp.BaseApp, height, 5)
	ctx = mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators = keeper.GetLastValidators(ctx)
	mApp.Logger.Debug("new validators elected", "validators", validators)
	require.Len(t, validators, 11)

	//// addr1 create validator on behalf of addr2
	//createValidatorMsgOnBehalfOf := NewMsgCreateValidatorOnBehalfOf(
	//	addr1, sdk.ValAddress(addr2), priv2.PubKey(), bondCoin, description, commissionMsg,
	//)
	//
	//mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{createValidatorMsgOnBehalfOf}, []int64{0, 0}, []int64{1, 0}, true, true, priv1, priv2)
	//mock.GenSimTxs(t, mApp, []sdk.Msg{createValidatorMsgOnBehalfOf}, true, []{priv1})
	//mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin.Minus(bondCoin).Minus(bondCoin)})

	//validator = checkValidator(t, mApp, keeper, sdk.ValAddress(addr2), true)
	//require.Equal(t, sdk.ValAddress(addr2), validator.OperatorAddr)
	//require.Equal(t, sdk.Bonded, validator.Status)
	//require.True(sdk.DecEq(t, sdk.NewDecWithoutFra(10), validator.Tokens))
	//
	//// check the bond that should have been created as well
	//checkDelegation(t, mApp, keeper, addr1, sdk.ValAddress(addr1), true, sdk.NewDecWithoutFra(10))
	//
	//// edit the validator
	//description = NewDescription("bar_moniker", "", "", "")
	//editValidatorMsg := NewMsgEditValidator(sdk.ValAddress(addr1), description, nil)
	//
	//mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{editValidatorMsg}, []int64{0}, []int64{2}, true, true, priv1)
	//validator = checkValidator(t, mApp, keeper, sdk.ValAddress(addr1), true)
	//require.Equal(t, description, validator.Description)
	//
	//// delegate
	//mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})
	//delegateMsg := NewMsgDelegate(addr2, sdk.ValAddress(addr1), bondCoin)
	//
	//mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{delegateMsg}, []int64{0}, []int64{1}, true, true, priv2)
	//mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Minus(bondCoin)})
	//checkDelegation(t, mApp, keeper, addr2, sdk.ValAddress(addr1), true, sdk.NewDecWithoutFra(10))
	//
	//// begin unbonding
	//beginUnbondingMsg := NewMsgUndelegate(addr2, sdk.ValAddress(addr1), sdk.NewDecWithoutFra(10))
	//mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{beginUnbondingMsg}, []int64{0}, []int64{2}, true, true, priv2)
	//
	//// delegation should exist anymore
	//checkDelegation(t, mApp, keeper, addr2, sdk.ValAddress(addr1), false, sdk.Dec{})
	//
	//// balance should be the same because bonding not yet complete
	//mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Minus(bondCoin)})
}
