package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type crossChainConfig struct {
	srcIbcChainID sdk.IbcChainID

	nameToChannelID map[string]sdk.IbcChannelID
	channelIDToName map[sdk.IbcChannelID]string

	destChainNameToID map[string]sdk.IbcChainID
	destChainIDToName map[sdk.IbcChainID]string
}

var crossChainCfg = newCrossChainCfg()

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

func RegisterChannel(name string, id sdk.IbcChannelID) error {
	_, ok := crossChainCfg.nameToChannelID[name]
	if ok {
		return fmt.Errorf("duplicated channel name")
	}
	_, ok = crossChainCfg.channelIDToName[id]
	if ok {
		return fmt.Errorf("duplicated channel id")
	}
	crossChainCfg.nameToChannelID[name] = id
	crossChainCfg.channelIDToName[id] = name
	return nil
}

// internally, we use name as the id of the chain, must be unique
func RegisterDestChain(name string, ibcChainID sdk.IbcChainID) error {
	_, ok := crossChainCfg.destChainNameToID[name]
	if ok {
		return fmt.Errorf("duplicated destination chain name")
	}
	_, ok = crossChainCfg.destChainIDToName[ibcChainID]
	if ok {
		return fmt.Errorf("duplicated destination chain ibcChainID")
	}
	crossChainCfg.destChainNameToID[name] = ibcChainID
	crossChainCfg.destChainIDToName[ibcChainID] = name
	return nil
}

func GetChannelID(channelName string) (sdk.IbcChannelID, error) {
	id, ok := crossChainCfg.nameToChannelID[channelName]
	if !ok {
		return sdk.IbcChannelID(0), fmt.Errorf("non-existing channel")
	}
	return id, nil
}

func SetSrcIbcChainID(srcIbcChainID sdk.IbcChainID) {
	crossChainCfg.srcIbcChainID = srcIbcChainID
}

func GetSrcIbcChainID() sdk.IbcChainID {
	return crossChainCfg.srcIbcChainID
}

func GetDestIbcChainID(name string) (sdk.IbcChainID, error) {
	destChainID, exist := crossChainCfg.destChainNameToID[name]
	if !exist {
		return sdk.IbcChainID(0), fmt.Errorf("non-existing destination ibcChainID")
	}
	return destChainID, nil
}
