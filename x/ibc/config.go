package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type crossChainConfig struct {
	srcIbcChainID sdk.IbcChainID

	nameToChannelID map[string]sdk.IbcChannelID
	channelIDToName map[sdk.IbcChannelID]string

	destChainNameToID map[string]sdk.IbcChainID
	destChainIDToName map[sdk.IbcChainID]string
}

func newCrossChainCfg() *crossChainConfig {
	config := &crossChainConfig{
		srcIbcChainID:     0,
		nameToChannelID:   make(map[string]sdk.IbcChannelID),
		channelIDToName:   make(map[sdk.IbcChannelID]string),
		destChainNameToID: make(map[string]sdk.IbcChainID),
		destChainIDToName: make(map[sdk.IbcChainID]string),
	}
	return config
}
