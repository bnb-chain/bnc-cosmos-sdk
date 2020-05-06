package slashing

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/fees"
)

const (
	ClaimTypeDowntimeSlash sdk.ClaimType = 0x5

	ClaimNameDowntimeSlash = "DowntimeSlash"
)

type ClaimHooks struct {
	k Keeper
}

// Return the wrapper struct
func (k Keeper) ClaimHooks() ClaimHooks {
	return ClaimHooks{k}
}

var _ sdk.ClaimHooks = ClaimHooks{}

type SideDowntimeSlashClaim struct {
	SideConsAddr  []byte `json:"side_cons_addr"`
	SideHeight    int64  `json:"side_height"`
	SideChainId   string `json:"side_chain_id"`
	SideTimestamp int64  `json:"side_timestamp"`
}

// implement Claim hooks
func (h ClaimHooks) CheckClaim(ctx sdk.Context, claim string) sdk.Error {
	var slashClaim SideDowntimeSlashClaim
	err := json.Unmarshal([]byte(claim), &slashClaim)
	if err != nil {
		return ErrInvalidClaim(h.k.Codespace, fmt.Sprintf("unmarshal side downtime slash claim error, claim=%s", claim))
	}

	if len(slashClaim.SideConsAddr) != sdk.AddrLen {
		return ErrInvalidClaim(h.k.Codespace, fmt.Sprintf("wrong sideConsAddr length, expected=%d", slashClaim.SideConsAddr))
	}

	if slashClaim.SideHeight <= 0 {
		return ErrInvalidClaim(h.k.Codespace, "side height must be positive")
	}

	if slashClaim.SideTimestamp <= 0 {
		return ErrInvalidClaim(h.k.Codespace, "invalid side timestamp")
	}
	return nil
}

func (h ClaimHooks) ExecuteClaim(ctx sdk.Context, finalClaim string) (sdk.Tags, sdk.Error) {
	var slashClaim SideDowntimeSlashClaim
	err := json.Unmarshal([]byte(finalClaim), &slashClaim)
	if err != nil {
		return sdk.EmptyTags(), ErrInvalidClaim(h.k.Codespace, fmt.Sprintf("unmarshal side downtime slash claim error, claim=%s", finalClaim))
	}

	sideCtx, err := h.k.ScKeeper.PrepareCtxForSideChain(ctx, slashClaim.SideChainId)
	if err != nil {
		return sdk.EmptyTags(), ErrInvalidSideChainId(DefaultCodespace)
	}

	header := sideCtx.BlockHeader()
	age := header.Time.Unix() - slashClaim.SideTimestamp
	if age > int64(h.k.MaxEvidenceAge(sideCtx).Seconds()) {
		return sdk.EmptyTags(), ErrExpiredEvidence(h.k.Codespace)
	}

	if h.k.hasSlashRecord(sideCtx, slashClaim.SideConsAddr, Downtime, slashClaim.SideHeight) {
		return sdk.EmptyTags(), ErrDuplicateDowntimeClaim(h.k.Codespace)
	}

	slashAmt := h.k.DowntimeSlashAmount(sideCtx)
	slashedAmt, err := h.k.validatorSet.SlashSideChain(ctx, slashClaim.SideChainId, slashClaim.SideConsAddr, sdk.NewDec(slashAmt))
	if err != nil {
		return sdk.EmptyTags(), ErrFailedToSlash(h.k.Codespace, err.Error())
	}

	downtimeClaimFee := h.k.DowntimeSlashFee(sideCtx)
	downtimeClaimFeeReal := sdk.MinInt64(downtimeClaimFee, slashedAmt.RawInt())
	bondDenom := h.k.validatorSet.BondDenom(sideCtx)
	if downtimeClaimFeeReal > 0 && ctx.IsDeliverTx() {
		feeCoinAdd := sdk.NewCoin(bondDenom, downtimeClaimFeeReal)
		fees.Pool.AddAndCommitFee("side_downtime_slash", sdk.NewFee(sdk.Coins{feeCoinAdd}, sdk.FeeForAll))
	}

	remaining := slashedAmt.RawInt() - downtimeClaimFeeReal
	if remaining > 0 {
		found, err := h.k.validatorSet.AllocateSlashAmtToValidators(sideCtx, slashClaim.SideConsAddr, sdk.NewDec(remaining))
		if err != nil {
			return sdk.EmptyTags(), ErrFailedToSlash(h.k.Codespace, err.Error())
		}
		if !found && ctx.IsDeliverTx() {
			remainingCoin := sdk.NewCoin(bondDenom, remaining)
			fees.Pool.AddAndCommitFee("side_downtime_slash_remaining", sdk.NewFee(sdk.Coins{remainingCoin}, sdk.FeeForAll))
		}
	}

	jailUtil := header.Time.Add(h.k.DowntimeUnbondDuration(sideCtx))
	sr := SlashRecord{
		ConsAddr:         slashClaim.SideConsAddr,
		InfractionType:   Downtime,
		InfractionHeight: slashClaim.SideHeight,
		SlashHeight:      header.Height,
		JailUntil:        jailUtil,
		SlashAmt:         slashedAmt.RawInt(),
		SideChainId:      slashClaim.SideChainId,
	}
	h.k.setSlashRecord(sideCtx, sr)

	// Set or updated validator jail duration
	signInfo, found := h.k.getValidatorSigningInfo(sideCtx, slashClaim.SideConsAddr)
	if !found {
		return sdk.EmptyTags(), sdk.ErrInternal(fmt.Sprintf("Expected signing info for validator %s but not found", sdk.HexEncode(slashClaim.SideConsAddr)))
	}
	signInfo.JailedUntil = jailUtil
	h.k.setValidatorSigningInfo(sideCtx, slashClaim.SideConsAddr, signInfo)

	return sdk.EmptyTags(), nil
}
