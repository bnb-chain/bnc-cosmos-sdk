package gov

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//-----------------------------------------------------------
// Proposal interface
type Proposal interface {
	GetProposalID() int64
	SetProposalID(int64)

	GetTitle() string
	SetTitle(string)

	GetDescription() string
	SetDescription(string)

	GetProposalType() ProposalKind
	SetProposalType(ProposalKind)

	GetStatus() ProposalStatus
	SetStatus(ProposalStatus)

	GetTallyResult() OldTallyResult
	SetTallyResult(OldTallyResult)

	GetNewTallyResult() NewTallyResult
	SetNewTallyResult(NewTallyResult)

	GetSubmitTime() time.Time
	SetSubmitTime(time.Time)

	GetTotalDeposit() sdk.Coins
	SetTotalDeposit(sdk.Coins)

	GetVotingStartTime() time.Time
	SetVotingStartTime(time.Time)

	GetVotingPeriod() time.Duration
	SetVotingPeriod(time.Duration)
}

// checks if two proposals are equal
func ProposalEqual(proposalA Proposal, proposalB Proposal) bool {
	if proposalA.GetProposalID() == proposalB.GetProposalID() &&
		proposalA.GetTitle() == proposalB.GetTitle() &&
		proposalA.GetDescription() == proposalB.GetDescription() &&
		proposalA.GetProposalType() == proposalB.GetProposalType() &&
		proposalA.GetStatus() == proposalB.GetStatus() &&
		proposalA.GetTallyResult().Equals(proposalB.GetTallyResult()) &&
		proposalA.GetSubmitTime().Equal(proposalB.GetSubmitTime()) &&
		proposalA.GetTotalDeposit().IsEqual(proposalB.GetTotalDeposit()) &&
		proposalA.GetVotingStartTime().Equal(proposalB.GetVotingStartTime()) &&
		proposalA.GetVotingPeriod() == proposalB.GetVotingPeriod() {
		return true
	}
	return false
}

//-----------------------------------------------------------
// Text Proposals
type TextProposal struct {
	ProposalID   int64         `json:"proposal_id"`   //  ID of the proposal
	Title        string        `json:"title"`         //  Title of the proposal
	Description  string        `json:"description"`   //  Description of the proposal
	ProposalType ProposalKind  `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	VotingPeriod time.Duration `json:"voting_period"` //  Length of the voting period

	Status      ProposalStatus `json:"proposal_status"` //  Status of the Proposal {Pending, Active, Passed, Rejected}
	TallyResult TallyResult    `json:"tally_result"`    //  Result of Tallys

	SubmitTime   time.Time `json:"submit_time"`   //  Height of the block where TxGovSubmitProposal was included
	TotalDeposit sdk.Coins `json:"total_deposit"` //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
}

type OldTextProposal struct {
	ProposalID   int64        `json:"proposal_id"`   //  ID of the proposal
	Title        string       `json:"title"`         //  Title of the proposal
	Description  string       `json:"description"`   //  Description of the proposal
	ProposalType ProposalKind `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}

	Status      ProposalStatus `json:"proposal_status"` //  Status of the Proposal {Pending, Active, Passed, Rejected}
	TallyResult OldTallyResult `json:"tally_result"`    //  Result of Tallys

	SubmitTime   time.Time `json:"submit_time"`   //  Height of the block where TxGovSubmitProposal was included
	TotalDeposit sdk.Coins `json:"total_deposit"` //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
}

type NewTextProposal struct {
	OldTextProposal

	VotingPeriod time.Duration  `json:"voting_period"`    //  Length of the voting period
	TallyResult  NewTallyResult `json:"new_tally_result"` //  Length of the voting period
}

// Implements Proposal Interface
var _ Proposal = (*OldTextProposal)(nil)
var _ Proposal = (*NewTextProposal)(nil)
var _ Proposal = (*TextProposal)(nil)

// nolint
func (tp TextProposal) GetProposalID() int64               { return tp.ProposalID }
func (tp *TextProposal) SetProposalID(proposalID int64)    { tp.ProposalID = proposalID }
func (tp TextProposal) GetTitle() string                   { return tp.Title }
func (tp *TextProposal) SetTitle(title string)             { tp.Title = title }
func (tp TextProposal) GetDescription() string             { return tp.Description }
func (tp *TextProposal) SetDescription(description string) { tp.Description = description }
func (tp TextProposal) GetProposalType() ProposalKind      { return tp.ProposalType }
func (tp *TextProposal) SetProposalType(proposalType ProposalKind) {
	tp.ProposalType = proposalType
}

