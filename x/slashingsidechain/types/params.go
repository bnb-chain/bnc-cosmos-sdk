package types

import (
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"time"
)

// Default parameter namespace
const (
	DefaultParamspace = "slashingsidechain"
)

func ParamTypeTable() subspace.TypeTable {

	return params.NewTypeTable().RegisterParamSet(&Params{})

}

var (
	KeySlashAmount     = []byte("SlashAmount")
	KeySubmitterReward = []byte("SubmitterReward")
	KeyMaxEvidenceAge  = []byte("MaxEvidenceAge")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	SlashAmount     int64 `json:"slash_amount"`
	SubmitterReward int64 `json:"submitter_reward"`
	MaxEvidenceAge  time.Duration `json:"max_evidence_age"`
}

func DefaultParams() Params{
	return Params{
		SlashAmount: 10000000000,
		SubmitterReward: 1000000000,
		MaxEvidenceAge: 60 * 60 * 24 * time.Second, // 1 day
	}
}

// Implements params.ParamSet
func (p *Params) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{KeySlashAmount, &p.SlashAmount},
		{KeySubmitterReward, &p.SubmitterReward},
		{KeyMaxEvidenceAge, &p.MaxEvidenceAge},
	}
}


