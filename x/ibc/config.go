package ibc

import (
	"sort"

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

func (cfg *crossChainConfig) getChannelIDList() []sdk.IbcChannelID {
	channelList := make([]sdk.IbcChannelID, 0, len(cfg.channelIDToName))
	for id, _ := range cfg.channelIDToName {
		channelList = append(channelList, id)
	}
	sort.Slice(channelList, func(i, j int) bool {
		return i < j
	})
	return channelList
}

func (cfg *crossChainConfig) getChainIDList() []sdk.IbcChainID {
	chainIDList := make([]sdk.IbcChainID, 0, len(cfg.channelIDToName))
	for id, _ := range cfg.destChainIDToName {
		chainIDList = append(chainIDList, id)
	}
	sort.Slice(chainIDList, func(i, j int) bool {
		return i < j
	})
	return chainIDList
}
