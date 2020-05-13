package types

type ClaimResult struct {
	Code int
	Msg  string
	Tags Tags
}

type ClaimHooks interface {
	CheckClaim(ctx Context, claim string) Error
	ExecuteClaim(ctx Context, finalClaim string) (ClaimResult, Error)
}

// Type that represents Claim Type as a byte
type ClaimType byte

type OracleKeeper interface {
	GetClaimTypeName(claimType ClaimType) string
	GetCurrentSequence(ctx Context, claimType ClaimType) int64
	IncreaseSequence(ctx Context, claimType ClaimType) int64
	RegisterClaimType(claimType ClaimType, claimTypeName string, hooks ClaimHooks) error
}
