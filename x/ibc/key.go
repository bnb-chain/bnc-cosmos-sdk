package ibc

import (
	"encoding/binary"
)

var (
	PrefixForCrossChainPackageKey = []byte{0x01}
	KeyForBindChannelSequence     = []byte{0x02}
	KeyForTransferChannelSequence = []byte{0x03}
	KeyForTimeoutChannelSequence  = []byte{0x04}
	KeyForStakingChannelSequence  = []byte{0x05}
)

func BuildIBCPackageKey(sourceChainID, destinationChainID string, channelID ChannelID, sequence int64) []byte {
	key := make([]byte, 32+32+1+8)
	copy(key[:32], sourceChainID)
	copy(key[32:64], destinationChainID)
	key[64] = byte(channelID)
	sequenceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceBytes, uint64(sequence))
	copy(key[65:], sequenceBytes)
	return append(PrefixForCrossChainPackageKey, key...)
}

func buildIBCPackageKeyPrefix(sourceChainID, destinationChainID string, channelID int8) []byte {
	key := make([]byte, 32+32+1)
	copy(key[:32], sourceChainID)
	copy(key[32:64], destinationChainID)
	key[64] = byte(channelID)
	return append(PrefixForCrossChainPackageKey, key...)
}