func (tp TextProposal) GetStatus() ProposalStatus        { return tp.Status }
func (tp *TextProposal) SetStatus(status ProposalStatus) { tp.Status = status }
func (tp TextProposal) GetTallyResult() OldTallyResult {
	return OldTallyResult{
		tp.TallyResult.Yes,
		tp.TallyResult.Abstain,
		tp.TallyResult.No,
		tp.TallyResult.NoWithVeto,
	}
}
func (tp *TextProposal) SetTallyResult(tallyResult OldTallyResult) {
	tp.TallyResult = TallyResult{
		tallyResult.Yes,
		tallyResult.Abstain,
		tallyResult.No,
		tallyResult.NoWithVeto,
		sdk.ZeroDec(),
	}
}

func (tp TextProposal) GetSubmitTime() time.Time                { return tp.SubmitTime }
func (tp *TextProposal) SetSubmitTime(submitTime time.Time)     { tp.SubmitTime = submitTime }
func (tp TextProposal) GetTotalDeposit() sdk.Coins              { return tp.TotalDeposit }
func (tp *TextProposal) SetTotalDeposit(totalDeposit sdk.Coins) { tp.TotalDeposit = totalDeposit }
func (tp TextProposal) GetVotingStartTime() time.Time           { return tp.VotingStartTime }
func (tp *TextProposal) SetVotingStartTime(votingStartTime time.Time) {
	tp.VotingStartTime = votingStartTime
}
func (tp TextProposal) GetNewTallyResult() NewTallyResult {
	return NewTallyResult{
		tp.TallyResult.Yes,
		tp.TallyResult.Abstain,
		tp.TallyResult.No,
		tp.TallyResult.NoWithVeto,
		tp.TallyResult.Total,
	}
}
func (tp *TextProposal) SetNewTallyResult(tallyResult NewTallyResult) {
	tp.TallyResult = TallyResult{
		tallyResult.Yes,
		tallyResult.Abstain,
		tallyResult.No,
		tallyResult.NoWithVeto,
		tallyResult.Total,
	}
}

func (tp TextProposal) GetVotingPeriod() time.Duration {
	return tp.VotingPeriod
}
func (tp *TextProposal) SetVotingPeriod(votingPeriod time.Duration) {
	tp.VotingPeriod = votingPeriod
}

func (tp OldTextProposal) GetProposalID() int64               { return tp.ProposalID }
func (tp *OldTextProposal) SetProposalID(proposalID int64)    { tp.ProposalID = proposalID }
func (tp OldTextProposal) GetTitle() string                   { return tp.Title }
func (tp *OldTextProposal) SetTitle(title string)             { tp.Title = title }
func (tp OldTextProposal) GetDescription() string             { return tp.Description }
func (tp *OldTextProposal) SetDescription(description string) { tp.Description = description }
func (tp OldTextProposal) GetProposalType() ProposalKind      { return tp.ProposalType }
func (tp *OldTextProposal) SetProposalType(proposalType ProposalKind) {
	tp.ProposalType = proposalType
}
func (tp OldTextProposal) GetStatus() ProposalStatus                  { return tp.Status }
func (tp *OldTextProposal) SetStatus(status ProposalStatus)           { tp.Status = status }
func (tp OldTextProposal) GetTallyResult() OldTallyResult             { return tp.TallyResult }
func (tp *OldTextProposal) SetTallyResult(tallyResult OldTallyResult) { tp.TallyResult = tallyResult }
func (tp OldTextProposal) GetSubmitTime() time.Time                   { return tp.SubmitTime }
func (tp *OldTextProposal) SetSubmitTime(submitTime time.Time)        { tp.SubmitTime = submitTime }
func (tp OldTextProposal) GetTotalDeposit() sdk.Coins                 { return tp.TotalDeposit }
func (tp *OldTextProposal) SetTotalDeposit(totalDeposit sdk.Coins)    { tp.TotalDeposit = totalDeposit }
func (tp OldTextProposal) GetVotingStartTime() time.Time              { return tp.VotingStartTime }
func (tp *OldTextProposal) SetVotingStartTime(votingStartTime time.Time) {
	tp.VotingStartTime = votingStartTime
}
func (tp OldTextProposal) GetNewTallyResult() NewTallyResult             { return NewTallyResult{} }
func (tp *OldTextProposal) SetNewTallyResult(tallyResult NewTallyResult) {}

func (tp OldTextProposal) GetVotingPeriod() time.Duration {
	return 0
}
func (tp *OldTextProposal) SetVotingPeriod(votingPeriod time.Duration) {
}

