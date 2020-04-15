package types

type ClaimHooks interface {
	CheckClaim(ctx Context, claim string) Error
	ExecuteClaim(ctx Context, finalClaim string) (Tags, Error)
}

// Type that represents Claim Type as a byte
type ClaimType byte
