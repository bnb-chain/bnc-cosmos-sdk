package types

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/bsc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type CrossStakeEventType uint8

const (
	CrossStakeChannel = "crossStake"

	CrossStakeChannelID sdk.ChannelID = 16

	TagCrossStakeChannel      = "CrossStakeChannel"
	TagCrossStakePackageType  = "CrossStakePackageType"
	TagCrossStakeSendSequence = "CrossStakeSendSequence"

	CrossDistributeRewardRelayFee      = "crossDistributeRewardRelayFee"
	CrossDistributeUndelegatedRelayFee = "crossDistributeUndelegatedRelayFee"

	CrossStakeTypeDelegate              CrossStakeEventType = 1
	CrossStakeTypeUndelegate            CrossStakeEventType = 2
	CrossStakeTypeRedelegate            CrossStakeEventType = 3
	CrossStakeTypeDistributeReward      CrossStakeEventType = 4
	CrossStakeTypeDistributeUndelegated CrossStakeEventType = 5

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

type CrossStakeDelegateAckPackage struct {
	EventType CrossStakeEventType
	DelAddr   sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
	ErrorCode uint8
}

func NewCrossStakeDelegationAckPackage(synPack *CrossStakeDelegateSynPackage, eventType CrossStakeEventType, errCode uint8) *CrossStakeDelegateAckPackage {
	return &CrossStakeDelegateAckPackage{eventType, synPack.DelAddr, synPack.Validator, bsc.ConvertBCAmountToBSCAmount(synPack.Amount.Int64()), errCode}
}

type CrossStakeUndelegateSynPackage struct {
	DelAddr   sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
}

type CrossStakeUndelegateAckPackage struct {
	EventType CrossStakeEventType
	DelAddr   sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
	ErrorCode uint8
}

func NewCrossStakeUndelegateAckPackage(synPack *CrossStakeUndelegateSynPackage, eventType CrossStakeEventType, errCode uint8) *CrossStakeUndelegateAckPackage {
	return &CrossStakeUndelegateAckPackage{eventType, synPack.DelAddr, synPack.Validator, bsc.ConvertBCAmountToBSCAmount(synPack.Amount.Int64()), errCode}
}

type CrossStakeRedelegateSynPackage struct {
	DelAddr sdk.SmartChainAddress
	ValSrc  sdk.ValAddress
	ValDst  sdk.ValAddress
	Amount  *big.Int
}

type CrossStakeRedelegateAckPackage struct {
	EventType CrossStakeEventType
	DelAddr   sdk.SmartChainAddress
	ValSrc    sdk.ValAddress
	ValDst    sdk.ValAddress
	Amount    *big.Int
	ErrorCode uint8
}

func NewCrossStakeRedelegationAckPackage(synPack *CrossStakeRedelegateSynPackage, eventType CrossStakeEventType, errCode uint8) *CrossStakeRedelegateAckPackage {
	return &CrossStakeRedelegateAckPackage{eventType, synPack.DelAddr, synPack.ValSrc, synPack.ValDst, bsc.ConvertBCAmountToBSCAmount(synPack.Amount.Int64()), errCode}
}

type CrossStakeDistributeRewardSynPackage struct {
	EventType CrossStakeEventType
	Amount    *big.Int
	Recipient sdk.SmartChainAddress
}

type CrossStakeDistributeUndelegatedSynPackage struct {
	EventType CrossStakeEventType
	Amount    *big.Int
	Recipient sdk.SmartChainAddress
	Validator sdk.ValAddress
}

type RefundError uint32

const (
	DecodeFailed      RefundError = 100
	WithdrawBNBFailed RefundError = 101
)

type CrossStakeRefundPackage struct {
	EventType CrossStakeEventType
	Amount    *big.Int
	Recipient sdk.SmartChainAddress
	ErrorCode RefundError
}

func GetStakeCAoB(sourceAddr []byte, salt string) sdk.AccAddress {
	saltBytes := []byte("Staking" + salt + "Address Anchor")
	return sdk.XOR(tmhash.SumTruncated(saltBytes), sourceAddr)
}