func (tp NewTextProposal) GetVotingPeriod() time.Duration              { return tp.VotingPeriod }
func (tp *NewTextProposal) SetVotingPeriod(votingPeriod time.Duration) { tp.VotingPeriod = votingPeriod }
func (tp NewTextProposal) GetTallyResult() OldTallyResult              { return tp.OldTextProposal.TallyResult }
func (tp *NewTextProposal) SetTallyResult(tallyResult OldTallyResult) {
	tp.OldTextProposal.TallyResult = tallyResult
}
func (tp NewTextProposal) GetNewTallyResult() NewTallyResult { return tp.TallyResult }
func (tp *NewTextProposal) SetNewTallyResult(tallyResult NewTallyResult) {
	tp.TallyResult = tallyResult
}

//-----------------------------------------------------------
// ProposalQueue
type ProposalQueue []int64

//-----------------------------------------------------------
// ProposalKind

// Type that represents Proposal Type as a byte
type ProposalKind byte

//nolint
const (
	ProposalTypeNil             ProposalKind = 0x00
	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
	ProposalTypeListTradingPair ProposalKind = 0x04
	// ProposalTypeFeeChange belongs to ProposalTypeParameterChange. We use this to make it easily to distinguish。
	ProposalTypeFeeChange       ProposalKind = 0x05
	ProposalTypeCreateValidator ProposalKind = 0x06
	ProposalTypeRemoveValidator ProposalKind = 0x07
)

// String to proposalType byte.  Returns ff if invalid.
func ProposalTypeFromString(str string) (ProposalKind, error) {
	switch str {
	case "Text":
		return ProposalTypeText, nil
	case "ParameterChange":
		return ProposalTypeParameterChange, nil
	case "SoftwareUpgrade":
		return ProposalTypeSoftwareUpgrade, nil
	case "ListTradingPair":
		return ProposalTypeListTradingPair, nil
	case "FeeChange":
		return ProposalTypeFeeChange, nil
	case "CreateValidator":
		return ProposalTypeCreateValidator, nil
	case "RemoveValidator":
		return ProposalTypeRemoveValidator, nil
	default:
		return ProposalKind(0xff), errors.Errorf("'%s' is not a valid proposal type", str)
	}
}

// is defined ProposalType?
func validProposalType(pt ProposalKind) bool {
	if pt == ProposalTypeText ||
		pt == ProposalTypeParameterChange ||
		pt == ProposalTypeSoftwareUpgrade ||
		pt == ProposalTypeListTradingPair ||
		pt == ProposalTypeFeeChange ||
		pt == ProposalTypeCreateValidator ||
		pt == ProposalTypeRemoveValidator {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (pt ProposalKind) Marshal() ([]byte, error) {
	return []byte{byte(pt)}, nil
}

// Unmarshal needed for protobuf compatibility
func (pt *ProposalKind) Unmarshal(data []byte) error {
	*pt = ProposalKind(data[0])
	return nil
}

// Marshals to JSON using string
func (pt ProposalKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (pt *ProposalKind) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := ProposalTypeFromString(s)
	if err != nil {
		return err
	}
	*pt = bz2
	return nil
}

// Turns VoteOption byte to String
func (pt ProposalKind) String() string {
	switch pt {
	case ProposalTypeText:
		return "Text"
	case ProposalTypeParameterChange:
		return "ParameterChange"
	case ProposalTypeSoftwareUpgrade:
		return "SoftwareUpgrade"
	case ProposalTypeListTradingPair:
		return "ListTradingPair"
	case ProposalTypeFeeChange:
		return "FeeChange"
	case ProposalTypeCreateValidator:
		return "CreateValidator"
	case ProposalTypeRemoveValidator:
		return "RemoveValidator"
	default:
		return ""
	}
}

// For Printf / Sprintf, returns bech32 when using %s
// nolint: errcheck
func (pt ProposalKind) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(pt.String()))
	default:
		// TODO: Do this conversion more directly
		s.Write([]byte(fmt.Sprintf("%v", byte(pt))))
	}
}

//-----------------------------------------------------------
// ProposalStatus

// Type that represents Proposal Status as a byte
type ProposalStatus byte

//nolint
const (
	StatusNil           ProposalStatus = 0x00
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
)

// ProposalStatusToString turns a string into a ProposalStatus
func ProposalStatusFromString(str string) (ProposalStatus, error) {
	switch str {
	case "DepositPeriod":
		return StatusDepositPeriod, nil
	case "VotingPeriod":
		return StatusVotingPeriod, nil
	case "Passed":
		return StatusPassed, nil
	case "Rejected":
		return StatusRejected, nil
	case "":
		return StatusNil, nil
	default:
		return ProposalStatus(0xff), errors.Errorf("'%s' is not a valid proposal status", str)
	}
}

