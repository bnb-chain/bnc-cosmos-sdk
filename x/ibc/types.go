package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ibcPackageRecord struct {
	destChainName string
	destChainID   sdk.IbcChainID
	channelID     sdk.IbcChannelID
	sequence      uint64
}
