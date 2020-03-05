package ibc

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	PrefixForCrossChainPackageKey = []byte{0x00}
	PrefixForSequenceKey          = []byte{0x01}
)

func BuildIBCPackageKey(sourceChainID, destinationChainID sdk.CrossChainID, channelID sdk.ChannelID, sequence uint64) []byte {
	key := make([]byte, sourceChainIDLength+destChainIDLength+channelIDLength+sequenceLength)
	binary.BigEndian.PutUint16(key[:sourceChainIDLength], uint16(sourceChainID))
	binary.BigEndian.PutUint16(key[sourceChainIDLength:sourceChainIDLength+destChainIDLength], uint16(destinationChainID))
	copy(key[sourceChainIDLength+destChainIDLength:], []byte{byte(channelID)})

	sequenceBytes := make([]byte, sequenceLength)
	binary.BigEndian.PutUint64(sequenceBytes, sequence)
	copy(key[sourceChainIDLength+destChainIDLength+channelIDLength:], sequenceBytes)

	return append(PrefixForCrossChainPackageKey, key...)
}

func buildIBCPackageKeyPrefix(sourceChainID, destinationChainID sdk.CrossChainID, channelID sdk.ChannelID) []byte {
	key := make([]byte, sourceChainIDLength+destChainIDLength+channelIDLength)
	binary.BigEndian.PutUint16(key[:sourceChainIDLength], uint16(sourceChainID))
	binary.BigEndian.PutUint16(key[sourceChainIDLength:sourceChainIDLength+destChainIDLength], uint16(destinationChainID))
	copy(key[sourceChainIDLength+destChainIDLength:], []byte{byte(channelID)})

	return append(PrefixForCrossChainPackageKey, key...)
}

func buildChannelSequenceKey(channelID sdk.ChannelID) []byte {
	return append(PrefixForSequenceKey, byte(channelID))
}
