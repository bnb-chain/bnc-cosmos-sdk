// test BEP159: open staking mechanism
package stake

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/sidechain"
	stake "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
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

func getNewStakeMockApp(t *testing.T) (*mock.App, Keeper, []Account) {
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
	keeper := NewKeeper(mApp.Cdc, keyStake, keyStakeReward, tkeyStake, bankKeeper, nil, paramsKeeper.Subspace(DefaultParamspace), mApp.RegisterCodespace(DefaultCodespace), sdk.ChainID(0), "")
	govKeeper := gov.NewKeeper(mApp.Cdc, keyGov, paramsKeeper, paramsKeeper.Subspace(gov.DefaultParamSpace), bankKeeper, keeper, mApp.RegisterCodespace(DefaultCodespace), nil)
	keeper.SetupForSideChain(&scKeeper, &ibcKeeper)

	sdk.UpgradeMgr.AddUpgradeHeight(sdk.LaunchBscUpgrade, 6)
	sdk.UpgradeMgr.AddUpgradeHeight(sdk.BEP128, 7)
	sdk.UpgradeMgr.AddUpgradeHeight(sdk.BEP159, 8)
	sdk.UpgradeMgr.AddUpgradeHeight(sdk.BEP159Phase2, 8)
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
	sdk.UpgradeMgr.RegisterBeginBlocker(sdk.BEP159, func(ctx sdk.Context) {
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

	// set init accounts
	accs := []sdk.Account{}
	accounts := GenAccounts(100)
	for _, acc := range accounts {
		accs = append(accs, acc.BaseAccount)
	}
	mock.SetGenesis(mApp, accs)
	// create validator
	description := NewDescription("foo_moniker", "", "", "")
	bondCoin := sdk.NewCoin("BNB", sdk.NewDecWithoutFra(10).RawInt())
	createValidatorMsg0 := NewMsgCreateValidator(
		sdk.ValAddress(accounts[0].Address), accounts[0].Priv.PubKey(), bondCoin, description, commissionMsg,
	)
	createValidatorProposal0 := MsgCreateValidatorProposal{
		MsgCreateValidator: createValidatorMsg0,
		ProposalId:         0,
	}
	mApp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Height: 0,
	}})
	tx := mock.GenTx([]sdk.Msg{createValidatorProposal0}, []int64{0}, []int64{0}, accounts[0].Priv)
	res := mApp.Deliver(tx)
	require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	mApp.Commit()

	return mApp, keeper, accounts
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

func setupTest() {
	sdk.UpgradeMgr.Reset()
}

func TestNewStake(t *testing.T) {
	setupTest()
	mApp, keeper, accounts := getNewStakeMockApp(t)

	var height int64 = 1

	// hardfork
	height = mock.ApplyEmptyBlocks(t, mApp.BaseApp, height, 200)
	ctx := mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators := keeper.GetLastValidators(ctx)
	fmt.Printf("%+v\n", validators)

	// fail to create validator after hardfork, self delegation not enough
	acc := accounts[0]
	description3 := NewDescription("validator3", "", "", "")
	bondCoin := sdk.NewCoin("BNB", sdk.NewDecWithoutFra(10).RawInt())
	createValidatorMsg3 := NewMsgCreateValidator(
		sdk.ValAddress(acc.Address), acc.Priv.PubKey(), bondCoin, description3, commissionMsg,
	)
	txs := mock.GenSimTxs(t, mApp, []sdk.Msg{createValidatorMsg3}, false, acc.Priv)

	// create validators
	var msgs []sdk.Msg
	var privs []crypto.PrivKey
	for i := 2; i < 12; i++ {
		newBondCoin := sdk.NewCoin("BNB", sdk.NewDecWithoutFra(20000+int64(i)).RawInt())
		description := NewDescription(fmt.Sprintf("account%d", i), "", "", "")
		createValidatorOpenMsg := MsgCreateValidatorOpen{
			DelegatorAddr: accounts[i].Address,
			ValidatorAddr: sdk.ValAddress(accounts[i].Address),
			Delegation:    newBondCoin,
			Description:   description,
			Commission:    commissionMsg,
			PubKey:        sdk.MustBech32ifyConsPub(accounts[i].Priv.PubKey()),
		}
		msgs = append(msgs, createValidatorOpenMsg)
		privs = append(privs, accounts[i].Priv)
	}
	txs = mock.GenSimTxs(t, mApp, msgs, true, privs...)
	height = mock.ApplyBlock(t, mApp.BaseApp, height, txs)
	ctx = mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators = keeper.GetLastValidators(ctx)
	require.Len(t, validators, 1)

	// new validators elected
	height = mock.ApplyEmptyBlocks(t, mApp.BaseApp, height, 5)
	ctx = mApp.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{Height: height})
	validators = keeper.GetLastValidators(ctx)
	mApp.Logger.Debug("new validators elected", "validators", validators)
	require.Len(t, validators, 11)
}
