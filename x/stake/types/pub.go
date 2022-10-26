package types

import (
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	Topic = pubsub.Topic("stake")
)

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
	Operator sdk.ValAddress
	ChainId  string
}

// delegation update
type DelegationUpdateEvent struct {
	StakeEvent
	Delegation Delegation
	ChainId    string
}

// delegation removed
type DelegationRemovedEvent struct {
	StakeEvent
	DvPair  DVPair
	ChainId string
}

// UBDs update
type UBDUpdateEvent struct {
	StakeEvent
	UBD     UnbondingDelegation
	ChainId string
}

// RED update
type REDUpdateEvent struct {
	StakeEvent
	RED     Redelegation
	ChainId string
}

// completed unBonding event
type CompletedUBDEvent struct {
	StakeEvent
	ChainId  string
	CompUBDs []UnbondingDelegation
}

// completed reDelegation event
type CompletedREDEvent struct {
	StakeEvent
	ChainId  string
	CompREDs []DVVTriplet
}

// chain reward distribution event after BEP128
type DistributionEvent struct {
	StakeEvent
	ChainId string
	Data    []DistributionData
}

type DistributionData struct {
	Validator      sdk.ValAddress
	SelfDelegator  sdk.AccAddress
	DistributeAddr sdk.AccAddress
	ValShares      sdk.Dec
	ValTokens      sdk.Dec
	TotalReward    sdk.Dec
	Commission     sdk.Dec
	Rewards        []Reward
}

// delegate event
type DelegateEvent struct {
	StakeEvent
	Delegator  sdk.AccAddress
	Validator  sdk.ValAddress
	Amount     int64
	Denom      string
	TxHash     string
	CrossStake bool
}

type ChainDelegateEvent struct {
	DelegateEvent
	ChainId string
}

// undelegate
type UndelegateEvent struct {
	StakeEvent
	Delegator sdk.AccAddress
	Validator sdk.ValAddress
	Amount    int64
	Denom     string
	TxHash    string
}

type ChainUndelegateEvent struct {
	UndelegateEvent
	ChainId string
}

// redelegate
type RedelegateEvent struct {
	StakeEvent
	Delegator    sdk.AccAddress
	SrcValidator sdk.ValAddress
	DstValidator sdk.ValAddress
	Amount       int64
	Denom        string
	TxHash       string
}

type ChainRedelegateEvent struct {
	RedelegateEvent
	ChainId string
}

type ElectedValidatorsEvent struct {
	StakeEvent
	Validators []Validator
	ChainId    string
}
