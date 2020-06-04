package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const StartSequence = 0

var claimTypeSequencePrefix = []byte("claimTypeSeq:")

func GetClaimTypeSequence(claimType sdk.ClaimType) []byte {
	return append(claimTypeSequencePrefix, byte(claimType))
}

const (
	ClaimResultCode = "ClaimResultCode"
	ClaimResultMsg  = "ClaimResultMsg"
)
