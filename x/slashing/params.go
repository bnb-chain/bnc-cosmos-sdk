package slashing

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = "slashing"
)

// Parameter store key
var (
	KeyMaxEvidenceAge           = []byte("MaxEvidenceAge")
	KeySignedBlocksWindow       = []byte("SignedBlocksWindow")
	KeyMinSignedPerWindow       = []byte("MinSignedPerWindow")
	KeyDoubleSignUnbondDuration = []byte("DoubleSignUnbondDuration")
	KeyDowntimeUnbondDuration   = []byte("DowntimeUnbondDuration")
	KeyTooLowDelUnbondDuration  = []byte("TooLowDelUnbondDuration")
	KeySlashFractionDoubleSign  = []byte("SlashFractionDoubleSign")
	KeySlashFractionDowntime    = []byte("SlashFractionDowntime")
	KeyDoubleSignSlashAmount    = []byte("DoubleSignSlashAmount")
	KeyDowntimeSlashAmount      = []byte("DowntimeSlashAmount")
	KeySubmitterReward          = []byte("SubmitterReward")
	KeyBscSideChainId           = []byte("BscSideChainId")
)

// ParamTypeTable for slashing module
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for slashing at genesis
type Params struct {
	MaxEvidenceAge           time.Duration `json:"max_evidence_age"`
	SignedBlocksWindow       int64         `json:"signed_blocks_window"`
	MinSignedPerWindow       sdk.Dec       `json:"min_signed_per_window"`
	DoubleSignUnbondDuration time.Duration `json:"double_sign_unbond_duration"`
	DowntimeUnbondDuration   time.Duration `json:"downtime_unbond_duration"`
	TooLowDelUnbondDuration  time.Duration `json:"too_low_del_unbond_duration"`
	SlashFractionDoubleSign  sdk.Dec       `json:"slash_fraction_double_sign"`
	SlashFractionDowntime    sdk.Dec       `json:"slash_fraction_downtime"`
	DoubleSignSlashAmount    int64         `json:"double_sign_slash_amount"`
	DowntimeSlashAmount      int64         `json:"downtime_slash_amount"`
	SubmitterReward          int64         `json:"submitter_reward"`
	BscSideChainId           string        `json:"bsc_side_chain_id"`
}

// Implements params.ParamStruct
func (p *Params) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{KeyMaxEvidenceAge, &p.MaxEvidenceAge},
		{KeySignedBlocksWindow, &p.SignedBlocksWindow},
		{KeyMinSignedPerWindow, &p.MinSignedPerWindow},
		{KeyDoubleSignUnbondDuration, &p.DoubleSignUnbondDuration},
		{KeyDowntimeUnbondDuration, &p.DowntimeUnbondDuration},
		{KeyTooLowDelUnbondDuration, &p.TooLowDelUnbondDuration},
		{KeySlashFractionDoubleSign, &p.SlashFractionDoubleSign},
		{KeySlashFractionDowntime, &p.SlashFractionDowntime},
		{KeyDoubleSignSlashAmount, &p.DoubleSignSlashAmount},
		{KeyDowntimeSlashAmount, &p.DowntimeSlashAmount},
		{KeySubmitterReward, &p.SubmitterReward},
		{KeyBscSideChainId, &p.BscSideChainId},
	}
}

// Default parameters used by Cosmos Hub
func DefaultParams() Params {
	return Params{
		// defaultMaxEvidenceAge = 60 * 60 * 24 * 7 * 3
		// TODO Temporarily set to 2 minutes for testnets.
		MaxEvidenceAge: 60 * 2 * time.Second,

		// TODO Temporarily set to five minutes for testnets
		DoubleSignUnbondDuration: 60 * 5 * time.Second,

		// TODO Temporarily set to 100 blocks for testnets
		SignedBlocksWindow: 100,

		// TODO Temporarily set to 10 minutes for testnets
		DowntimeUnbondDuration: 60 * 10 * time.Second,

		// TODO Temporarily set to 5 minutes for testnets
		TooLowDelUnbondDuration: 60 * 5 * time.Second,

		MinSignedPerWindow: sdk.NewDecWithPrec(5, 1),

		SlashFractionDoubleSign: sdk.OneDec().Quo(sdk.NewDecWithoutFra(20)),

		SlashFractionDowntime: sdk.OneDec().Quo(sdk.NewDecWithoutFra(100)),

		DoubleSignSlashAmount: 100e8,

		DowntimeSlashAmount: 50e8,

		SubmitterReward: 10e8,

		BscSideChainId: "bsc",
	}
}

// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
func (k Keeper) MaxEvidenceAge(ctx sdk.Context) (res time.Duration) {
	k.paramspace.Get(ctx, KeyMaxEvidenceAge, &res)
	return
}

// SignedBlocksWindow - sliding window for downtime slashing
func (k Keeper) SignedBlocksWindow(ctx sdk.Context) (res int64) {
	k.paramspace.Get(ctx, KeySignedBlocksWindow, &res)
	return
}

// Downtime slashing thershold - default 50% of the SignedBlocksWindow
func (k Keeper) MinSignedPerWindow(ctx sdk.Context) int64 {
	var minSignedPerWindow sdk.Dec
	k.paramspace.Get(ctx, KeyMinSignedPerWindow, &minSignedPerWindow)
	signedBlocksWindow := k.SignedBlocksWindow(ctx)
	return sdk.NewDec(signedBlocksWindow).Mul(minSignedPerWindow).RawInt()
}

// Double-sign unbond duration
func (k Keeper) DoubleSignUnbondDuration(ctx sdk.Context) (res time.Duration) {
	k.paramspace.Get(ctx, KeyDoubleSignUnbondDuration, &res)
	return
}

// Downtime unbond duration
func (k Keeper) DowntimeUnbondDuration(ctx sdk.Context) (res time.Duration) {
	k.paramspace.Get(ctx, KeyDowntimeUnbondDuration, &res)
	return
}

func (k Keeper) TooLowDelUnbondDuration(ctx sdk.Context) (res time.Duration) {
	k.paramspace.Get(ctx, KeyTooLowDelUnbondDuration, &res)
	return
}

// SlashFractionDoubleSign - currently default 5%
func (k Keeper) SlashFractionDoubleSign(ctx sdk.Context) (res sdk.Dec) {
	k.paramspace.Get(ctx, KeySlashFractionDoubleSign, &res)
	return
}

// SlashFractionDowntime - currently default 1%
func (k Keeper) SlashFractionDowntime(ctx sdk.Context) (res sdk.Dec) {
	k.paramspace.Get(ctx, KeySlashFractionDowntime, &res)
	return
}

func (k Keeper) DoubleSignSlashAmount(ctx sdk.Context) (slashAmt int64) {
	k.paramspace.Get(ctx, KeyDoubleSignSlashAmount, &slashAmt)
	return
}

func (k Keeper) DowntimeSlashAmount(ctx sdk.Context) (slashAmt int64) {
	k.paramspace.Get(ctx, KeyDowntimeSlashAmount, &slashAmt)
	return
}

func (k Keeper) SubmitterReward(ctx sdk.Context) (submitterReward int64) {
	k.paramspace.Get(ctx, KeySubmitterReward, &submitterReward)
	return
}

func (k Keeper) BscSideChainId(ctx sdk.Context) (sideChainId string) {
	k.paramspace.Get(ctx, KeyBscSideChainId, &sideChainId)
	return
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params Params) {
	k.paramspace.SetParamSet(ctx, &params)
}
