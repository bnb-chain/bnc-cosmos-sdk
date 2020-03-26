package types

import (
	"fmt"
	"math"
	"strconv"
)

type ChannelID uint8
type CrossChainID uint16

type crossChainConfig struct {
	sourceChainID CrossChainID

	nameToChannelID map[string]ChannelID
	channelIDToName map[ChannelID]string

	destChainNameToID map[string]CrossChainID
	destChainIDToName map[CrossChainID]string
}

var crossChainCfg = newCrossChainCfg()

func newCrossChainCfg() *crossChainConfig {
	config := &crossChainConfig{
		sourceChainID:     0,
		nameToChannelID:   make(map[string]ChannelID),
		channelIDToName:   make(map[ChannelID]string),
		destChainNameToID: make(map[string]CrossChainID),
		destChainIDToName: make(map[CrossChainID]string),
	}
	return config
}

func RegisterCrossChainChannel(name string, id ChannelID) error {
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

func RegisterDestChainID(name string, id CrossChainID) error {
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

func GetChannelID(channelName string) (ChannelID, error) {
	id, ok := crossChainCfg.nameToChannelID[channelName]
	if !ok {
		return ChannelID(0), fmt.Errorf("non-existing channel")
	}
	return id, nil
}

func (channelID ChannelID) String() string {
	return crossChainCfg.channelIDToName[channelID]
}

func SetSourceChainID(sourceChainID CrossChainID) {
	crossChainCfg.sourceChainID = sourceChainID
}

func GetSourceChainID() CrossChainID {
	return crossChainCfg.sourceChainID
}

func GetDestChainID(name string) (CrossChainID, error) {
	destChainID, exist := crossChainCfg.destChainNameToID[name]
	if !exist {
		return CrossChainID(0), fmt.Errorf("non-existing destination chainID")
	}
	return destChainID, nil
}

func ParseCrossChainID(input string) (CrossChainID, error) {
	chainID, err := strconv.Atoi(input)
	if err != nil {
		return CrossChainID(0), err
	}
	if chainID > math.MaxUint16 || chainID < 0 {
		return CrossChainID(0), fmt.Errorf("cross chainID must be in [0, 1<<16 - 1]")
	}
	return CrossChainID(chainID), nil
}
