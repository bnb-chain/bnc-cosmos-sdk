package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	separator                    = "::"
	ibcEventType                 = "IBCPackage"
	ibcPackageInfoAttributeKey   = "IBCPackageInfo"
	ibcPackageInfoAttributeValue = "%s" + separator + "%d" + separator + "%d" + separator + "%d" //destChainName destChainID channelID sequence
)

func buildIBCPackageAttributeValue(sideChainName string, sideChainID sdk.IbcChainID, channelID sdk.IbcChannelID, sequence uint64) string {
	return fmt.Sprintf(ibcPackageInfoAttributeValue, sideChainName, sideChainID, channelID, sequence)
}
