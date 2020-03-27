package types

import (
	"fmt"
	"math"
	"strconv"
)

type CrossChainChannelID uint8
type CrossChainID uint16

func ParseCrossChainChannelID(input string) (CrossChainChannelID, error) {
	channelID, err := strconv.Atoi(input)
	if err != nil {
		return CrossChainChannelID(0), err
	}
	if channelID > math.MaxInt8 || channelID < 0 {
		return CrossChainChannelID(0), fmt.Errorf("channelID must be in [0, 255]")
	}
	return CrossChainChannelID(channelID), nil
}

func ParseCrossChainID(input string) (CrossChainID, error) {
	chainID, err := strconv.Atoi(input)
	if err != nil {
		return CrossChainID(0), err
	}
	if chainID > math.MaxUint16 || chainID < 0 {
		return CrossChainID(0), fmt.Errorf("cross chainID must be in [0, 65535]")
	}
	return CrossChainID(chainID), nil
}
