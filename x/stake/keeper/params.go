package keeper

import (
	"time"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Default parameter namespace
const (
	DefaultParamspace = "stake"
)

var (
	FeeCollectorAddr       = sdk.AccAddress(crypto.AddressHash([]byte("FeeCollector")))
	DelegationAccAddr      = sdk.AccAddress(crypto.AddressHash([]byte("BinanceChainStakeDelegation")))
	FeeForAllBcValsAccAddr = sdk.AccAddress(crypto.AddressHash([]byte("BinanceChainStakeFeeForAllBcVals")))
)

// ParamTable for stake module
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable().RegisterParamSet(&types.Params{})
}

// UnbondingTime
func (k Keeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyUnbondingTime, &res)
	return
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) (res uint16) {
	k.paramstore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyBondDenom, &res)
	return
}

func (k Keeper) MinSelfDelegation(ctx sdk.Context) (res int64) {
	k.paramstore.GetIfExists(ctx, types.KeyMinSelfDelegation, &res)
	return
}

func (k Keeper) MaxStakeSnapshots(ctx sdk.Context) (res uint16) {
	k.paramstore.GetIfExists(ctx, types.KeyMaxStakeSnapshots, &res)
	return
}

func (k Keeper) MinDelegationChange(ctx sdk.Context) (res int64) {
	k.paramstore.GetIfExists(ctx, types.KeyMinDelegationChange, &res)
	return
}

func (k Keeper) RewardDistributionBatchSize(ctx sdk.Context) (res int64) {
	k.paramstore.GetIfExists(ctx, types.KeyRewardDistributionBatchSize, &res)
	return
}

func (k Keeper) BaseProposerRewardRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.GetIfExists(ctx, types.KeyBaseProposerRewardRatio, &res)
	return
}

func (k Keeper) BonusProposerRewardRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.GetIfExists(ctx, types.KeyBonusProposerRewardRatio, &res)
	return
}

func (k Keeper) FeeFromBscToBcRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.GetIfExists(ctx, types.KeyFeeFromBscToBcRatio, &res)
	return
}

// Get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (res types.Params) {
	res.UnbondingTime = k.UnbondingTime(ctx)
	res.MaxValidators = k.MaxValidators(ctx)
	res.BondDenom = k.BondDenom(ctx)
	res.MinSelfDelegation = k.MinSelfDelegation(ctx)
	res.MinDelegationChange = k.MinDelegationChange(ctx)
	res.RewardDistributionBatchSize = k.RewardDistributionBatchSize(ctx)
	res.BaseProposerRewardRatio = k.BaseProposerRewardRatio(ctx)
	res.BonusProposerRewardRatio = k.BonusProposerRewardRatio(ctx)
	res.MaxStakeSnapshots = k.MaxStakeSnapshots(ctx)
	res.FeeFromBscToBcRatio = k.FeeFromBscToBcRatio(ctx)
	return
}

// in order to be compatible with before
type paramBeforeBscUpgrade struct {
	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

// Implements params.ParamSet
func (p *paramBeforeBscUpgrade) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{types.KeyUnbondingTime, &p.UnbondingTime},
		{types.KeyMaxValidators, &p.MaxValidators},
		{types.KeyBondDenom, &p.BondDenom},
	}
}

// in order to be compatible with before
type paramBeforeBEP128Upgrade struct {
	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators       uint16 `json:"max_validators"`        // maximum number of validators
	BondDenom           string `json:"bond_denom"`            // bondable coin denomination
	MinSelfDelegation   int64  `json:"min_self_delegation"`   // the minimal self-delegation amount
	MinDelegationChange int64  `json:"min_delegation_change"` // the minimal delegation amount changed
}

// Implements params.ParamSet
func (p *paramBeforeBEP128Upgrade) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{types.KeyUnbondingTime, &p.UnbondingTime},
		{types.KeyMaxValidators, &p.MaxValidators},
		{types.KeyBondDenom, &p.BondDenom},
		{types.KeyMinSelfDelegation, &p.MinSelfDelegation},
		{types.KeyMinDelegationChange, &p.MinDelegationChange},
	}
}

// in order to be compatible with before
type paramBeforeBEP159Upgrade struct {
	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators       uint16 `json:"max_validators"`        // maximum number of validators
	BondDenom           string `json:"bond_denom"`            // bondable coin denomination
	MinSelfDelegation   int64  `json:"min_self_delegation"`   // the minimal self-delegation amount
	MinDelegationChange int64  `json:"min_delegation_change"` // the minimal delegation amount changed

	RewardDistributionBatchSize int64 `json:"reward_distribution_batch_size"` // the batch size for distributing rewards in blocks
}

// Implements params.ParamSet
func (p *paramBeforeBEP159Upgrade) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{types.KeyUnbondingTime, &p.UnbondingTime},
		{types.KeyMaxValidators, &p.MaxValidators},
		{types.KeyBondDenom, &p.BondDenom},
		{types.KeyMinSelfDelegation, &p.MinSelfDelegation},
		{types.KeyMinDelegationChange, &p.MinDelegationChange},
		{types.KeyRewardDistributionBatchSize, &p.RewardDistributionBatchSize},
	}
}

// set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.Set(ctx, types.KeyUnbondingTime, params.UnbondingTime)
	k.paramstore.Set(ctx, types.KeyMaxValidators, params.MaxValidators)
	k.paramstore.Set(ctx, types.KeyBondDenom, params.BondDenom)
	if sdk.IsUpgrade(sdk.LaunchBscUpgrade) {
		if ctx.ChainID() == sdk.ChainIdGanges {
			k.paramstore.Set(ctx, types.KeyMaxValidators, uint16(11))
			k.paramstore.Set(ctx, types.KeyMinSelfDelegation, int64(5000000000000))
			k.paramstore.Set(ctx, types.KeyMinDelegationChange, params.MinDelegationChange)
		} else {
			k.paramstore.Set(ctx, types.KeyMinSelfDelegation, params.MinSelfDelegation)
			k.paramstore.Set(ctx, types.KeyMinDelegationChange, params.MinDelegationChange)
		}
	}
	if sdk.IsUpgrade(sdk.BEP128) {
		k.paramstore.Set(ctx, types.KeyRewardDistributionBatchSize, params.RewardDistributionBatchSize)
	}
	if sdk.IsUpgrade(sdk.BEP159) {
		k.paramstore.Set(ctx, types.KeyMaxStakeSnapshots, params.MaxStakeSnapshots)
		k.paramstore.Set(ctx, types.KeyBaseProposerRewardRatio, params.BaseProposerRewardRatio)
		k.paramstore.Set(ctx, types.KeyBonusProposerRewardRatio, params.BonusProposerRewardRatio)
		k.paramstore.Set(ctx, types.KeyFeeFromBscToBcRatio, params.FeeFromBscToBcRatio)
	}
}
