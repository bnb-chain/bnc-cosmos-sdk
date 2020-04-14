package types

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Type that represents Claim Type as a byte
type ClaimType byte

const (
	ClaimTypeSkipSequence      ClaimType = 0x1
	ClaimTypeUpdateBind        ClaimType = 0x2
	ClaimTypeUpdateTransferOut ClaimType = 0x3
	ClaimTypeTransferIn        ClaimType = 0x4

	ClaimTypeSkipSequenceName      = "SkipSequence"
	ClaimTypeUpdateBindName        = "UpdateBind"
	ClaimTypeUpdateTransferOutName = "UpdateTransferOut"
	ClaimTypeTransferInName        = "TransferIn"
)

var claimTypeToName = map[ClaimType]string{
	ClaimTypeSkipSequence:      ClaimTypeSkipSequenceName,
	ClaimTypeUpdateBind:        ClaimTypeUpdateBindName,
	ClaimTypeUpdateTransferOut: ClaimTypeUpdateTransferOutName,
	ClaimTypeTransferIn:        ClaimTypeTransferInName,
}

var claimNameToType = map[string]ClaimType{
	ClaimTypeSkipSequenceName:      ClaimTypeSkipSequence,
	ClaimTypeUpdateBindName:        ClaimTypeUpdateBind,
	ClaimTypeUpdateTransferOutName: ClaimTypeUpdateTransferOut,
	ClaimTypeTransferInName:        ClaimTypeTransferIn,
}

// String to claim type byte.  Returns ff if invalid.
func ClaimTypeFromString(str string) (ClaimType, error) {
	claimType, ok := claimNameToType[str]
	if !ok {
		return ClaimType(0xff), errors.Errorf("'%s' is not a valid claim type", str)
	}
	return claimType, nil
}

func ClaimTypeToString(typ ClaimType) string {
	return claimTypeToName[typ]
}

func IsValidClaimType(ct ClaimType) bool {
	if _, ok := claimTypeToName[ct]; ok {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (ct ClaimType) Marshal() ([]byte, error) {
	return []byte{byte(ct)}, nil
}

// Unmarshal needed for protobuf compatibility
func (ct *ClaimType) Unmarshal(data []byte) error {
	*ct = ClaimType(data[0])
	return nil
}

// Marshals to JSON using string
func (ct ClaimType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (ct *ClaimType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := ClaimTypeFromString(s)
	if err != nil {
		return err
	}
	*ct = bz2
	return nil
}

// Turns VoteOption byte to String
func (ct ClaimType) String() string {
	claimTypeName, _ := claimTypeToName[ct]
	return claimTypeName
}

var claimHooksMap = map[ClaimType]ClaimHooks{}

func GetClaimHooks(claimType ClaimType) ClaimHooks {
	return claimHooksMap[claimType]
}

func RegisterClaimHooks(claimType ClaimType, hooks ClaimHooks) error {
	_, ok := claimHooksMap[claimType]
	if ok {
		return fmt.Errorf("hooks of claim type %s already exists", claimType.String())
	}
	claimHooksMap[claimType] = hooks
	return nil
}

type ClaimHooks interface {
	CheckClaim(ctx sdk.Context, claim string) sdk.Error
	ExecuteClaim(ctx sdk.Context, prophecy Prophecy) (sdk.Tags, sdk.Error)
}

func GetClaimId(claimType ClaimType, sequence int64) string {
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
