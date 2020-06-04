package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	RouteOracle = "oracle"

	ClaimMsgType = "oracleClaim"
)

var _ sdk.Msg = ClaimMsg{}

type ClaimMsg struct {
	ClaimType        sdk.ClaimType  `json:"claim_type"`
	Sequence         int64          `json:"sequence"`
	Claim            string         `json:"claim"`
	ValidatorAddress sdk.AccAddress `json:"validator_address"`
}

func NewClaimMsg(claimType sdk.ClaimType, sequence int64, claim string, validatorAddr sdk.AccAddress) ClaimMsg {
	return ClaimMsg{
		ClaimType:        claimType,
		Sequence:         sequence,
		Claim:            claim,
		ValidatorAddress: validatorAddr,
	}
}

// nolint
func (msg ClaimMsg) Route() string { return RouteOracle }
func (msg ClaimMsg) Type() string  { return ClaimMsgType }
func (msg ClaimMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.ValidatorAddress}
}

func (msg ClaimMsg) String() string {
	return fmt.Sprintf("Claim{%v#%v#%v#%v}",
		msg.ClaimType, msg.Sequence, msg.Claim, msg.ValidatorAddress.String())
}

// GetSignBytes - Get the bytes for the message signer to sign on
func (msg ClaimMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func (msg ClaimMsg) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}

// ValidateBasic is used to quickly disqualify obviously invalid messages quickly
func (msg ClaimMsg) ValidateBasic() sdk.Error {
	if msg.Sequence < 0 {
		return ErrInvalidSequence("sequence should not be less than 0")
	}

	if len(msg.Claim) == 0 {
		return ErrInvalidClaim()
	}

	if len(msg.ValidatorAddress) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(msg.ValidatorAddress.String())
	}
	return nil
}
