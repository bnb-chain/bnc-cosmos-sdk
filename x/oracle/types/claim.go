package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetClaimId(claimType sdk.ClaimType, sequence int64) string {
	return fmt.Sprintf("%d:%d", claimType, sequence)
}

// Claim contains an arbitrary claim with arbitrary content made by a given validator
type Claim struct {
	ID               string         `json:"id"`
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
	Content          string         `json:"content"`
}

// NewClaim returns a new Claim
func NewClaim(id string, validatorAddress sdk.ValAddress, content string) Claim {
	return Claim{
		ID:               id,
		ValidatorAddress: validatorAddress,
		Content:          content,
	}
}
