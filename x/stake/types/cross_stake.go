package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type CrossStakeEventType uint8
type CrossStakeStatus uint8

const (
	CrossStakeChannel = "crossStake"

	CrossStakeChannelID sdk.ChannelID = 16

	TagCrossStakeChannel      = "CrossStakeChannel"
	TagCrossStakePackageType  = "CrossStakePackageType"
	TagCrossStakeSendSequence = "CrossStakeSendSequence"

	CrossDistributeRewardRelayFee      = "crossDistributeRewardRelayFee"
	CrossDistributeUndelegatedRelayFee = "crossDistributeUndelegatedRelayFee"

	CrossStakeFailed  CrossStakeStatus = 0
	CrossStakeSuccess CrossStakeStatus = 1

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

type CrossStakeAckPackage struct {
	Status    CrossStakeStatus
	ErrorCode uint8
	PackBytes []byte
}

type CrossStakeDelegateSynPackage struct {
	DelAddr   sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
}

type CrossStakeUndelegateSynPackage struct {
	DelAddr   sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
}

type CrossStakeRedelegateSynPackage struct {
	DelAddr sdk.SmartChainAddress
	ValSrc  sdk.ValAddress
	ValDst  sdk.ValAddress
	Amount  *big.Int
}

type CrossStakeDistributeRewardSynPackage struct {
	EventType CrossStakeEventType
	Recipient sdk.SmartChainAddress
	Amount    *big.Int
}

type CrossStakeDistributeUndelegatedSynPackage struct {
	EventType CrossStakeEventType
	Recipient sdk.SmartChainAddress
	Validator sdk.ValAddress
	Amount    *big.Int
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
