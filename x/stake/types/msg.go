package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// name to identify transaction routes
const MsgRoute = "stake"

// Verify interface at compile time
var _, _, _ sdk.Msg = &MsgCreateValidator{}, &MsgEditValidator{}, &MsgDelegate{}

//______________________________________________________________________

// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	Description   Description
	Commission    CommissionMsg
	DelegatorAddr sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr sdk.ValAddress `json:"validator_address"`
	PubKey        crypto.PubKey  `json:"pubkey"`
	Delegation    sdk.Coin       `json:"delegation"`
}

type CreateValidatorJsonMsg struct {
	Description   Description
	Commission    CommissionMsg
	DelegatorAddr sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr sdk.ValAddress `json:"validator_address"`
	PubKey        []byte         `json:"pubkey"`
	Delegation    sdk.Coin       `json:"delegation"`
}

func (jsonMsg CreateValidatorJsonMsg) ToMsgCreateValidator() (MsgCreateValidator, error) {
	if len(jsonMsg.PubKey) != ed25519.PubKeyEd25519Size {
		return MsgCreateValidator{}, fmt.Errorf("pubkey size should be %d", ed25519.PubKeyEd25519Size)
	}

	var pubkey ed25519.PubKeyEd25519
	copy(pubkey[:], jsonMsg.PubKey)

	return MsgCreateValidator{
		Description:   jsonMsg.Description,
		Commission:    jsonMsg.Commission,
		DelegatorAddr: jsonMsg.DelegatorAddr,
		ValidatorAddr: jsonMsg.ValidatorAddr,
		PubKey:        pubkey,
		Delegation:    jsonMsg.Delegation,
	}, nil
}

