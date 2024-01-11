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

	MsgTypeSideChainStakeMigration = "side_stake_migration"

	StakeMigrationRelayFee int64 = 500000 // decimal 8
)

type StakeMigrationSynPackage struct {
	OperatorAddress  sdk.SmartChainAddress
	DelegatorAddress sdk.SmartChainAddress
	RefundAddress    sdk.AccAddress
	Amount           *big.Int
}

type MsgSideChainStakeMigration struct {
	ValidatorSrcAddr sdk.ValAddress        `json:"validator_src_addr"`
	ValidatorDstAddr sdk.SmartChainAddress `json:"ValidatorDstAddr"`
	DelegatorAddr    sdk.SmartChainAddress `json:"delegator_addr"`
	RefundAddr       sdk.AccAddress        `json:"refund_addr"`
	Amount           sdk.Coin              `json:"amount"`
}

func NewMsgSideChainStakeMigration(valAddr sdk.ValAddress, operatorAddr, delegatorAddr sdk.SmartChainAddress, refundAddr sdk.AccAddress, amount sdk.Coin) MsgSideChainStakeMigration {
	return MsgSideChainStakeMigration{
		ValidatorSrcAddr: valAddr,
		ValidatorDstAddr: operatorAddr,
		DelegatorAddr:    delegatorAddr,
		RefundAddr:       refundAddr,
		Amount:           amount,
	}
}

func (msg MsgSideChainStakeMigration) Route() string { return MsgRoute }
func (msg MsgSideChainStakeMigration) Type() string  { return MsgTypeSideChainStakeMigration }
func (msg MsgSideChainStakeMigration) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.RefundAddr}
}

func (msg MsgSideChainStakeMigration) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSideChainStakeMigration) ValidateBasic() sdk.Error {
	if len(msg.ValidatorSrcAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorSrcAddr)))
	}
	if msg.ValidatorDstAddr.IsEmpty() {
		return sdk.ErrInvalidAddress("smart chain operator address is empty")
	}
	if msg.DelegatorAddr.IsEmpty() {
		return sdk.ErrInvalidAddress("smart chain beneficiary address is empty")
	}
	if len(msg.RefundAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected refund address length is %d, actual length is %d", sdk.AddrLen, len(msg.RefundAddr)))
	}
	if msg.Amount.Amount <= 0 {
		return ErrBadDelegationAmount(DefaultCodespace, "stake migration amount must be positive")
	}
	return nil
}

func (msg MsgSideChainStakeMigration) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.RefundAddr, sdk.AccAddress(msg.ValidatorSrcAddr)}
}
