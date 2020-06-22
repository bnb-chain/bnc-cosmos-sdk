package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type packageRecord struct {
	destChainID sdk.IbcChainID
	channelID   sdk.IbcChannelID
	sequence    uint64
}

type packageCollector struct {
	collectedPackages []packageRecord
}

func newPackageCollector() *packageCollector {
	return &packageCollector{
		collectedPackages: nil,
	}
}
