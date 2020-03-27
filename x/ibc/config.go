package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type crossChainConfig struct {
	sourceChainID sdk.CrossChainID

	nameToChannelID map[string]sdk.CrossChainChannelID
	channelIDToName map[sdk.CrossChainChannelID]string

	destChainNameToID map[string]sdk.CrossChainID
	destChainIDToName map[sdk.CrossChainID]string
}

var crossChainCfg = newCrossChainCfg()

func newCrossChainCfg() *crossChainConfig {
	config := &crossChainConfig{
		sourceChainID:     0,
		nameToChannelID:   make(map[string]sdk.CrossChainChannelID),
		channelIDToName:   make(map[sdk.CrossChainChannelID]string),
		destChainNameToID: make(map[string]sdk.CrossChainID),
		destChainIDToName: make(map[sdk.CrossChainID]string),
	}
	return config
}

func RegisterCrossChainChannel(name string, id sdk.CrossChainChannelID) error {
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

func RegisterDestChainID(name string, id sdk.CrossChainID) error {
	_, ok := crossChainCfg.destChainNameToID[name]
	if ok {
		return fmt.Errorf("duplicated destination chain name")
	}
	_, ok = crossChainCfg.destChainIDToName[id]
	if ok {
		return fmt.Errorf("duplicated destination chain id")
	}
	crossChainCfg.destChainNameToID[name] = id
	crossChainCfg.destChainIDToName[id] = name
	return nil
}

func GetChannelID(channelName string) (sdk.CrossChainChannelID, error) {
	id, ok := crossChainCfg.nameToChannelID[channelName]
	if !ok {
		return sdk.CrossChainChannelID(0), fmt.Errorf("non-existing channel")
	}
	return id, nil
}

func SetSourceChainID(sourceChainID sdk.CrossChainID) {
	crossChainCfg.sourceChainID = sourceChainID
}

func GetSourceChainID() sdk.CrossChainID {
	return crossChainCfg.sourceChainID
}

func GetDestChainID(name string) (sdk.CrossChainID, error) {
	destChainID, exist := crossChainCfg.destChainNameToID[name]
	if !exist {
		return sdk.CrossChainID(0), fmt.Errorf("non-existing destination chainID")
	}
	return destChainID, nil
}
