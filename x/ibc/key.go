package ibc

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	prefixLength          = 1
	srcIbcChainIdLength   = 2
	destIbcChainIDLength  = 2
	channelIDLength       = 1
	sequenceLength        = 8
	totalPackageKeyLength = prefixLength + srcIbcChainIdLength + destIbcChainIDLength + channelIDLength + sequenceLength
)

var (
	PrefixForIbcPackageKey = []byte{0x00}
	PrefixForSequenceKey   = []byte{0x01}
)

func buildIBCPackageKey(srcIbcChainID, destIbcChainID sdk.IbcChainID, channelID sdk.IbcChannelID, sequence uint64) []byte {
	key := make([]byte, totalPackageKeyLength)

	copy(key[:prefixLength], PrefixForIbcPackageKey)
	binary.BigEndian.PutUint16(key[prefixLength:srcIbcChainIdLength+prefixLength], uint16(srcIbcChainID))
	binary.BigEndian.PutUint16(key[prefixLength+srcIbcChainIdLength:prefixLength+srcIbcChainIdLength+destIbcChainIDLength], uint16(destIbcChainID))
	copy(key[prefixLength+srcIbcChainIdLength+destIbcChainIDLength:], []byte{byte(channelID)})
	binary.BigEndian.PutUint64(key[prefixLength+srcIbcChainIdLength+destIbcChainIDLength+channelIDLength:], sequence)

	return key
}

func buildIBCPackageKeyPrefix(srcIbcChainID, destIbcChainID sdk.IbcChainID, channelID sdk.IbcChannelID) []byte {
	key := make([]byte, totalPackageKeyLength-sequenceLength)

	copy(key[:prefixLength], PrefixForIbcPackageKey)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+srcIbcChainIdLength], uint16(srcIbcChainID))
	binary.BigEndian.PutUint16(key[prefixLength+srcIbcChainIdLength:prefixLength+srcIbcChainIdLength+destIbcChainIDLength], uint16(destIbcChainID))
	copy(key[prefixLength+srcIbcChainIdLength+destIbcChainIDLength:], []byte{byte(channelID)})

	return key
}

func buildChannelSequenceKey(destIbcChainID sdk.IbcChainID, channelID sdk.IbcChannelID) []byte {
	key := make([]byte, prefixLength+destIbcChainIDLength+channelIDLength)

	copy(key[:prefixLength], PrefixForSequenceKey)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+destIbcChainIDLength], uint16(destIbcChainID))
	copy(key[prefixLength+destIbcChainIDLength:], []byte{byte(channelID)})

	return key
}
