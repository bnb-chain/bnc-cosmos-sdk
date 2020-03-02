package ibc

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	msgCdc *codec.Codec
)

func init() {
	msgCdc = codec.New()
}

type ChannelID int8

const (
	BindChannelID     ChannelID = 0x01
	TransferChannelID ChannelID = 0x02
	TimeoutChannelID  ChannelID = 0x03
	StakingChannelID  ChannelID = 0x04
)

func NameToChannelID(channelName string) (ChannelID, error) {
	switch channelName {
	case "bind", "Bind":
		return BindChannelID, nil
	case "transfer", "Transfer":
		return TransferChannelID, nil
	case "timeout", "Timeout":
		return TimeoutChannelID, nil
	case "staking", "Staking":
		return StakingChannelID, nil
	default:
		return 0, fmt.Errorf("unsupported channel: %s", channelName)
	}
}

func (channelID ChannelID) String() string {
	switch channelID {
	case BindChannelID:
		return "bind"
	case TransferChannelID:
		return "transfer"
	case TimeoutChannelID:
		return "timeout"
	case StakingChannelID:
		return "Staking"
	default:
		return ""
	}
}
