package types

import (
	"bytes"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/sidechain"
)

// name to identify transaction routes
const (
	MsgRoute              = "slashingsidechain"
	TypeMsgSubmitEvidence = "submit_evidence"
)

type MsgSubmitEvidence struct {
	Submitter   sdk.AccAddress       `json:"submitter"`
	SideChainId string               `json:"side_chain_id"`
	Headers     [2]*sidechain.Header `json:"headers"`
}

func NewMsgSubmitEvidence(submitter sdk.AccAddress, sideChainId string, headers [2]*sidechain.Header) MsgSubmitEvidence {

	return MsgSubmitEvidence{
		Submitter:   submitter,
		SideChainId: sideChainId,
		Headers:     headers,
	}
}

var _ sdk.Msg = &MsgSubmitEvidence{}

func (MsgSubmitEvidence) Route() string {
	return MsgRoute
}

func (MsgSubmitEvidence) Type() string {
	return TypeMsgSubmitEvidence
}

func (msg MsgSubmitEvidence) ValidateBasic() sdk.Error {
	if len(msg.Submitter) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.Submitter)))
	}

	if msg.Headers[0] == nil || msg.Headers[1] == nil {
		return sdk.ErrInvalidEvidence("Both two block headers can not be nil")
	}
	if msg.Headers[0].Number.Cmp(msg.Headers[1].Number) != 0 {
		return sdk.ErrInvalidEvidence("The numbers of two block headers are not the same")
	}
	if msg.Headers[0].ParentHash.Cmp(msg.Headers[1].ParentHash) != 0 {
		return sdk.ErrInvalidEvidence("The parent hash of two block headers are not the same")
	}
	signature1, err := msg.Headers[0].GetSignature()
	if err != nil {
		return sdk.ErrInvalidEvidence(fmt.Sprintf("Failed to get signature from block header, %s", err.Error()))
	}
	signature2, err := msg.Headers[1].GetSignature()
	if err != nil {
		return sdk.ErrInvalidEvidence(fmt.Sprintf("Failed to get signature from block header, %s", err.Error()))
	}
	if bytes.Compare(signature1, signature2) == 0 {
		return sdk.ErrInvalidEvidence("The two blocks are the same")
	}

	signer1, err := msg.Headers[0].ExtractSignerFromHeader()
	if err != nil {
		return sdk.ErrInvalidEvidence(fmt.Sprintf("Failed to extract signer from block header, %s", err.Error()))
	}
	signer2, err := msg.Headers[1].ExtractSignerFromHeader()
	if err != nil {
		return sdk.ErrInvalidEvidence(fmt.Sprintf("Failed to extract signer from block header, %s", err.Error()))
	}
	if bytes.Compare(signer1.Bytes(), signer2.Bytes()) != 0 {
		return sdk.ErrInvalidEvidence("The signers of two block headers are not the same")
	}

	return nil
}

func (msg MsgSubmitEvidence) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSubmitEvidence) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Submitter}
}

func (msg MsgSubmitEvidence) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}
