package types

import (
	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const Topic = pubsub.Topic("stake")

type SideDistributionEvent struct {
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

func (event SideDistributionEvent) GetTopic() pubsub.Topic {
	return Topic
}

type SideCompletedUBDEvent struct {
	CompUBDs    []UnbondingDelegation
	SideChainId string
}

func (event SideCompletedUBDEvent) GetTopic() pubsub.Topic {
	return Topic
}
