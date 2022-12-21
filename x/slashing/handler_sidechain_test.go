package slashing

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/bsc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/fees"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/stretchr/testify/require"
)

func TestSideChainSlashDoubleSign(t *testing.T) {
	slashParams := DefaultParams()
	slashParams.DoubleSignUnbondDuration = 5 * time.Second
	slashParams.MaxEvidenceAge = math.MaxInt64
	slashParams.DoubleSignSlashAmount = 6000e8
	slashParams.SubmitterReward = 3000e8
	submitter := sdk.AccAddress(addrs[2])
	ctx, sideCtx, bankKeeper, stakeKeeper, _, keeper := createSideTestInput(t, slashParams)

	// create a malicious validator
	ctx = ctx.WithBlockHeight(100)
	bondAmount := int64(10000e8)
	mValAddr := addrs[0]
	mSideConsAddr, err := sdk.HexDecode("0xed24ff64903c07B5bD57C898CE0967D407aFCB0d")
	require.Nil(t, err)
	mSideFeeAddr := createSideAddr(20)
	msgCreateVal := newTestMsgCreateSideValidator(mValAddr, mSideConsAddr, mSideFeeAddr, bondAmount)
	got := stake.NewHandler(stakeKeeper, gov.Keeper{})(ctx, msgCreateVal)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got: %v", got)
	// end block
	stake.EndBreatheBlock(ctx, stakeKeeper)

	ctx = ctx.WithBlockHeight(200)
	ValAddr1 := addrs[1]
	sideConsAddr1, sideFeeAddr1 := createSideAddr(20), createSideAddr(20)
	msgCreateVal1 := newTestMsgCreateSideValidator(ValAddr1, sideConsAddr1, sideFeeAddr1, bondAmount)
	got1 := stake.NewHandler(stakeKeeper, gov.Keeper{})(ctx, msgCreateVal1)
	require.True(t, got1.IsOK(), "expected create validator msg to be ok, got: %v", got1)
	// end block
	stake.EndBreatheBlock(ctx, stakeKeeper)
	stakingPoolBalance := bankKeeper.GetCoins(ctx, stake.DelegationAccAddr).AmountOf("steak")
	require.EqualValues(t, bondAmount*2, stakingPoolBalance)

	ctx = ctx.WithBlockHeight(300)
	headers := make([]bsc.Header, 0)
	headersJson := `[{"parentHash":"0x6116de25352c93149542e950162c7305f207bbc17b0eb725136b78c80aed79cc","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0xe7cb9d2fd449f7bd11126bff55266e7b74936f2f230e21d44d75c04b7780dfeb","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x1","gasLimit":"0x47e7c4","gasUsed":"0x0","timestamp":"0x5ea6a002","extraData":"0x0000000000000000000000000000000000000000000000000000000000000000fc3e4bbcd4936a8e1fd9fc45461d071ca571ca80fbed85e0cc52e007ed557aff0a6ea1875b4e13171d301037036b3a26af3c7c2b317487323fd7557df717856b00","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x1532065752393ff2f6e7ef9b64f80d6e10efe42a4d9bdd8149fcbac6f86b365b"},{"parentHash":"0x6116de25352c93149542e950162c7305f207bbc17b0eb725136b78c80aed79cc","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0xe7cb9d2fd449f7bd11126bff55266e7b74936f2f230e21d44d75c04b7780dfeb","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x1","gasLimit":"0x47e7c4","gasUsed":"0x64","timestamp":"0x5ea6a002","extraData":"0x00000000000000000000000000000000000000000000000000000000000000003a849df14e9cc1502f218431c449f239a51fddb1fd408ca37e61834adf921f0c21fd269c86acf7f0b40aa7ce691bbd7f446d8234a4a6b19a98c77614da9a5fcb01","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x811a42453f826f05e9d85998551636f59eb740d5b03fe2416700058a4f31ca1e"}]`
	err = json.Unmarshal([]byte(headersJson), &headers)
	require.Nil(t, err)

	feesInPoolBefore := fees.Pool.BlockFees().Tokens.AmountOf("steak")
	msgSubmitEvidence := NewMsgBscSubmitEvidence(submitter, headers)

	sdk.UpgradeMgr.AddUpgradeHeight(sdk.FixDoubleSignChainId, 199)
	sdk.UpgradeMgr.SetHeight(200)
	got = NewHandler(keeper)(ctx, msgSubmitEvidence)
	require.True(t, got.IsOK(), "expected submit evidence msg to be ok, got: %v", got)

	mValidator, found := stakeKeeper.GetValidator(sideCtx, mValAddr)
	require.True(t, found)
	require.True(t, mValidator.Jailed)
	require.EqualValues(t, bondAmount-slashParams.DoubleSignSlashAmount, mValidator.Tokens.RawInt())
	require.EqualValues(t, bondAmount-slashParams.DoubleSignSlashAmount, mValidator.DelegatorShares.RawInt())

	submitterBalance := bankKeeper.GetCoins(ctx, submitter).AmountOf("steak")
	require.EqualValues(t, initCoins+slashParams.SubmitterReward, submitterBalance)

	require.EqualValues(t, slashParams.DoubleSignSlashAmount-slashParams.SubmitterReward, fees.Pool.BlockFees().Tokens.AmountOf("steak")-feesInPoolBefore)

	slashRecord, found := keeper.getSlashRecord(sideCtx, mSideConsAddr, DoubleSign, 1)
	require.True(t, found)
	require.EqualValues(t, slashParams.DoubleSignSlashAmount, slashRecord.SlashAmt)
	require.EqualValues(t, ctx.BlockHeader().Time.Add(slashParams.DoubleSignUnbondDuration).Unix(), slashRecord.JailUntil.Unix())

	expectedStakingPoolBalance := stakingPoolBalance - slashParams.DoubleSignSlashAmount
	stakingPoolBalance = bankKeeper.GetCoins(ctx, stake.DelegationAccAddr).AmountOf("steak")
	require.EqualValues(t, expectedStakingPoolBalance, stakingPoolBalance)
	// end block
	stake.EndBreatheBlock(ctx, stakeKeeper)

	realSlashedAmt := sdk.MinInt64(slashParams.DoubleSignSlashAmount, mValidator.Tokens.RawInt())
	realSubmitterReward := sdk.MinInt64(slashParams.SubmitterReward, mValidator.Tokens.RawInt())
	expectedAfterValTokensLeft := mValidator.Tokens.RawInt() - realSlashedAmt
	expectedAfterSubmitterBalance := submitterBalance + realSubmitterReward
	// send submit evidence tx
	ctx = ctx.WithBlockHeight(350).WithBlockTime(time.Now())
	headersJson = `[{"parentHash":"0x9dc70cfc956472119b82b6bbc1e6be139a68d03e99a4dcec1ccd0d9b4fd9c822","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0x0988fe1673073b5e1c5f052e5a9a30ec871f90768041a7bfed5ee03f6304b138","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x2","gasLimit":"0x47e7c4","gasUsed":"0x0","timestamp":"0x5eb8fc64","extraData":"0x00000000000000000000000000000000000000000000000000000000000000001d5b50270c673b96065304de53acb9617ef235be1ab6a7c16d7a660c2b13a8c22f513f4f8f43f427d64cbe57e23cd73f86e1efc79cabc1a89392a9e1c267f57d00","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x132a6caa72f3e5e98b086c5bcf2d7fe95ac612152114caca3e95bc8ec8e068a0"},{"parentHash":"0x9dc70cfc956472119b82b6bbc1e6be139a68d03e99a4dcec1ccd0d9b4fd9c822","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0x0988fe1673073b5e1c5f052e5a9a30ec871f90768041a7bfed5ee03f6304b138","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x2","gasLimit":"0x47e7c4","gasUsed":"0x64","timestamp":"0x5eb8fc64","extraData":"0x000000000000000000000000000000000000000000000000000000000000000065f3ddb3c4d6a42f220ead9cb60eebaef4f31f198dbfc18f3b226b84dde4f9232dca010aff3c8a647fe9e51a3f59b6e46fee461d456e69fafd33362faec0813601","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x9c4f11247e697a75ba633f87112895d537156265bf52b6e85e43e551b4d1cb78"}]`
	err = json.Unmarshal([]byte(headersJson), &headers)
	require.Nil(t, err)
	msgSubmitEvidence = NewMsgBscSubmitEvidence(submitter, headers)
	got = NewHandler(keeper)(ctx, msgSubmitEvidence)
	require.True(t, got.IsOK(), "expected submit evidence msg to be ok, got: %v", got)

	// check balance
	expectedStakingPoolBalance = stakingPoolBalance - realSlashedAmt
	stakingPoolBalance = bankKeeper.GetCoins(ctx, stake.DelegationAccAddr).AmountOf("steak")
	require.EqualValues(t, expectedStakingPoolBalance, stakingPoolBalance)

	mValidator, found = stakeKeeper.GetValidator(sideCtx, mValAddr)
	require.True(t, found)
	require.True(t, mValidator.Jailed)
	require.EqualValues(t, expectedAfterValTokensLeft, mValidator.Tokens.RawInt())
	require.EqualValues(t, expectedAfterValTokensLeft, mValidator.DelegatorShares.RawInt())

	submitterBalance = bankKeeper.GetCoins(ctx, submitter).AmountOf("steak")
	require.EqualValues(t, expectedAfterSubmitterBalance, submitterBalance)

	validator1, found := stakeKeeper.GetValidator(sideCtx, ValAddr1)
	require.True(t, found)
	distributionAddr1 := validator1.DistributionAddr
	distributionAddr1Balance := bankKeeper.GetCoins(ctx, distributionAddr1).AmountOf("steak")
	require.EqualValues(t, realSlashedAmt-realSubmitterReward, distributionAddr1Balance)

	slashRecord, found = keeper.getSlashRecord(sideCtx, mSideConsAddr, DoubleSign, 2)
	require.True(t, found)
	require.EqualValues(t, realSlashedAmt, slashRecord.SlashAmt)
	require.EqualValues(t, ctx.BlockHeader().Time.Add(slashParams.DoubleSignUnbondDuration).Unix(), slashRecord.JailUntil.Unix())
}

