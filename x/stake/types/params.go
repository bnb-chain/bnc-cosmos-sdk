package types

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// defaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	defaultUnbondingTime time.Duration = 60 * 60 * 24 * 3 * time.Second

	// Delay, in blocks, between when validator updates are returned to Tendermint and when they are applied
	// For example, if this is 0, the validator set at the end of a block will sign the next block, or
	// if this is 1, the validator set at the end of a block will sign the block after the next.
	// Constant as this should not change without a hard fork.
	ValidatorUpdateDelay int64 = 1

	// if the self delegation is below the MinSelfDelegation,
	// the creation of validator would be rejected or the validator would be jailed.
	defaultMinSelfDelegation int64 = 10000e8

	// defaultMinDelegationChange represents the default minimal allowed amount for delegator to transfer their delegation tokens, including delegate, unDelegate, reDelegate
	defaultMinDelegationChange int64 = 1e8

	// defaultRewardDistributionBatchSize represents the default batch size for distributing delegators' staking rewards in blocks
	defaultRewardDistributionBatchSize = 1000

	ConsAddrUpdateIntervalInHours = 24 * 30
)

// nolint - Keys for parameter access
var (
	KeyUnbondingTime               = []byte("UnbondingTime")
	KeyMaxValidators               = []byte("MaxValidators")
	KeyBondDenom                   = []byte("BondDenom")
	KeyMinSelfDelegation           = []byte("MinSelfDelegation")
	KeyMinDelegationChange         = []byte("MinDelegationChanged")
	KeyRewardDistributionBatchSize = []byte("RewardDistributionBatchSize")
	KeyMaxStakeSnapshots           = []byte("MaxStakeSnapshots")
	KeyBaseProposerRewardRatio     = []byte("BaseProposerRewardRatio")
	KeyBonusProposerRewardRatio    = []byte("BonusProposerRewardRatio")
	KeyFeeFromBscToBcRatio         = []byte("FeeFromBscToBcRatio")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators               uint16 `json:"max_validators"`                 // maximum number of validators
	BondDenom                   string `json:"bond_denom"`                     // bondable coin denomination
	MinSelfDelegation           int64  `json:"min_self_delegation"`            // the minimal self-delegation amount
	MinDelegationChange         int64  `json:"min_delegation_change"`          // the minimal delegation amount changed
	RewardDistributionBatchSize int64  `json:"reward_distribution_batch_size"` // the batch size for distributing rewards in blocks
	// added in BEP159
	MaxStakeSnapshots        uint16    `json:"max_stake_snapshots"`         // maximum number of stake snapshots, also used as the accumulated stake duration
	BaseProposerRewardRatio  types.Dec `json:"base_proposer_reward_ratio"`  // the base proposer reward ratio
	BonusProposerRewardRatio types.Dec `json:"bonus_proposer_reward_ratio"` // the bonus proposer reward ratio
	FeeFromBscToBcRatio      types.Dec `json:"fee_from_bsc_to_bc_ratio"`    // the fee from bsc to bc ratio
}

func (p *Params) GetBCParamAttribute() string {
	return "staking"
}

func (p *Params) GetParamAttribute() (string, bool) {
	return "staking", false
}

