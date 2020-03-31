package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/types"
	"time"
)

type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         *codec.Codec
	paramstore  params.Subspace
	stakeKeeper StakeKeeper

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, stakeKeeper StakeKeeper, paramstore params.Subspace, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		paramstore:  paramstore.WithTypeTable(types.ParamTypeTable()),
		codespace:   codespace,
		stakeKeeper: stakeKeeper,
	}
	return keeper
}

func (k Keeper) GetStakeKeeper() StakeKeeper {
	return k.stakeKeeper
}

type StakeKeeper interface {
	SlashForSideChain(ctx sdk.Context, sideChainId string, cosAddr sdk.ConsAddress, slashAmount sdk.Dec, submitterReward sdk.Dec, submitter sdk.AccAddress) error
}

// return the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}


func (k Keeper) SlashAmount(ctx sdk.Context) (slashAmt int64) {
	k.paramstore.Get(ctx, types.KeySlashAmount, &slashAmt)
	return
}

func (k Keeper) SubmitterReward(ctx sdk.Context) (submitterReward int64) {
	k.paramstore.Get(ctx, types.KeySubmitterReward, &submitterReward)
	return
}

func (k Keeper) MaxEvidenceAge(ctx sdk.Context) (maxEvidenceAge time.Duration) {
	k.paramstore.Get(ctx, types.KeyMaxEvidenceAge, &maxEvidenceAge)
	return
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
