package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IBCEventType                 = "IBCPackage"
	IBCPackageInfoAttributeKey   = "IBCPackageInfo"
	IBCPackageInfoAttributeValue = "%s::%d::%d"
)

func BuildIBCPackageAttributeValue(sideChainID string, channelID sdk.IbcChannelID, sequence uint64) string {
	return fmt.Sprintf(IBCPackageInfoAttributeValue, sideChainID, channelID, sequence)
}
