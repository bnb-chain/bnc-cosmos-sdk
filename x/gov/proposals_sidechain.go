package gov

//nolint
const (
	SideProposalTypeParametersChange      ProposalKind = 0x81
	SideProposalTypeCrossParametersChange ProposalKind = 0x82
)

// is defined SideProposalType?
func validSideProposalType(pt ProposalKind) bool {
	if pt == SideProposalTypeParametersChange ||
		pt == SideProposalTypeCrossParametersChange {
		return true
	}
	return false
}
