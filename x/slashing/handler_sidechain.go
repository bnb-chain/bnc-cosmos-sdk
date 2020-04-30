package slashing

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/fees"
)

func handleMsgBscSubmitEvidence(ctx sdk.Context, msg MsgBscSubmitEvidence, k Keeper) sdk.Result {
	sideChainId := k.BscSideChainId(ctx)
	sideCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, sideChainId)
	if err != nil {
		return ErrInvalidSideChainId(DefaultCodespace).Result()
	}

	header := ctx.BlockHeader()
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

	if k.hasSlashRecord(sideCtx, sideConsAddr.Bytes(), DoubleSign, msg.Headers[0].Number) {
		return ErrEvidenceHasBeenHandled(k.Codespace).Result()
	}

	//verify evidence age
	evidenceTime := msg.Headers[0].Time
	if msg.Headers[0].Time < msg.Headers[1].Time {
		evidenceTime = msg.Headers[1].Time
	}
	age := sideCtx.BlockHeader().Time.Sub(time.Unix(int64(evidenceTime), 0))
	if age > k.MaxEvidenceAge(sideCtx) {
		return ErrExpiredEvidence(k.Codespace).Result()
	}

	slashAmount := k.DoubleSignSlashAmount(sideCtx)
	slashedAmount, slashErr := k.validatorSet.SlashSideChain(ctx, sideChainId, sideConsAddr.Bytes(), sdk.NewDec(slashAmount))
	if slashErr != nil {
		return ErrFailedToSlash(k.Codespace, slashErr.Error()).Result()
	}

	bondDenom := k.validatorSet.BondDenom(sideCtx)
	submitterReward := k.SubmitterReward(sideCtx)
	submitterRewardReal := sdk.MinInt64(slashedAmount.RawInt(), submitterReward)
	submitterRewardCoin := sdk.NewCoin(bondDenom, submitterRewardReal)

	if submitterRewardReal > 0 {
		submitterBalance := k.BankKeeper.GetCoins(ctx, msg.Submitter)
		if err := k.BankKeeper.SetCoins(ctx, msg.Submitter, submitterBalance.Plus(sdk.Coins{submitterRewardCoin})); err != nil {
			return ErrFailedToSlash(k.Codespace, err.Error()).Result()
		}
	}

	remainingReward := slashedAmount.RawInt() - submitterRewardReal
	if remainingReward > 0 {
		found, err := k.validatorSet.AllocateSlashAmtToValidators(sideCtx, sideConsAddr.Bytes(), sdk.NewDec(remainingReward))
		if !found { // if the related validators are not found, the amount will be added to fee pool
			remainingCoin := sdk.NewCoin(bondDenom, remainingReward)
			fees.Pool.AddAndCommitFee("side_double_sign_slash", sdk.NewFee(sdk.Coins{remainingCoin}, sdk.FeeForAll))
		}
		if err != nil {
			return ErrFailedToSlash(k.Codespace, err.Error()).Result()
		}
	}

	jailUtil := header.Time.Add(k.DoubleSignUnbondDuration(sideCtx))
	sr := SlashRecord{
		ConsAddr:         sideConsAddr.Bytes(),
		InfractionType:   DoubleSign,
		InfractionHeight: msg.Headers[0].Number,
		SlashHeight:      header.Height,
		JailUntil:        jailUtil,
		SlashAmt:         slashedAmount.RawInt(),
		SideChainId:      sideChainId,
	}
	k.setSlashRecord(sideCtx, sr)

	// Set or updated validator jail duration
	signInfo, found := k.getValidatorSigningInfo(sideCtx, sideConsAddr.Bytes())
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", sideConsAddr.Hex()))
	}
	signInfo.JailedUntil = jailUtil
	k.setValidatorSigningInfo(sideCtx, sideConsAddr.Bytes(), signInfo)

	return sdk.Result{}
}

func handleMsgSideChainUnjail(ctx sdk.Context, msg MsgSideChainUnjail, k Keeper) sdk.Result {

	scCtx, err := k.ScKeeper.PrepareCtxForSideChain(ctx, msg.SideChainId)
	if err != nil {
		return ErrInvalidSideChainId(DefaultCodespace).Result()
	}

	if err := k.Unjail(scCtx, msg.ValidatorAddr); err != nil {
		return err.Result()
	}

	tags := sdk.NewTags("sideChainId", []byte(msg.SideChainId), "validator", []byte(msg.ValidatorAddr.String()))

	return sdk.Result{
		Tags: tags,
	}
}
