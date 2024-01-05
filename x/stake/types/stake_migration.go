package types

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	StakeMigrationChannel = "stakeMigration"

	StakeMigrationChannelID sdk.ChannelID = 17

	TagStakeMigrationSendSequence = "StakeMigrationSendSequence"

	MsgTypeSideChainStakeMigration = "stake_migration"

	StakeMigrationRelayFee int64 = 1000000 // decimal 8
)

type StakeMigrationSynPackage struct {
	OperatorAddress  sdk.SmartChainAddress
	DelegatorAddress sdk.SmartChainAddress
	RefundAddress    sdk.AccAddress
	Amount           *big.Int
}

type MsgSideChainStakeMigration struct {
	Validator        sdk.ValAddress        `json:"validator"`
	OperatorAddress  sdk.SmartChainAddress `json:"operator_address"`
	DelegatorAddress sdk.SmartChainAddress `json:"delegator_address"`
	RefundAddress    sdk.AccAddress        `json:"refund_address"`
	Amount           sdk.Coin              `json:"amount"`
}

func NewMsgSideChainStakeMigration(valAddr sdk.ValAddress, operatorAddr, delegatorAddr sdk.SmartChainAddress, refundAddr sdk.AccAddress, amount sdk.Coin) MsgSideChainStakeMigration {
	return MsgSideChainStakeMigration{
		Validator:        valAddr,
		OperatorAddress:  operatorAddr,
		DelegatorAddress: delegatorAddr,
		RefundAddress:    refundAddr,
		Amount:           amount,
	}
}

func (msg MsgSideChainStakeMigration) Route() string { return MsgRoute }
func (msg MsgSideChainStakeMigration) Type() string  { return MsgTypeSideChainStakeMigration }
func (msg MsgSideChainStakeMigration) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.RefundAddress}
}

func (msg MsgSideChainStakeMigration) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSideChainStakeMigration) ValidateBasic() sdk.Error {
	if len(msg.Validator) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.Validator)))
	}
	if msg.OperatorAddress.IsEmpty() {
		return sdk.ErrInvalidAddress("operator address is empty")
	}
	if msg.DelegatorAddress.IsEmpty() {
		return sdk.ErrInvalidAddress("delegator address is empty")
	}
	if len(msg.RefundAddress) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected refund address length is %d, actual length is %d", sdk.AddrLen, len(msg.RefundAddress)))
	}
	if msg.Amount.Amount <= 0 {
		return ErrBadDelegationAmount(DefaultCodespace, "stake migration amount must be positive")
	}
	return nil
}

func (msg MsgSideChainStakeMigration) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.RefundAddress, sdk.AccAddress(msg.Validator)}
}
