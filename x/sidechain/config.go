package sidechain

import sdk "github.com/cosmos/cosmos-sdk/types"

type crossChainConfig struct {
	srcChainID sdk.ChainID

	nameToChannelID map[string]sdk.ChannelID
	channelIDToName map[sdk.ChannelID]string
	channelIDToApp  map[sdk.ChannelID]sdk.CrossChainApplication

	destChainNameToID map[string]sdk.ChainID
	destChainIDToName map[sdk.ChainID]string
}

func newCrossChainCfg() *crossChainConfig {
	config := &crossChainConfig{
		srcChainID:        0,
		nameToChannelID:   make(map[string]sdk.ChannelID),
		channelIDToName:   make(map[sdk.ChannelID]string),
		destChainNameToID: make(map[string]sdk.ChainID),
		destChainIDToName: make(map[sdk.ChainID]string),
		channelIDToApp:    make(map[sdk.ChannelID]sdk.CrossChainApplication),
	}
	return config
}

func (c *crossChainConfig) DestChainNameToID(name string) sdk.ChainID {
	return c.destChainNameToID[name]
}

func (c *crossChainConfig) ChannelIDs() []sdk.ChannelID {
	IDs := make([]sdk.ChannelID, 0, len(c.channelIDToName))
	for _, id := range c.nameToChannelID {
		IDs = append(IDs, id)
	}
	return IDs
}
