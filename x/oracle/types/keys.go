package types

const StartSequence = 0

var claimTypeSequencePrefix = []byte("claimTypeSeq:")

func GetClaimTypeSequence(claimType ClaimType) []byte {
	return append(claimTypeSequencePrefix, byte(claimType))
}
