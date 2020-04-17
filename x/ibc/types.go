package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PackageRecord struct {
	destChainName string
	destChainID   sdk.IbcChainID
	channelID     sdk.IbcChannelID
	sequence      uint64
}
