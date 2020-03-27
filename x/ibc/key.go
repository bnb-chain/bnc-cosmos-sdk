package ibc

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	prefixLength          = 1
	sourceChainIDLength   = 2
	destChainIDLength     = 2
	channelIDLength       = 1
	sequenceLength        = 8
	totalPackageKeyLength = prefixLength + sourceChainIDLength + destChainIDLength + channelIDLength + sequenceLength
)

var (
	PrefixForCrossChainPackageKey = []byte{0x00}
	PrefixForSequenceKey          = []byte{0x01}
)

func buildIBCPackageKey(sourceChainID, destinationChainID sdk.CrossChainID, channelID sdk.CrossChainChannelID, sequence uint64) []byte {
	key := make([]byte, totalPackageKeyLength)

	copy(key[:prefixLength], PrefixForCrossChainPackageKey)
	binary.BigEndian.PutUint16(key[prefixLength:sourceChainIDLength+prefixLength], uint16(sourceChainID))
	binary.BigEndian.PutUint16(key[prefixLength+sourceChainIDLength:prefixLength+sourceChainIDLength+destChainIDLength], uint16(destinationChainID))
	copy(key[prefixLength+sourceChainIDLength+destChainIDLength:], []byte{byte(channelID)})
	binary.BigEndian.PutUint64(key[prefixLength+sourceChainIDLength+destChainIDLength+channelIDLength:], sequence)

	return key
}

func buildIBCPackageKeyPrefix(sourceChainID, destinationChainID sdk.CrossChainID, channelID sdk.CrossChainChannelID) []byte {
	key := make([]byte, totalPackageKeyLength-sequenceLength)

	copy(key[:prefixLength], PrefixForCrossChainPackageKey)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+sourceChainIDLength], uint16(sourceChainID))
	binary.BigEndian.PutUint16(key[prefixLength+sourceChainIDLength:prefixLength+sourceChainIDLength+destChainIDLength], uint16(destinationChainID))
	copy(key[prefixLength+sourceChainIDLength+destChainIDLength:], []byte{byte(channelID)})

	return key
}

func buildChannelSequenceKey(destChainID sdk.CrossChainID, channelID sdk.CrossChainChannelID) []byte {
	key := make([]byte, prefixLength+destChainIDLength+channelIDLength)

	copy(key[:prefixLength], PrefixForSequenceKey)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+destChainIDLength], uint16(destChainID))
	copy(key[prefixLength+destChainIDLength:], []byte{byte(channelID)})

	return key
}
