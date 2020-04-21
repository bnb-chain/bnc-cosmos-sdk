package slashing

import (
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing/bsc"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
)

func TestSideChainSlashDoubleSign(t *testing.T) {
	// initial setup
	slashParams := DefaultParams()
	slashParams.DoubleSignUnbondDuration = 5 * time.Second
	slashParams.DoubleSignSlashAmount = 15000e8
	ctx, sideCtx, bankKeeper, stakeKeeper, _, keeper := createSideTestInput(t, slashParams)

	// create a validator
	bondAmount := int64(10000e8)
	realSlashAmt := sdk.MinInt64(slashParams.DoubleSignSlashAmount,bondAmount)
	realSubmitterReward := sdk.MinInt64(slashParams.SubmitterReward,bondAmount)

	valAddr1 := addrs[0]

	sideConsAddr1, err := sdk.HexDecode("0x625448c3f21AB4636bBCef84Baaf8D6cCdE13c3F")
	require.Nil(t, err)

	sideFeeAddr1 := createSideAddr(20)
	msgCreateVal1 := newTestMsgCreateSideValidator(valAddr1, sideConsAddr1, sideFeeAddr1, bondAmount)
	got1 := stake.NewHandler(stakeKeeper, gov.Keeper{})(ctx, msgCreateVal1)
	require.True(t, got1.IsOK(), "expected create validator msg to be ok, got: %v", got1)

	valAddr2 := addrs[1]
	sideConsAddr2, sideFeeAddr2 := createSideAddr(20), createSideAddr(20)
	msgCreateVal2 := newTestMsgCreateSideValidator(valAddr2, sideConsAddr2, sideFeeAddr2, bondAmount)
	got2 := stake.NewHandler(stakeKeeper, gov.Keeper{})(ctx, msgCreateVal2)
	require.True(t, got2.IsOK(), "expected create validator msg to be ok, got: %v", got2)

	// end block
	stake.EndBreatheBlock(ctx, stakeKeeper)

	submitter := sdk.AccAddress(addrs[2])
	headers := make([]bsc.Header, 0)
	headersJson := `[{"parentHash":"0x6116de25352c93149542e950162c7305f207bbc17b0eb725136b78c80aed79cc","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0xe7cb9d2fd449f7bd11126bff55266e7b74936f2f230e21d44d75c04b7780dfeb","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x1","gasLimit":"0x47e7c4","gasUsed":"0x0","timestamp":"0x5ea6a002","extraData":"0x0000000000000000000000000000000000000000000000000000000000000000bb4a77b57c2a82de97b557442883ee19d481a415fc76d3833de83ba37f2d8674375f85fd96affd603244e3448a2b101c40511aa18ce8c1edf4e940dec648ac1300","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x1532065752393ff2f6e7ef9b64f80d6e10efe42a4d9bdd8149fcbac6f86b365b"},{"parentHash":"0x6116de25352c93149542e950162c7305f207bbc17b0eb725136b78c80aed79cc","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","miner":"0x0000000000000000000000000000000000000000","stateRoot":"0xe7cb9d2fd449f7bd11126bff55266e7b74936f2f230e21d44d75c04b7780dfeb","transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","difficulty":"0x20000","number":"0x1","gasLimit":"0x47e7c4","gasUsed":"0x64","timestamp":"0x5ea6a002","extraData":"0x000000000000000000000000000000000000000000000000000000000000000055a9a47820e18c025d0b98a722c3fb83d28e4547e0090cbe5cc17683b7f25d5e18c6e359631ec10d9c08ceaafc9e9847de3de18694d073af9515638eee73c58e00","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x0000000000000000","hash":"0x811a42453f826f05e9d85998551636f59eb740d5b03fe2416700058a4f31ca1e"}]`
	err = json.Unmarshal([]byte(headersJson), &headers)
	require.Nil(t, err)

	msgSubmitEvidence := NewMsgBscSubmitEvidence(submitter, headers)
	got := NewSlashingHandler(keeper)(ctx, msgSubmitEvidence)
	require.True(t, got.IsOK(), "expected submit evidence msg to be ok, got: %v", got)

	// check bad validator state
	validator1, found := stakeKeeper.GetValidatorBySideConsAddr(sideCtx, sideConsAddr1)
	require.True(t, found)
	require.EqualValues(t, 0, validator1.Tokens.RawInt()) // should be 0, as the delegation left is not enough to pay a fine
	require.EqualValues(t, 0, validator1.DelegatorShares.RawInt())

	_, found = stakeKeeper.GetDelegation(sideCtx, validator1.FeeAddr, validator1.OperatorAddr)
	require.False(t, found)

	validator2, found := stakeKeeper.GetValidatorBySideConsAddr(sideCtx, sideConsAddr2)
	require.True(t, found)
	balance := bankKeeper.GetCoins(ctx, validator2.DistributionAddr).AmountOf("steak")
	submitterBalance := bankKeeper.GetCoins(ctx, submitter).AmountOf("steak")
	require.EqualValues(t, initCoins + realSubmitterReward, submitterBalance)
	require.EqualValues(t, realSlashAmt - realSubmitterReward, balance)

	slashRecord, found := keeper.getSlashRecord(sideCtx, sideConsAddr1, DoubleSign, 1)
	require.True(t, found)
	require.EqualValues(t, realSlashAmt, slashRecord.SlashAmt.RawInt())
	require.EqualValues(t, ctx.BlockHeader().Time.Add(slashParams.DoubleSignUnbondDuration).Unix(), slashRecord.JailUntil.Unix())
}
