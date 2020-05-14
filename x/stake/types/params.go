package types

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
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
	defaultMinSelfDelegation int64 = 10e8

	// defaultMinDelegationChanged represents the default minimal allowed amount for delegator to transfer their delegation tokens, including delegate, unDelegate, reDelegate
	defaultMinDelegationChange int64 = 1e8
)

// nolint - Keys for parameter access
var (
	KeyUnbondingTime       = []byte("UnbondingTime")
	KeyMaxValidators       = []byte("MaxValidators")
	KeyBondDenom           = []byte("BondDenom")
	KeyMinSelfDelegation   = []byte("MinSelfDelegation")
	KeyMinDelegationChange = []byte("MinDelegationChanged")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators       uint16 `json:"max_validators"`        // maximum number of validators
	BondDenom           string `json:"bond_denom"`            // bondable coin denomination
	MinSelfDelegation   int64  `json:"min_self_delegation"`   // the minimal self-delegation amount
	MinDelegationChange int64  `json:"min_delegation_change"` // the minimal delegation amount changed
}

// Implements params.ParamSet
func (p *Params) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{KeyUnbondingTime, &p.UnbondingTime},
		{KeyMaxValidators, &p.MaxValidators},
		{KeyBondDenom, &p.BondDenom},
		{KeyMinSelfDelegation, &p.MinSelfDelegation},
		{KeyMinDelegationChange, &p.MinDelegationChange},
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
		UnbondingTime:       defaultUnbondingTime,
		MaxValidators:       100,
		BondDenom:           "steak",
		MinSelfDelegation:   defaultMinSelfDelegation,
		MinDelegationChange: defaultMinDelegationChange,
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
