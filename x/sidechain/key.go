package sidechain

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	prefixLength         = 1
	destIbcChainIDLength = 2
	channelIDLength      = 1
	sequenceLength       = 8
	permissionLength     = 1
)

var (
	SideChainStorePrefixByIdKey = []byte{0x01} // prefix for each key to a side chain store prefix, by side chain id

	PrefixForSendSequenceKey    = []byte{0xf0}
	PrefixForReceiveSequenceKey = []byte{0xf1}

	PrefixForChannelPermissionKey = []byte{0xc0}
)

func GetSideChainStorePrefixKey(sideChainId string) []byte {
	return append(SideChainStorePrefixByIdKey, []byte(sideChainId)...)
}

func buildChannelSequenceKey(destIbcChainID sdk.IbcChainID, channelID sdk.IbcChannelID, prefix []byte) []byte {
	key := make([]byte, prefixLength+destIbcChainIDLength+channelIDLength)

	copy(key[:prefixLength], prefix)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+destIbcChainIDLength], uint16(destIbcChainID))
	copy(key[prefixLength+destIbcChainIDLength:], []byte{byte(channelID)})
	return key
}

func buildChannelPermissionKey(destIbcChainID sdk.IbcChainID, channelID sdk.IbcChannelID) []byte {
	key := make([]byte, prefixLength+destIbcChainIDLength+channelIDLength)

	copy(key[:prefixLength], PrefixForChannelPermissionKey)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+destIbcChainIDLength], uint16(destIbcChainID))
	copy(key[prefixLength+destIbcChainIDLength:], []byte{byte(channelID)})
	return key
}

func buildChannelPermissionsPrefixKey(destIbcChainID sdk.IbcChainID) []byte {
	key := make([]byte, prefixLength+destIbcChainIDLength)

	copy(key[:prefixLength], PrefixForChannelPermissionKey)
	binary.BigEndian.PutUint16(key[prefixLength:prefixLength+destIbcChainIDLength], uint16(destIbcChainID))
	return key
}
