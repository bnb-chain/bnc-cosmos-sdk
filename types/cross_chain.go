package types

import (
	"fmt"
	"math"
	"strconv"
)

type ChannelID uint8
type CrossChainID uint16

type crossChainConfig struct {
	sourceChainID   CrossChainID
	nameToChannelID map[string]ChannelID
	channelIDToName map[ChannelID]string
}

var crossChainCfg = newCrossChainCfg()

func newCrossChainCfg() *crossChainConfig {
	config := &crossChainConfig{
		sourceChainID:   0,
		nameToChannelID: make(map[string]ChannelID),
		channelIDToName: make(map[ChannelID]string),
	}
	return config
}

func RegisterNewCrossChainChannel(name string, id ChannelID) error {
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

func GetChannelID(channelName string) (ChannelID, error) {
	id, ok := crossChainCfg.nameToChannelID[channelName]
	if !ok {
		return ChannelID(0), fmt.Errorf("non-existing channel")
	}
	return id, nil
}

func IsChannelRegistered(channelID ChannelID) bool {
	_, exist := crossChainCfg.channelIDToName[channelID]
	return exist;
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