func (p *Params) UpdateCheck() error {
	if p.BondDenom != types.NativeTokenSymbol {
		return fmt.Errorf("only native token is availabe as bond_denom so far")
	}
	// the valid range is 1 minute to 100 day.
	if p.UnbondingTime > 100*24*time.Hour || p.UnbondingTime < time.Minute {
		return fmt.Errorf("the UnbondingTime should be in range 1 minute to 100 days")
	}
	if p.MaxValidators < 1 || p.MaxValidators > 500 {
		return fmt.Errorf("the max validator should be in range 1 to 500")
	}
	// BondDenom do not check here, it should be native token and do not support update so far.
	// Leave the check in node repo.

	if p.MinSelfDelegation > 10000000e8 || p.MinSelfDelegation < 1e8 {
		return fmt.Errorf("the min_self_delegation should be in range 1e8 to 10000000e8")
	}
	if p.MinDelegationChange < 1e5 {
		return fmt.Errorf("the min_delegation_change should be no less than 1e5")
	}

	if p.RewardDistributionBatchSize < 1000 || p.RewardDistributionBatchSize > 5000 {
		return fmt.Errorf("the reward_distribution_batch_size should be in range 1000 to 5000")
	}
	if p.BaseProposerRewardRatio.LT(types.ZeroDec()) {
		return fmt.Errorf("the base_proposer_reward_ratio should be no less than 0")
	}
	if p.BonusProposerRewardRatio.LT(types.ZeroDec()) {
		return fmt.Errorf("the bonus_proposer_reward_ratio should be no less than 0")
	}
	if p.BonusProposerRewardRatio.Add(p.BaseProposerRewardRatio).GT(types.OneDec()) {
		return fmt.Errorf("the base_proposer_reward_ratio + bonus_proposer_reward_ratio should be no greater than 1")
	}
	if p.FeeFromBscToBcRatio.LT(types.ZeroDec()) {
		return fmt.Errorf("the fee_from_bsc_to_bc_ratio should be no less than 0")
	}

	return nil
}

// Implements params.ParamSet
func (p *Params) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{KeyUnbondingTime, &p.UnbondingTime},
		{KeyMaxValidators, &p.MaxValidators},
		{KeyBondDenom, &p.BondDenom},
		{KeyMinSelfDelegation, &p.MinSelfDelegation},
		{KeyMinDelegationChange, &p.MinDelegationChange},
		{KeyRewardDistributionBatchSize, &p.RewardDistributionBatchSize},
		{KeyMaxStakeSnapshots, &p.MaxStakeSnapshots},
		{KeyBaseProposerRewardRatio, &p.BaseProposerRewardRatio},
		{KeyBonusProposerRewardRatio, &p.BonusProposerRewardRatio},
		{KeyFeeFromBscToBcRatio, &p.FeeFromBscToBcRatio},
	}
}

// Equal returns a boolean determining if two Param types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := MsgCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := MsgCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		UnbondingTime:               defaultUnbondingTime,
		MaxValidators:               100,
		BondDenom:                   "steak",
		MinSelfDelegation:           defaultMinSelfDelegation,
		MinDelegationChange:         defaultMinDelegationChange,
		RewardDistributionBatchSize: defaultRewardDistributionBatchSize,
		MaxStakeSnapshots:           30,
		BaseProposerRewardRatio:     types.NewDec(1e6),
		BonusProposerRewardRatio:    types.NewDec(4e6),
		FeeFromBscToBcRatio:         types.NewDec(1e7),
	}
}

// HumanReadableString returns a human readable string representation of the
// parameters.
func (p Params) HumanReadableString() string {

	resp := "Params \n"
	resp += fmt.Sprintf("Unbonding Time: %s\n", p.UnbondingTime)
	resp += fmt.Sprintf("Max Validators: %d: \n", p.MaxValidators)
	resp += fmt.Sprintf("Bonded Coin Denomination: %s\n", p.BondDenom)
	resp += fmt.Sprintf("Minimal self-delegation amount: %d\n", p.MinSelfDelegation)
	resp += fmt.Sprintf("The minimum value allowed to change the delegation amount: %d\n", p.MinDelegationChange)
	resp += fmt.Sprintf("The batch size to distribute staking rewards: %d\n", p.RewardDistributionBatchSize)
	resp += fmt.Sprintf("Max stake snapshots: %d\n", p.MaxStakeSnapshots)
	resp += fmt.Sprintf("Base proposer reward ratio: %s\n", p.BaseProposerRewardRatio)
	resp += fmt.Sprintf("Bonus proposer reward ratio: %s\n", p.BonusProposerRewardRatio)
	resp += fmt.Sprintf("Fee from BSC to BC ratio: %s\n", p.FeeFromBscToBcRatio)
	return resp
}

// unmarshal the current staking params value from store key or panic
func MustUnmarshalParams(cdc *codec.Codec, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}
	return params
}

// unmarshal the current staking params value from store key
func UnmarshalParams(cdc *codec.Codec, value []byte) (params Params, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &params)
	if err != nil {
		return
	}
	return
}
