package slashingsidechain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/types"
	"time"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgSubmitEvidence:
			return handleMsgSubmitEvidence(k,ctx,msg)
		default:
			return sdk.ErrTxDecode("invalid message parse in slashingsidechain module").Result()
		}
	}
}

func handleMsgSubmitEvidence(k keeper.Keeper, ctx sdk.Context, msg types.MsgSubmitEvidence) sdk.Result {
	//verify evidence age
	sideConsAddr,_ := msg.Headers[0].ExtractSignerFromHeader()
	if k.GetSlashRecord(ctx, sideConsAddr.Bytes(), msg.Headers[0].Number.Int64()) != nil {
		return types.ErrHandledEvidence(k.Codespace()).Result()
	}

	evidenceTime := int64(min(msg.Headers[0].Time,msg.Headers[1].Time))
	age := ctx.BlockHeader().Time.Sub(time.Unix(evidenceTime,0))
	if age > k.MaxEvidenceAge(ctx) {
		return types.ErrExpiredEvidence(k.Codespace()).Result()
	}

	slashAmount := k.SlashAmount(ctx)
	submitterReward := k.SubmitterReward(ctx)
	slashErr := k.GetStakeKeeper().SlashForSideChain(ctx,msg.SideChainId, sideConsAddr.Bytes(), sdk.NewDec(slashAmount), sdk.NewDec(submitterReward), msg.Submitter)
	if slashErr != nil {
		return types.ErrFailedToSlash(k.Codespace(), slashErr.Error()).Result()
	}

	k.SetSlashRecord(ctx, sideConsAddr.Bytes(), msg.Headers[0].Number.Int64())
	return sdk.Result{}
}


func min( a,b uint64) uint64{
	if a < b {
		return a
	}
	return b
}


