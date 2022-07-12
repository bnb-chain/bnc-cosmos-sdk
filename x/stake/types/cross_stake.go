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

	CrossStakeTransferOutRewardRelayFee      = "crossStakeTransferOutRewardRelayFee"
	CrossStakeTransferOutUndelegatedRelayFee = "crossStakeTransferOutUndelegatedRelayFee"

	CrossStakeTypeDelegate               CrossStakePackageType = 1
	CrossStakeTypeUndelegate             CrossStakePackageType = 2
	CrossStakeTypeRedelegate             CrossStakePackageType = 3
	CrossStakeTypeTransferOutReward      CrossStakePackageType = 4
	CrossStakeTypeTransferOutUndelegated CrossStakePackageType = 5

	CrossStakeTopic = pubsub.Topic("cross-stake")

	CrossStakeDelegateType               string = "CSD"
	CrossStakeUndelegateType             string = "CSU"
	CrossStakeTransferOutRewardType      string = "CSTR"
	CrossStakeTransferOutUndelegatedType string = "CSTU"
	CrossStakeRedelegateType             string = "CSRD"
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

type TransferOutRewardEvent struct {
	ChainId       string
	Type          string
	Delegators    []sdk.AccAddress
	Receivers     []sdk.SmartChainAddress
	Amounts       []int64
	BSCRelayerFee int64
}

func (event TransferOutRewardEvent) GetTopic() pubsub.Topic {
	return CrossStakeTopic
}

type TransferOutUndelegatedEvent struct {
	ChainId       string
	Type          string
	Delegator     sdk.AccAddress
	Validator     sdk.ValAddress
	Receiver      sdk.SmartChainAddress
	Amount        int64
	BSCRelayerFee int64
}

func (event TransferOutUndelegatedEvent) GetTopic() pubsub.Topic {
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

type CrossStakeTransferOutRewardSynPackage struct {
	EventCode   CrossStakePackageType
	Amounts     []*big.Int
	Recipients  []sdk.SmartChainAddress
	RefundAddrs []sdk.AccAddress
}

type CrossStakeTransferOutUndelegatedSynPackage struct {
	EventCode  CrossStakePackageType
	Amount     *big.Int
	Recipient  sdk.SmartChainAddress
	RefundAddr sdk.AccAddress
	Validator  sdk.ValAddress
}

func GetStakeCAoB(sourceAddr []byte, salt string) sdk.AccAddress {
	saltBytes := []byte("Staking" + salt + "Address Anchor")
	return sdk.XOR(tmhash.SumTruncated(saltBytes), sourceAddr)
}
