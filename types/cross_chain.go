package types

import (
	"fmt"
	"math"
	"strconv"
)

type IbcChannelID uint8
type IbcChainID uint16

func ParseIbcChannelID(input string) (IbcChannelID, error) {
	channelID, err := strconv.Atoi(input)
	if err != nil {
		return IbcChannelID(0), err
	}
	if channelID > math.MaxInt8 || channelID < 0 {
		return IbcChannelID(0), fmt.Errorf("channelID must be in [0, 255]")
	}
	return IbcChannelID(channelID), nil
}

func ParseIbcChainID(input string) (IbcChainID, error) {
	chainID, err := strconv.Atoi(input)
	if err != nil {
		return IbcChainID(0), err
	}
	if chainID > math.MaxUint16 || chainID < 0 {
		return IbcChainID(0), fmt.Errorf("cross chainID must be in [0, 65535]")
	}
	return IbcChainID(chainID), nil
}
