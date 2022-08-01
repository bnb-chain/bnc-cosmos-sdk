package types

import (
	"math/big"

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

	CrossDistributeRewardRelayFee      = "crossDistributeRewardRelayFee"
	CrossDistributeUndelegatedRelayFee = "crossDistributeUndelegatedRelayFee"

	CrossStakeTypeDelegate              CrossStakePackageType = 1
	CrossStakeTypeUndelegate            CrossStakePackageType = 2
	CrossStakeTypeRedelegate            CrossStakePackageType = 3
	CrossStakeTypeDistributeReward      CrossStakePackageType = 4
	CrossStakeTypeDistributeUndelegated CrossStakePackageType = 5

	CrossStakeDelegateType                           string = "CSD"
	CrossStakeDistributeRewardType                   string = "CSDR"
	CrossStakeDistributeUndelegatedType              string = "CSDU"
	CrossStakeDistributeRewardFailAckRefundType      string = "CSDRFAR"
	CrossStakeDistributeUndelegatedFailAckRefundType string = "CSDUFAR"

	DelegateCAoBSalt string = "Delegate"
	RewardCAoBSalt   string = "Reward"

	MinRewardThreshold int64 = 1e8
)

type CrossStakeDelegateSynPackage struct {
	DelAddr   sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
}

type CrossStakeDelegationAckPackage struct {
	PackageType CrossStakePackageType
	DelAddr     sdk.SmartChainAddress
	Validator   sdk.ValAddress
	Amount      *big.Int
	ErrorCode   uint8
}

func NewCrossStakeDelegationAckPackage(synPack *CrossStakeDelegateSynPackage, packageType CrossStakePackageType, errCode uint8) *CrossStakeDelegationAckPackage {
	return &CrossStakeDelegationAckPackage{packageType, synPack.DelAddr, synPack.Validator, synPack.Amount, errCode}
}

type CrossStakeUndelegateSynPackage struct {
	DelAddr   sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
}

type CrossStakeUndelegateAckPackage struct {
	PackageType CrossStakePackageType
	DelAddr     sdk.SmartChainAddress
	Validator   sdk.ValAddress
	Amount      *big.Int
	ErrorCode   uint8
}

func NewCrossStakeUndelegateAckPackage(synPack *CrossStakeUndelegateSynPackage, packageType CrossStakePackageType, errCode uint8) *CrossStakeUndelegateAckPackage {
	return &CrossStakeUndelegateAckPackage{packageType, synPack.DelAddr, synPack.Validator, synPack.Amount, errCode}
}

type CrossStakeRedelegateSynPackage struct {
	DelAddr sdk.SmartChainAddress
	ValSrc  sdk.ValAddress
	ValDst  sdk.ValAddress
	Amount  *big.Int
}

type CrossStakeRedelegateAckPackage struct {
	PackageType CrossStakePackageType
	DelAddr     sdk.SmartChainAddress
	ValSrc      sdk.ValAddress
	ValDst      sdk.ValAddress
	Amount      *big.Int
	ErrorCode   uint8
}

func NewCrossStakeRedelegationAckPackage(synPack *CrossStakeRedelegateSynPackage, packageType CrossStakePackageType, errCode uint8) *CrossStakeRedelegateAckPackage {
	return &CrossStakeRedelegateAckPackage{packageType, synPack.DelAddr, synPack.ValSrc, synPack.ValDst, synPack.Amount, errCode}
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
