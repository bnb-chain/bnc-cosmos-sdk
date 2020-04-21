package slashing

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ClaimTypeDowntimeSlash sdk.ClaimType = 0x5

	ClaimNameDowntimeSlash = "DowntimeSlash"
)

var _ sdk.ClaimHooks = Hooks{}

type SideDowntimeSlashClaim struct {
	SideConsAddr []byte
	SideHeight   int64
	SideChainId  string
}

// implement Claim hooks
func (h Hooks) CheckClaim(ctx sdk.Context, claim string) sdk.Error {
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
	return nil
}

func (h Hooks) ExecuteClaim(ctx sdk.Context, finalClaim string) (sdk.Tags, sdk.Error) {
	var slashClaim SideDowntimeSlashClaim
	err := json.Unmarshal([]byte(finalClaim), &slashClaim)
	if err != nil {
		return sdk.EmptyTags(), ErrInvalidClaim(h.k.Codespace, fmt.Sprintf("unmarshal side downtime slash claim error, claim=%s", finalClaim))
	}

	sideCtx, err := h.k.ScKeeper.PrepareCtxForSideChain(ctx, slashClaim.SideChainId)
	if err != nil {
		return sdk.EmptyTags(), ErrInvalidSideChainId(DefaultCodespace)
	}

	if h.k.hasSlashRecord(sideCtx, slashClaim.SideConsAddr, Downtime, slashClaim.SideHeight) {
		return sdk.EmptyTags(), ErrDuplicateDowntimeClaim(h.k.Codespace)
	}

	slashAmt := h.k.DowntimeSlashAmount(sideCtx)
	slashedAmt, err := h.k.validatorSet.SlashSideChain(ctx, slashClaim.SideChainId, slashClaim.SideConsAddr, sdk.NewDec(slashAmt), sdk.ZeroDec(), nil)
	if err != nil {
		return sdk.EmptyTags(), ErrFailedToSlash(h.k.Codespace, err.Error())
	}

	jailUtil := sideCtx.BlockHeader().Time.Add(h.k.DowntimeUnbondDuration(sideCtx))

	sr := SlashRecord{
		ConsAddr:         slashClaim.SideConsAddr,
		InfractionType:   Downtime,
		InfractionHeight: slashClaim.SideHeight,
		SlashHeight:      sideCtx.BlockHeight(),
		JailUntil:        jailUtil,
		SlashAmt:         slashedAmt,
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
