package gov

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sidechain"
)

const (
	MsgTypeSideSubmitProposal = "side_submit_proposal"
	MsgTypeSideDeposit        = "side_deposit"
	MsgTypeSideVote           = "side_vote"
)

var _, _, _ sdk.Msg = MsgSideChainSubmitProposal{}, MsgSideChainDeposit{}, MsgSideChainVote{}

//-----------------------------------------------------------
// MsgSideChainSubmitProposal
type MsgSideChainSubmitProposal struct {
	MsgSubmitProposal
	SideChainId string `json:"side_chain_id"`
}

func NewMsgSideChainSubmitProposal(title string, description string, proposalType ProposalKind, proposer sdk.AccAddress, initialDeposit sdk.Coins, votingPeriod time.Duration, sideChainId string) MsgSideChainSubmitProposal {
	subMsg := NewMsgSubmitProposal(title, description, proposalType, proposer, initialDeposit, votingPeriod)
	return MsgSideChainSubmitProposal{
		MsgSubmitProposal: subMsg,
		SideChainId:       sideChainId,
	}
}

//nolint
func (msg MsgSideChainSubmitProposal) Route() string { return MsgRoute }
func (msg MsgSideChainSubmitProposal) Type() string  { return MsgTypeSideSubmitProposal }

// Implements Msg.
func (msg MsgSideChainSubmitProposal) ValidateBasic() sdk.Error {
	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > sidechain.MaxSideChainIdLength {
		return ErrInvalidSideChainId(DefaultCodespace, msg.SideChainId)
	}
	if len(msg.Title) == 0 {
		return ErrInvalidTitle(DefaultCodespace, "No title present in proposal")
	}
	if len(msg.Title) > MaxTitleLength {
		return ErrInvalidTitle(DefaultCodespace, fmt.Sprintf("Proposal title is longer than max length of %d", MaxTitleLength))
	}
	if len(msg.Description) == 0 {
		return ErrInvalidDescription(DefaultCodespace, "No description present in proposal")
	}
	if len(msg.Description) > MaxDescriptionLength {
		return ErrInvalidDescription(DefaultCodespace, fmt.Sprintf("Proposal description is longer than max length of %d", MaxDescriptionLength))
	}
	if !validSideProposalType(msg.ProposalType) {
		return ErrInvalidProposalType(DefaultCodespace, msg.ProposalType)
	}
	if len(msg.Proposer) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("length of address(%s) should be %d", string(msg.Proposer), sdk.AddrLen))
	}
	if !msg.InitialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if !msg.InitialDeposit.IsNotNegative() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if msg.VotingPeriod <= 0 || msg.VotingPeriod > MaxVotingPeriod {
		return ErrInvalidVotingPeriod(DefaultCodespace, msg.VotingPeriod)
	}
	return nil
}

func (msg MsgSideChainSubmitProposal) String() string {
	return fmt.Sprintf("MsgSideChainSubmitProposal{%s, %s, %s, %v, %s, %s}", msg.Title, msg.Description, msg.ProposalType, msg.InitialDeposit, msg.VotingPeriod, msg.SideChainId)
}

// Implements Msg.
func (msg MsgSideChainSubmitProposal) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg. Identical to MsgSubmitProposal, keep here for code readability.
func (msg MsgSideChainSubmitProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

// Implements Msg. Identical to MsgSubmitProposal, keep here for code readability.
func (msg MsgSideChainSubmitProposal) GetInvolvedAddresses() []sdk.AccAddress {
	// Better include DepositedCoinsAccAddr, before further discussion, follow the old rule.
	return msg.GetSigners()
}

//-----------------------------------------------------------
// MsgSideChainDeposit
type MsgSideChainDeposit struct {
	MsgDeposit
	SideChainId string `json:"side_chain_id"`
}

func NewMsgSideChainDeposit(depositer sdk.AccAddress, proposalID int64, amount sdk.Coins, sideChainId string) MsgSideChainDeposit {
	subMsg := NewMsgDeposit(depositer, proposalID, amount)
	return MsgSideChainDeposit{
		MsgDeposit:  subMsg,
		SideChainId: sideChainId,
	}
}

// nolint
func (msg MsgSideChainDeposit) Route() string { return MsgRoute }
func (msg MsgSideChainDeposit) Type() string  { return MsgTypeSideDeposit }

// Implements Msg.
func (msg MsgSideChainDeposit) ValidateBasic() sdk.Error {
	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > sidechain.MaxSideChainIdLength {
		return ErrInvalidSideChainId(DefaultCodespace, msg.SideChainId)
	}
	return msg.MsgDeposit.ValidateBasic()
}

func (msg MsgSideChainDeposit) String() string {
	return fmt.Sprintf("MsgSideChainDeposit{%s=>%v: %v, %s}", msg.Depositer, msg.ProposalID, msg.Amount, msg.SideChainId)
}

// Implements Msg.
func (msg MsgSideChainDeposit) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg. Identical to MsgDeposit, keep here for code readability.
func (msg MsgSideChainDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositer}
}

// Implements Msg. Identical to MsgDeposit, keep here for code readability.
func (msg MsgSideChainDeposit) GetInvolvedAddresses() []sdk.AccAddress {
	// Better include DepositedCoinsAccAddr, before further discussion, follow the old rule.
	return msg.GetSigners()
}

//-----------------------------------------------------------
// MsgSideChainVote

type MsgSideChainVote struct {
	MsgVote
	SideChainId string `json:"side_chain_id"`
}

func NewMsgSideChainVote(voter sdk.AccAddress, proposalID int64, option VoteOption, sideChainId string) MsgSideChainVote {
	subMsg := NewMsgVote(voter, proposalID, option)
	return MsgSideChainVote{
		MsgVote:     subMsg,
		SideChainId: sideChainId,
	}
}

func (msg MsgSideChainVote) Route() string { return MsgRoute }
func (msg MsgSideChainVote) Type() string  { return MsgTypeSideVote }

// Implements Msg.
func (msg MsgSideChainVote) ValidateBasic() sdk.Error {
	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > sidechain.MaxSideChainIdLength {
		return ErrInvalidSideChainId(DefaultCodespace, msg.SideChainId)
	}
	return msg.MsgVote.ValidateBasic()
}

func (msg MsgSideChainVote) String() string {
	return fmt.Sprintf("MsgSideChainVote{%v - %s, %s}", msg.ProposalID, msg.Option, msg.SideChainId)
}

// Implements Msg.
func (msg MsgSideChainVote) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg. Identical to MsgVote, keep here for code readability.
func (msg MsgSideChainVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}

// Implements Msg. Identical to MsgVote, keep here for code readability.
func (msg MsgSideChainVote) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}
