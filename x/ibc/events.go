package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	EventTypeSaveIBCChannelSettingFailed  = "save_ibc_channel_setting_failed"
	EventTypeSaveIBCChannelSettingSucceed = "save_ibc_channel_setting_succeed"

	AttributeKeySideChainId = "side_chain_id"
	AttributeKeyChannelId   = "channel_id"
	AttributeKeyError       = "error"
)

const (
	separator                    = "::"
	ibcEventType                 = "IBCPackage"
	ibcPackageInfoAttributeKey   = "IBCPackageInfo"
	ibcPackageInfoAttributeValue = "%d" + separator + "%d" + separator + "%d" // destChainID channelID sequence
)

func buildIBCPackageAttributeValue(sideChainID sdk.ChainID, channelID sdk.ChannelID, sequence uint64) string {
	return fmt.Sprintf(ibcPackageInfoAttributeValue, sideChainID, channelID, sequence)
}