func TestSideChainSlashDoubleSignUBD(t *testing.T) {

	slashParams := DefaultParams()
	slashParams.MaxEvidenceAge = math.MaxInt64
	slashParams.DoubleSignSlashAmount = 6000e8
	slashParams.SubmitterReward = 3000e8
	submitter := sdk.AccAddress(addrs[2])
	ctx, sideCtx, bankKeeper, stakeKeeper, _, keeper := createSideTestInput(t, slashParams)

	// create a malicious validator
	ctx = ctx.WithBlockHeight(100)
	bondAmount := int64(10000e8)
	mValAddr := addrs[0]
	mSideConsAddr, err := sdk.HexDecode("0xed24ff64903c07B5bD57C898CE0967D407aFCB0d")
	require.Nil(t, err)
	mSideFeeAddr := createSideAddr(20)
	msgCreateVal := newTestMsgCreateSideValidator(mValAddr, mSideConsAddr, mSideFeeAddr, bondAmount)
	got := stake.NewHandler(stakeKeeper, gov.Keeper{})(ctx, msgCreateVal)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got: %v", got)
	// end block
	stake.EndBreatheBlock(ctx, stakeKeeper)

	ctx = ctx.WithBlockHeight(150)
	msgUnDelegate := newTestMsgSideUnDelegate(sdk.AccAddress(mValAddr), mValAddr, 5000e8)
	got = stake.NewHandler(stakeKeeper, gov.Keeper{})(ctx, msgUnDelegate)
	require.True(t, got.IsOK(), "expected unDelegate msg to be ok, got: %v", got)
	ubd, found := stakeKeeper.GetUnbondingDelegation(sideCtx, sdk.AccAddress(mValAddr), mValAddr)
	require.True(t, found)
	require.EqualValues(t, 5000e8, ubd.Balance.Amount)

	ctx = ctx.WithBlockHeight(200)
	// end block
	stake.EndBreatheBlock(ctx, stakeKeeper)

	ctx = ctx.WithBlockHeight(201)
	headers := make([]bsc.Header, 0)
	headersJson := `[{"parentHash":"0x6116de25352c93149542e950162c7305f207bbc17b0eb725136b78c80aed79cc","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0xe7cb9d2fd449f7bd11126bff55266e7b74936f2f230e21d44d75c04b7780dfeb","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x1","gasLimit":"0x47e7c4","gasUsed":"0x0","timestamp":"0x5ea6a002","extraData":"0x0000000000000000000000000000000000000000000000000000000000000000fc3e4bbcd4936a8e1fd9fc45461d071ca571ca80fbed85e0cc52e007ed557aff0a6ea1875b4e13171d301037036b3a26af3c7c2b317487323fd7557df717856b00","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x1532065752393ff2f6e7ef9b64f80d6e10efe42a4d9bdd8149fcbac6f86b365b"},{"parentHash":"0x6116de25352c93149542e950162c7305f207bbc17b0eb725136b78c80aed79cc","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0xe7cb9d2fd449f7bd11126bff55266e7b74936f2f230e21d44d75c04b7780dfeb","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x1","gasLimit":"0x47e7c4","gasUsed":"0x64","timestamp":"0x5ea6a002","extraData":"0x00000000000000000000000000000000000000000000000000000000000000003a849df14e9cc1502f218431c449f239a51fddb1fd408ca37e61834adf921f0c21fd269c86acf7f0b40aa7ce691bbd7f446d8234a4a6b19a98c77614da9a5fcb01","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x811a42453f826f05e9d85998551636f59eb740d5b03fe2416700058a4f31ca1e"}]`
	err = json.Unmarshal([]byte(headersJson), &headers)
	require.Nil(t, err)

	feesInPoolBefore := fees.Pool.BlockFees().Tokens.AmountOf("steak")

	sdk.UpgradeMgr.AddUpgradeHeight(sdk.FixDoubleSignChainId, 199)
	sdk.UpgradeMgr.SetHeight(200)
	msgSubmitEvidence := NewMsgBscSubmitEvidence(submitter, headers)
	got = NewHandler(keeper)(ctx, msgSubmitEvidence)
	require.True(t, got.IsOK(), "expected submit evidence msg to be ok, got: %v", got)

	mValidator, found := stakeKeeper.GetValidator(sideCtx, mValAddr)
	require.True(t, found)
	require.True(t, mValidator.Jailed)
	require.EqualValues(t, 0, mValidator.Tokens.RawInt())
	require.EqualValues(t, 0, mValidator.DelegatorShares.RawInt())

	ubd, found = stakeKeeper.GetUnbondingDelegation(sideCtx, sdk.AccAddress(mValAddr), mValAddr)
	require.True(t, found)
	require.EqualValues(t, 4000e8, ubd.Balance.Amount)

	submitterBalance := bankKeeper.GetCoins(ctx, submitter).AmountOf("steak")
	require.EqualValues(t, initCoins+slashParams.SubmitterReward, submitterBalance)

	require.EqualValues(t, slashParams.DoubleSignSlashAmount-slashParams.SubmitterReward, fees.Pool.BlockFees().Tokens.AmountOf("steak")-feesInPoolBefore)

	stakingPoolBalance := bankKeeper.GetCoins(ctx, stake.DelegationAccAddr).AmountOf("steak")
	require.EqualValues(t, 4000e8, stakingPoolBalance)

}