type MsgCreateValidatorProposal struct {
	MsgCreateValidator
	ProposalId int64 `json:"proposal_id"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(valAddr sdk.ValAddress, pubkey crypto.PubKey,
	selfDelegation sdk.Coin, description Description, commission CommissionMsg) MsgCreateValidator {

	return NewMsgCreateValidatorOnBehalfOf(
		sdk.AccAddress(valAddr), valAddr, pubkey, selfDelegation, description, commission,
	)
}

// Creates validator msg by delegator address on behalf of validator address
func NewMsgCreateValidatorOnBehalfOf(delAddr sdk.AccAddress, valAddr sdk.ValAddress,
	pubkey crypto.PubKey, delegation sdk.Coin, description Description, commission CommissionMsg) MsgCreateValidator {
	return MsgCreateValidator{
		Description:   description,
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		PubKey:        pubkey,
		Delegation:    delegation,
		Commission:    commission,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return MsgRoute }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{msg.DelegatorAddr}

	if !bytes.Equal(msg.DelegatorAddr.Bytes(), msg.ValidatorAddr.Bytes()) {
		// if validator addr is not same as delegator addr, validator must sign
		// msg as well
		addrs = append(addrs, sdk.AccAddress(msg.ValidatorAddr))
	}
	return addrs
}

// get the bytes for the message signer to sign on
func (msg MsgCreateValidator) GetSignBytes() []byte {
	b := MsgCdc.MustMarshalJSON(struct {
		Description
		DelegatorAddr sdk.AccAddress `json:"delegator_address"`
		ValidatorAddr sdk.ValAddress `json:"validator_address"`
		PubKey        string         `json:"pubkey"`
		Delegation    sdk.Coin       `json:"delegation"`
	}{
		Description:   msg.Description,
		ValidatorAddr: msg.ValidatorAddr,
		PubKey:        sdk.MustBech32ifyConsPub(msg.PubKey),
		Delegation:    msg.Delegation,
	})
	return sdk.MustSortJSON(b)
}

func (msg MsgCreateValidator) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}
	if msg.Delegation.Amount < 1e8 {
		return ErrBadDelegationAmount(DefaultCodespace, "self delegation must not be less than 1e8")
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}
	commission := NewCommission(msg.Commission.Rate, msg.Commission.MaxRate, msg.Commission.MaxChangeRate)
	if err := commission.Validate(); err != nil {
		return err
	}

	return nil
}

func (msg MsgCreateValidator) Equals(other MsgCreateValidator) bool {
	if !msg.Commission.Equal(other.Commission) {
		return false
	}

	if !msg.PubKey.Equals(other.PubKey) {
		return false
	}

	return msg.Delegation.IsEqual(other.Delegation) &&
		msg.DelegatorAddr.Equals(other.DelegatorAddr) &&
		msg.ValidatorAddr.Equals(other.ValidatorAddr) &&
		msg.PubKey.Equals(other.PubKey) &&
		msg.Description.Equals(other.Description)
}

//______________________________________________________________________

// the MsgCreateValidator has been taken by MsgCreateValidatorProposal,
// use a new Msg for CreateValidatorOpen for open staking, and change pubkey to string(bech32 format)
// to make it friendly for multi-language implementation.
type MsgCreateValidatorOpen struct {
	Description   Description    `json:"description"`
	Commission    CommissionMsg  `json:"commission"`
	DelegatorAddr sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr sdk.ValAddress `json:"validator_address"`
	PubKey        string         `json:"pubkey"`
	Delegation    sdk.Coin       `json:"delegation"`
}

func (msg MsgCreateValidatorOpen) Route() string { return MsgRoute }
func (msg MsgCreateValidatorOpen) Type() string  { return "create_validator_open" }

func (msg MsgCreateValidatorOpen) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{msg.DelegatorAddr}

	if !bytes.Equal(msg.DelegatorAddr.Bytes(), msg.ValidatorAddr.Bytes()) {
		// if validator addr is not same as delegator addr, validator must sign
		// msg as well
		addrs = append(addrs, sdk.AccAddress(msg.ValidatorAddr))
	}
	return addrs
}

func (msg MsgCreateValidatorOpen) GetSignBytes() []byte {
	b := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(b)
}

func (msg MsgCreateValidatorOpen) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}

func (msg MsgCreateValidatorOpen) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}
	if msg.Delegation.Amount < 1e8 {
		return ErrBadDelegationAmount(DefaultCodespace, "self delegation must not be less than 1e8")
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}
	commission := NewCommission(msg.Commission.Rate, msg.Commission.MaxRate, msg.Commission.MaxChangeRate)
	if err := commission.Validate(); err != nil {
		return err
	}
	if len(msg.PubKey) != 0 {
		if _, err := sdk.GetConsPubKeyBech32(msg.PubKey); err != nil {
			return sdk.ErrInvalidPubKey(err.Error())
		}
	}

	return nil
}

func (msg MsgCreateValidatorOpen) Equals(other MsgCreateValidatorOpen) bool {
	if !msg.Commission.Equal(other.Commission) {
		return false
	}

	return msg.Delegation.IsEqual(other.Delegation) &&
		msg.DelegatorAddr.Equals(other.DelegatorAddr) &&
		msg.ValidatorAddr.Equals(other.ValidatorAddr) &&
		msg.PubKey == other.PubKey &&
		msg.Description.Equals(other.Description)
}

//______________________________________________________________________

// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	Description   Description    `json:"description"`
	ValidatorAddr sdk.ValAddress `json:"address"`

	// We pass a reference to the new commission rate as it's not mandatory to
	// update. If not updated, the deserialized rate will be zero with no way to
	// distinguish if an update was intended.
	//
	// REF: #2373
	CommissionRate *sdk.Dec `json:"commission_rate"`
	PubKey         string   `json:"pubkey"`
}

func NewMsgEditValidator(valAddr sdk.ValAddress, description Description, newRate *sdk.Dec, pubkey string) MsgEditValidator {
	return MsgEditValidator{
		Description:    description,
		CommissionRate: newRate,
		ValidatorAddr:  valAddr,
		PubKey:         pubkey,
	}
}

//nolint
func (msg MsgEditValidator) Route() string { return MsgRoute }
func (msg MsgEditValidator) Type() string  { return "edit_validator" }
func (msg MsgEditValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddr)}
}

// get the bytes for the message signer to sign on
func (msg MsgEditValidator) GetSignBytes() []byte {
	b := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "nil validator address")
	}

	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
	}

	if len(msg.PubKey) != 0 {
		if _, err := sdk.GetConsPubKeyBech32(msg.PubKey); err != nil {
			return sdk.ErrInvalidPubKey(err.Error())
		}
	}

	return nil
}

func (msg MsgEditValidator) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgDelegate struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
	Delegation    sdk.Coin       `json:"delegation"`
}

func NewMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, delegation sdk.Coin) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Delegation:    delegation,
	}
}

//nolint
func (msg MsgDelegate) Route() string { return MsgRoute }
func (msg MsgDelegate) Type() string  { return "delegate" }
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgDelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddr == nil {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddr == nil {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Delegation.Amount < 1e8 {
		return ErrBadDelegationAmount(DefaultCodespace, "delegation must not be less than 1e8")
	}
	return nil
}

func (msg MsgDelegate) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr, sdk.AccAddress(msg.ValidatorAddr)}
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgRedelegate struct {
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`
	ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"`
	ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"`
	Amount           sdk.Coin       `json:"amount"`
}

func NewMsgRedelegate(delAddr sdk.AccAddress, valSrcAddr,
	valDstAddr sdk.ValAddress, amount sdk.Coin) MsgRedelegate {

	return MsgRedelegate{
		DelegatorAddr:    delAddr,
		ValidatorSrcAddr: valSrcAddr,
		ValidatorDstAddr: valDstAddr,
		Amount:           amount,
	}
}

