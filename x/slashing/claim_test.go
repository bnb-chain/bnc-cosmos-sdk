package slashing

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/types/fees"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
)

func TestSideChainSlashDowntime(t *testing.T) {

	slashingParams := DefaultParams()
	slashingParams.MaxEvidenceAge = 12 * 60 * 60 * time.Second
	ctx, sideCtx, _, stakeKeeper, _, keeper := createSideTestInput(t, slashingParams)
	hooks := keeper.Hooks()

	// create a validator
	bondAmount := int64(10000e8)
	realSlashAmt := sdk.MinInt64(slashingParams.DowntimeSlashAmount, bondAmount)
	realPoolFeeAdd := sdk.MinInt64(realSlashAmt, slashingParams.DowntimeSlashFee)
	valAddr := addrs[0]
	sideConsAddr, sideFeeAddr := createSideAddr(20), createSideAddr(20)
	msgCreateVal := newTestMsgCreateSideValidator(valAddr, sideConsAddr, sideFeeAddr, bondAmount)
	got := stake.NewHandler(stakeKeeper, gov.Keeper{})(ctx, msgCreateVal)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got: %v", got)
	// end block
	stake.EndBreatheBlock(ctx, stakeKeeper)

	sideHeight := int64(100)
	sideChainId := "bsc"
	sideTimestamp := ctx.BlockHeader().Time.Add(-6 * 60 * 60 * time.Second)
	claim := SideDowntimeSlashClaim{
		SideConsAddr:  sideConsAddr,
		SideHeight:    sideHeight,
		SideChainId:   sideChainId,
		SideTimestamp: sideTimestamp.Unix(),
	}

	jsonClaim, err := json.Marshal(claim)
	require.Nil(t, err)

	sdkErr := hooks.CheckClaim(ctx, string(jsonClaim))
	require.Nil(t, sdkErr)

	_, sdkErr = hooks.ExecuteClaim(ctx, string(jsonClaim))
	require.Nil(t, sdkErr, "Expected nil, but got : %v", sdkErr)
	require.EqualValues(t, realPoolFeeAdd, fees.Pool.BlockFees().Tokens.AmountOf("steak"))

	info, found := keeper.getValidatorSigningInfo(sideCtx, sideConsAddr)
	require.True(t, found)
	require.EqualValues(t, ctx.BlockHeader().Time.Add(slashingParams.DowntimeUnbondDuration).Unix(), info.JailedUntil.Unix())

	slashRecord, found := keeper.getSlashRecord(sideCtx, sideConsAddr, Downtime, sideHeight)
	require.True(t, found)
	require.EqualValues(t, sideHeight, slashRecord.InfractionHeight)
	require.EqualValues(t, sideChainId, slashRecord.SideChainId)
	require.EqualValues(t, realSlashAmt, slashRecord.SlashAmt.RawInt())
	require.EqualValues(t, ctx.BlockHeader().Time.Add(slashingParams.DowntimeUnbondDuration).Unix(), slashRecord.JailUntil.Unix())

	validator, found := stakeKeeper.GetValidatorBySideConsAddr(sideCtx, sideConsAddr)
	require.True(t, found)
	require.True(t, validator.Jailed)
	require.EqualValues(t, bondAmount-realSlashAmt, validator.Tokens.RawInt())
	require.EqualValues(t, bondAmount-realSlashAmt, validator.DelegatorShares.RawInt())

	delegation, found := stakeKeeper.GetDelegation(sideCtx, validator.FeeAddr, validator.OperatorAddr)
	require.True(t, found)
	require.EqualValues(t, bondAmount-realSlashAmt, delegation.Shares.RawInt())

	_, sdkErr = hooks.ExecuteClaim(ctx, string(jsonClaim))
	require.NotNil(t, sdkErr)
	require.EqualValues(t, CodeDuplicateDowntimeClaim, sdkErr.Code())

	sdkErr = hooks.CheckClaim(ctx, "")
	require.NotNil(t, sdkErr)

	claim.SideHeight = 0
	jsonClaim, err = json.Marshal(claim)
	require.Nil(t, err)
	sdkErr = hooks.CheckClaim(ctx, string(jsonClaim))
	require.NotNil(t, sdkErr)

	claim.SideHeight = sideHeight
	claim.SideConsAddr = createSideAddr(21)
	jsonClaim, err = json.Marshal(claim)
	require.Nil(t, err)
	sdkErr = hooks.CheckClaim(ctx, string(jsonClaim))
	require.NotNil(t, sdkErr)

	claim.SideConsAddr = sideConsAddr
	claim.SideTimestamp = ctx.BlockHeader().Time.Add(-24 * 60 * 60 * time.Second).Unix()
	jsonClaim, err = json.Marshal(claim)
	require.Nil(t, err)
	sdkErr = hooks.CheckClaim(ctx, string(jsonClaim))
	require.Nil(t, sdkErr)
	_, sdkErr = hooks.ExecuteClaim(ctx, string(jsonClaim))
	require.NotNil(t, sdkErr, "Exepcted get err, but got nil")
	require.EqualValues(t, CodeExpiredEvidence, sdkErr.Code(), "Expected got 201 err code, but got err: %v", sdkErr)

	claim.SideTimestamp = ctx.BlockHeader().Time.Add(-6 * 60 * 60 * time.Second).Unix()
	claim.SideConsAddr = sideConsAddr
	claim.SideChainId = "bcc"
	jsonClaim, err = json.Marshal(claim)
	require.Nil(t, err)
	sdkErr = hooks.CheckClaim(ctx, string(jsonClaim))
	require.Nil(t, sdkErr)
	_, sdkErr = hooks.ExecuteClaim(ctx, string(jsonClaim))
	require.NotNil(t, sdkErr, "Expected get err, but got nil")
	require.EqualValues(t, CodeInvalidSideChain, sdkErr.Code(), "Expected got 205 error code, but got err: %v", sdkErr)

	claim.SideHeight = sideHeight
	claim.SideConsAddr = createSideAddr(20)
	claim.SideChainId = sideChainId
	jsonClaim, err = json.Marshal(claim)
	require.Nil(t, err)
	sdkErr = hooks.CheckClaim(ctx, string(jsonClaim))
	require.Nil(t, sdkErr)
	_, sdkErr = hooks.ExecuteClaim(ctx, string(jsonClaim))
	require.NotNil(t, sdkErr, "Expected got err of no signing info found, but got nil")

}