// is defined ProposalType?
func validProposalStatus(status ProposalStatus) bool {
	if status == StatusDepositPeriod ||
		status == StatusVotingPeriod ||
		status == StatusPassed ||
		status == StatusRejected {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (status ProposalStatus) Marshal() ([]byte, error) {
	return []byte{byte(status)}, nil
}

// Unmarshal needed for protobuf compatibility
func (status *ProposalStatus) Unmarshal(data []byte) error {
	*status = ProposalStatus(data[0])
	return nil
}

// Marshals to JSON using string
func (status ProposalStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (status *ProposalStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := ProposalStatusFromString(s)
	if err != nil {
		return err
	}
	*status = bz2
	return nil
}

// Turns VoteStatus byte to String
func (status ProposalStatus) String() string {
	switch status {
	case StatusDepositPeriod:
		return "DepositPeriod"
	case StatusVotingPeriod:
		return "VotingPeriod"
	case StatusPassed:
		return "Passed"
	case StatusRejected:
		return "Rejected"
	default:
		return ""
	}
}

// For Printf / Sprintf, returns bech32 when using %s
// nolint: errcheck
func (status ProposalStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(status.String()))
	default:
		// TODO: Do this conversion more directly
		s.Write([]byte(fmt.Sprintf("%v", byte(status))))
	}
}

//-----------------------------------------------------------
// Tally Results
type TallyResult struct {
	Yes        sdk.Dec `json:"yes"`
	Abstain    sdk.Dec `json:"abstain"`
	No         sdk.Dec `json:"no"`
	NoWithVeto sdk.Dec `json:"no_with_veto"`
	Total      sdk.Dec `json:"total"`
}

// checks if two proposals are equal
func EmptyTallyResult() TallyResult {
	return TallyResult{
		Yes:        sdk.ZeroDec(),
		Abstain:    sdk.ZeroDec(),
		No:         sdk.ZeroDec(),
		NoWithVeto: sdk.ZeroDec(),
		Total:      sdk.ZeroDec(),
	}
}

// checks if two proposals are equal
func (resultA TallyResult) Equals(resultB TallyResult) bool {
	return resultA.Yes.Equal(resultB.Yes) &&
		resultA.Abstain.Equal(resultB.Abstain) &&
		resultA.No.Equal(resultB.No) &&
		resultA.NoWithVeto.Equal(resultB.NoWithVeto) &&
		resultA.Total.Equal(resultB.Total)
}

type OldTallyResult struct {
	Yes        sdk.Dec `json:"yes"`
	Abstain    sdk.Dec `json:"abstain"`
	No         sdk.Dec `json:"no"`
	NoWithVeto sdk.Dec `json:"no_with_veto"`
}

// checks if two proposals are equal
func EmptyOldTallyResult() OldTallyResult {
	return OldTallyResult{
		Yes:        sdk.ZeroDec(),
		Abstain:    sdk.ZeroDec(),
		No:         sdk.ZeroDec(),
		NoWithVeto: sdk.ZeroDec(),
	}
}

// checks if two proposals are equal
func (resultA OldTallyResult) Equals(resultB OldTallyResult) bool {
	return resultA.Yes.Equal(resultB.Yes) &&
		resultA.Abstain.Equal(resultB.Abstain) &&
		resultA.No.Equal(resultB.No) &&
		resultA.NoWithVeto.Equal(resultB.NoWithVeto)
}

type NewTallyResult struct {
	Yes        sdk.Dec `json:"yes"`
	Abstain    sdk.Dec `json:"abstain"`
	No         sdk.Dec `json:"no"`
	NoWithVeto sdk.Dec `json:"no_with_veto"`
	Total      sdk.Dec `json:"total"`
}

// checks if two proposals are equal
func EmptyNewTallyResult() NewTallyResult {
	return NewTallyResult{
		Yes:        sdk.ZeroDec(),
		Abstain:    sdk.ZeroDec(),
		No:         sdk.ZeroDec(),
		NoWithVeto: sdk.ZeroDec(),
		Total:      sdk.ZeroDec(),
	}
}

// checks if two proposals are equal
func (resultA NewTallyResult) Equals(resultB NewTallyResult) bool {
	return resultA.Yes.Equal(resultB.Yes) &&
		resultA.Abstain.Equal(resultB.Abstain) &&
		resultA.No.Equal(resultB.No) &&
		resultA.NoWithVeto.Equal(resultB.NoWithVeto) &&
		resultA.Total.Equal(resultB.Total)
}
