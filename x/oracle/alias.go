package oracle

import (
	"github.com/cosmos/cosmos-sdk/x/oracle/keeper"
	"github.com/cosmos/cosmos-sdk/x/oracle/types"
)

const (
	PendingStatusText = types.PendingStatusText
	SuccessStatusText = types.SuccessStatusText
	FailedStatusText  = types.FailedStatusText
	DefaultParamSpace = keeper.DefaultParamSpace

	ClaimTypeSkipSequence      = types.ClaimTypeSkipSequence
	ClaimTypeUpdateBind        = types.ClaimTypeUpdateBind
	ClaimTypeUpdateTransferOut = types.ClaimTypeUpdateTransferOut
	ClaimTypeTransferIn        = types.ClaimTypeTransferIn
)

var (
	// functions aliases
	NewKeeper = keeper.NewKeeper

	NewClaim                         = types.NewClaim
	ErrProphecyNotFound              = types.ErrProphecyNotFound
	ErrMinimumConsensusNeededInvalid = types.ErrMinimumConsensusNeededInvalid
	ErrNoClaims                      = types.ErrNoClaims
	ErrInvalidIdentifier             = types.ErrInvalidIdentifier
	ErrProphecyFinalized             = types.ErrProphecyFinalized
	ErrDuplicateMessage              = types.ErrDuplicateMessage
	ErrInvalidClaim                  = types.ErrInvalidClaim
	ErrInvalidValidator              = types.ErrInvalidValidator
	ErrInternalDB                    = types.ErrInternalDB
	ErrInvalidClaimType              = types.ErrInvalidClaimType

	NewProphecy = types.NewProphecy
	NewStatus   = types.NewStatus

	// variable aliases
	StatusTextToString = types.StatusTextToString
	StringToStatusText = types.StringToStatusText

	RegisterClaimHooks = types.RegisterClaimHooks

	IsValidClaimType = types.IsValidClaimType

	NewClaimMsg = types.NewClaimMsg
	RouteOracle = types.RouteOracle
	GetClaimId  = types.GetClaimId
)

type (
	Keeper     = keeper.Keeper
	Claim      = types.Claim
	Prophecy   = types.Prophecy
	DBProphecy = types.DBProphecy
	Status     = types.Status
	StatusText = types.StatusText

	ClaimMsg  = types.ClaimMsg
	ClaimType = types.ClaimType
)
