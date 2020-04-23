package slashing

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func handleMsgBscSubmitEvidence(ctx sdk.Context, msg MsgBscSubmitEvidence, k Keeper) sdk.Result {
	sideChainId := k.BscSideChainId(ctx)
	if scCtx, err := k.scKeeper.PrepareCtxForSideChain(ctx, sideChainId); err != nil {
		return ErrInvalidSideChainId(DefaultCodespace).Result()
	} else {
		ctx = scCtx
	}

	sideConsAddr, err := msg.Headers[0].ExtractSignerFromHeader()
	if err != nil {
		return ErrInvalidEvidence(DefaultCodespace, fmt.Sprintf("Failed to extract signer from block header, %s", err.Error())).Result()
	}
	sideConsAddr2, err := msg.Headers[1].ExtractSignerFromHeader()
	if err != nil {
		return ErrInvalidEvidence(DefaultCodespace, fmt.Sprintf("Failed to extract signer from block header, %s", err.Error())).Result()
	}
	if bytes.Compare(sideConsAddr.Bytes(), sideConsAddr2.Bytes()) != 0 {
		return ErrInvalidEvidence(DefaultCodespace, "The signers of two block headers are not the same").Result()
	}

	//verify evidence age
	if k.getSlashRecord(ctx, sideConsAddr.Bytes(), msg.Headers[0].Number) != nil {
		return ErrEvidenceHasBeenHandled(k.codespace).Result()
	}

	evidenceTime := int64(sdk.Min(msg.Headers[0].Time, msg.Headers[1].Time))
	age := ctx.BlockHeader().Time.Sub(time.Unix(evidenceTime, 0))
	if age > k.MaxEvidenceAge(ctx) {
		return ErrExpiredEvidence(k.codespace).Result()
	}

	slashAmount := k.SlashAmount(ctx)
	submitterReward := k.SubmitterReward(ctx)
	slashErr := k.validatorSet.SlashSideChain(ctx, sideChainId, sideConsAddr.Bytes(), sdk.NewDec(slashAmount), sdk.NewDec(submitterReward), msg.Submitter)
	if slashErr != nil {
		return ErrFailedToSlash(k.codespace, slashErr.Error()).Result()
	}

	k.setSlashRecord(ctx, sideConsAddr.Bytes(), msg.Headers[0].Number)

	// Set or updated validator jail duration
	signInfo, found := k.getValidatorSigningInfo(ctx, sideConsAddr.Bytes())
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", sideConsAddr.Hex()))
	}
	signInfo.JailedUntil = ctx.BlockHeader().Time.Add(k.DoubleSignUnbondDuration(ctx))
	k.setValidatorSigningInfo(ctx, sideConsAddr.Bytes(), signInfo)

	return sdk.Result{}
}

func handleMsgSideChainUnjail(ctx sdk.Context, msg MsgSideChainUnjail, k Keeper) sdk.Result {
	if scCtx, err := k.scKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId); err != nil {
		return ErrInvalidSideChainId(DefaultCodespace).Result()
	} else {
		ctx = scCtx
	}

	if err := k.Unjail(ctx, msg.ValidatorAddr); err != nil {
		return err.Result()
	}

	tags := sdk.NewTags("action", []byte("unjail"), "validator", []byte(msg.ValidatorAddr.String()))

	return sdk.Result{
		Tags: tags,
	}
}
