package types

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/pubsub"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type CrossStakePackageType uint8

const (
	CrossStakeChannel = "crossStake"

	CrossStakeChannelID sdk.ChannelID = 16

	TagCrossStakeChannel      = "CrossStakeChannel"
	TagCrossStakePackageType  = "CrossStakePackageType"
	TagCrossStakeSendSequence = "CrossStakeSendSequence"

	CrossStakeDistributeRewardRelayFee      = "crossStakeDistributeRewardRelayFee"
	CrossStakeDistributeUndelegatedRelayFee = "crossStakeDistributeUndelegatedRelayFee"

	CrossStakeTypeDelegate              CrossStakePackageType = 1
	CrossStakeTypeUndelegate            CrossStakePackageType = 2
	CrossStakeTypeRedelegate            CrossStakePackageType = 3
	CrossStakeTypeDistributeReward      CrossStakePackageType = 4
	CrossStakeTypeDistributeUndelegated CrossStakePackageType = 5

	CrossStakeTopic = pubsub.Topic("cross-stake")

	CrossStakeDelegateType              string = "CSD"
	CrossStakeUndelegateType            string = "CSU"
	CrossStakeDistributeRewardType      string = "CSDR"
	CrossStakeDistributeUndelegatedType string = "CSDU"
	CrossStakeRedelegateType            string = "CSRD"
)

type CrossStakeEvent struct {
	ChainId      string
	Type         string
	Delegator    sdk.AccAddress
	ValidatorSrc sdk.ValAddress
	ValidatorDst sdk.ValAddress
	RelayFee     int64
}

func (event CrossStakeEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type DistributeRewardEvent struct {
	ChainId       string
	Type          string
	Delegator     sdk.AccAddress
	Receiver      sdk.SmartChainAddress
	Amount        int64
	BSCRelayerFee int64
}

func (event DistributeRewardEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type DistributeUndelegatedEvent struct {
	ChainId       string
	Type          string
	Delegator     sdk.AccAddress
	Validator     sdk.ValAddress
	Receiver      sdk.SmartChainAddress
	Amount        int64
	BSCRelayerFee int64
}

func (event DistributeUndelegatedEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type RewardRefundEvent struct {
	RefundAddr sdk.AccAddress
	Amount     int64
	Recipient  sdk.SmartChainAddress
}

func (event RewardRefundEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type UndelegatedRefundEvent struct {
	RefundAddr sdk.AccAddress
	Amount     int64
	Recipient  sdk.SmartChainAddress
}

func (event UndelegatedRefundEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type CrossStakeDelegateSynPackage struct {
	PackageType CrossStakePackageType
	DelAddr     sdk.SmartChainAddress
	Validator   sdk.ValAddress
	Amount      *big.Int
}

type CrossStakeDelegationAckPackage struct {
	CrossStakeDelegateSynPackage
	ErrorCode uint8
}

type CrossStakeUndelegateSynPackage struct {
	PackageType CrossStakePackageType
	DelAddr     sdk.SmartChainAddress
	Validator   sdk.ValAddress
	Amount      *big.Int
}

type CrossStakeUndelegateAckPackage struct {
	CrossStakeUndelegateSynPackage
	ErrorCode uint8
}

type CrossStakeRedelegateSynPackage struct {
	PackageType CrossStakePackageType
	DelAddr     sdk.SmartChainAddress
	ValSrc      sdk.ValAddress
	ValDst      sdk.ValAddress
	Amount      *big.Int
}

type CrossStakeRedelegateAckPackage struct {
	CrossStakeRedelegateSynPackage
	ErrorCode uint8
}

type CrossStakeDistributeRewardSynPackage struct {
	EventCode CrossStakePackageType
	Amount    *big.Int
	Recipient sdk.SmartChainAddress
}

type CrossStakeDistributeUndelegatedSynPackage struct {
	EventCode CrossStakePackageType
	Amount    *big.Int
	Recipient sdk.SmartChainAddress
	Validator sdk.ValAddress
}

func GetStakeCAoB(sourceAddr []byte, salt string) sdk.AccAddress {
	saltBytes := []byte("Staking" + salt + "Address Anchor")
	return sdk.XOR(tmhash.SumTruncated(saltBytes), sourceAddr)
}
