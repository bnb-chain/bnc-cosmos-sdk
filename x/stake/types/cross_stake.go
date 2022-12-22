package types

import (
	"fmt"
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

	TransferInType  string = "TI"
	TransferOutType string = "TO"

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
	Recipient sdk.SmartChainAddress
	Amount    *big.Int
	ErrorCode RefundError
}

func GetStakeCAoB(sourceAddr []byte, salt string) sdk.AccAddress {
	saltBytes := []byte("Staking" + salt + "Address Anchor")
	return sdk.XOR(tmhash.SumTruncated(saltBytes), sourceAddr)
}

// ----------------------------------------------------------------------------
// Client Types

type CrossStakeInfoResponse struct {
	Reward            int64          `json:"reward"`
	DelegationAddress sdk.AccAddress `json:"delegation_address"`
	RewardAddress     sdk.AccAddress `json:"reward_address"`
}

// NewCrossStakeInfoResponse creates a new CrossStakeInfoResponse instance
func NewCrossStakeInfoResponse(
	delAddr sdk.AccAddress, rewardAddr sdk.AccAddress, reward int64,
) CrossStakeInfoResponse {
	return CrossStakeInfoResponse{
		reward,
		delAddr,
		rewardAddr,
	}
}

func (dr CrossStakeInfoResponse) HumanReadableString() (string, error) {
	resp := "Delegation \n"
	resp += fmt.Sprintf("Reward: %d\n", dr.Reward)
	resp += fmt.Sprintf("Delegation address: %s\n", dr.DelegationAddress.String())
	resp += fmt.Sprintf("Reward address: %s", dr.RewardAddress.String())

	return resp, nil
}