//nolint
func (msg MsgRedelegate) Route() string { return MsgRoute }
func (msg MsgRedelegate) Type() string  { return "redelegate" }
func (msg MsgRedelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgRedelegate) GetSignBytes() []byte {
	b := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(b)
}

func (msg MsgRedelegate) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr, sdk.AccAddress(msg.ValidatorSrcAddr), sdk.AccAddress(msg.DelegatorAddr)}
}

// ValidateBasic is used to quickly disqualify obviously invalid messages quickly
func (msg MsgRedelegate) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorSrcAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorSrcAddr)))
	}
	if len(msg.ValidatorDstAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorDstAddr)))
	}
	if msg.Amount.Amount <= 0 {
		return sdk.ErrInvalidCoins(fmt.Sprintf("Expected positive amount, actual amount is %v", msg.Amount.Amount))
	}
	return nil
}

type MsgUndelegate struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
	Amount        sdk.Coin       `json:"amount"`
}

func NewMsgUndelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) MsgUndelegate {
	return MsgUndelegate{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Amount:        amount,
	}
}

//nolint
func (msg MsgUndelegate) Route() string { return MsgRoute }
func (msg MsgUndelegate) Type() string  { return "undelegate" }
func (msg MsgUndelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgUndelegate) GetSignBytes() []byte {
	b := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgUndelegate) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}
	if msg.Amount.Amount <= 0 {
		return ErrBadDelegationAmount(DefaultCodespace, "undelegation amount must be positive")
	}
	return nil
}

func (msg MsgUndelegate) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr, sdk.AccAddress(msg.ValidatorAddr)}
}

// MsgBeginUnbonding - struct for unbonding transactions
type MsgBeginUnbonding struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
	SharesAmount  sdk.Dec        `json:"shares_amount"`
}

func NewMsgBeginUnbonding(delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) MsgBeginUnbonding {
	return MsgBeginUnbonding{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		SharesAmount:  sharesAmount,
	}
}

//nolint
func (msg MsgBeginUnbonding) Route() string { return MsgRoute }
func (msg MsgBeginUnbonding) Type() string  { return "begin_unbonding" }
func (msg MsgBeginUnbonding) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgBeginUnbonding) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(struct {
		DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
		ValidatorAddr sdk.ValAddress `json:"validator_addr"`
		SharesAmount  string         `json:"shares_amount"`
	}{
		DelegatorAddr: msg.DelegatorAddr,
		ValidatorAddr: msg.ValidatorAddr,
		SharesAmount:  msg.SharesAmount.String(),
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgBeginUnbonding) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}
	if msg.SharesAmount.LTE(sdk.ZeroDec()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}

func (msg MsgBeginUnbonding) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr, sdk.AccAddress(msg.ValidatorAddr)}
}

type MsgRemoveValidator struct {
	LauncherAddr sdk.AccAddress  `json:"launcher_addr"`
	ValAddr      sdk.ValAddress  `json:"val_addr"`
	ValConsAddr  sdk.ConsAddress `json:"val_cons_addr"`
	ProposalId   int64           `json:"proposal_id"`
}

func NewMsgRemoveValidator(launcherAddr sdk.AccAddress, valAddr sdk.ValAddress,
	valConsAddr sdk.ConsAddress, proposalId int64) MsgRemoveValidator {
	return MsgRemoveValidator{
		LauncherAddr: launcherAddr,
		ValAddr:      valAddr,
		ValConsAddr:  valConsAddr,
		ProposalId:   proposalId,
	}
}

//nolint
func (msg MsgRemoveValidator) Route() string { return MsgRoute }
func (msg MsgRemoveValidator) Type() string  { return "remove_validator" }
func (msg MsgRemoveValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.LauncherAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgRemoveValidator) GetSignBytes() []byte {
	b, err := MsgCdc.MarshalJSON(struct {
		LauncherAddr sdk.AccAddress  `json:"launcher_addr"`
		ValAddr      sdk.ValAddress  `json:"val_addr"`
		ValConsAddr  sdk.ConsAddress `json:"val_cons_addr"`
		ProposalId   int64           `json:"proposal_id"`
	}{
		LauncherAddr: msg.LauncherAddr,
		ValAddr:      msg.ValAddr,
		ValConsAddr:  msg.ValConsAddr,
		ProposalId:   msg.ProposalId,
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgRemoveValidator) ValidateBasic() sdk.Error {
	if len(msg.LauncherAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected launcher address length is %d, actual length is %d", sdk.AddrLen, len(msg.LauncherAddr)))
	}
	if len(msg.ValAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValAddr)))
	}
	if len(msg.ValConsAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator consensus address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValConsAddr)))
	}
	if msg.ProposalId <= 0 {
		return ErrInvalidProposal(DefaultCodespace, fmt.Sprintf("Proposal id is expected to be positive, actual value is %d", msg.ProposalId))
	}
	return nil
}

func (msg MsgRemoveValidator) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.LauncherAddr}
}
