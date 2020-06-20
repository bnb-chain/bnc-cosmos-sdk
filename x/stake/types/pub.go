package types

import (
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const Topic = pubsub.Topic("stake")

type StakeEvent struct {
	IsFromTx bool
}

func (event StakeEvent) GetTopic() pubsub.Topic {
	return Topic
}

func (event StakeEvent) FromTx() bool {
	return event.IsFromTx
}

//----------------------------------------------------------------------------------------------------

// validator update event
type ValidatorUpdateEvent struct {
	StakeEvent
	Validator Validator
}

// validator removed event
type ValidatorRemovedEvent struct {
	StakeEvent
	Operator    sdk.ValAddress
	SideChainId string
}

// delegation update
type DelegationUpdateEvent struct {
	StakeEvent
	Delegation Delegation
}

// side delegation update
type SideDelegationUpdateEvent struct {
	DelegationUpdateEvent
	SideChainId string
}

// delegation removed
type DelegationRemovedEvent struct {
	StakeEvent
	DvPair DVPair
}

// side delegation removed
type SideDelegationRemovedEvent struct {
	DelegationRemovedEvent
	SideChainId string
}

// UBDs update
type UBDUpdateEvent struct {
	StakeEvent
	UBD UnbondingDelegation
}

// side UBD update
type SideUBDUpdateEvent struct {
	UBDUpdateEvent
	SideChainId string
}

// RED update
type REDUpdateEvent struct {
	StakeEvent
	RED Redelegation
}

// Side RED update
type SideREDUpdateEvent struct {
	REDUpdateEvent
	SideChainId string
}

// side completed unBonding event
type SideCompletedUBDEvent struct {
	StakeEvent
	SideChainId string
	CompUBDs    []UnbondingDelegation
}

// side completed reDelegation event
type SideCompletedREDEvent struct {
	StakeEvent
	SideChainId string
	CompREDs    []DVVTriplet
}

// side distribution event
type SideDistributionEvent struct {
	StakeEvent
	SideChainId string
	Data        []DistributionData
}

type DistributionData struct {
	Validator     sdk.ValAddress
	SelfDelegator sdk.AccAddress
	ValShares     sdk.Dec
	ValTokens     sdk.Dec
	TotalReward   sdk.Dec
	Commission    sdk.Dec
	Rewards       []Reward
}
