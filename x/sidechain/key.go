package sidechain

var (
	SideChainStorePrefixByIdKey = []byte{0x01} // prefix for each key to a side chain store prefix, by side chain id
)

func GetSideChainStorePrefixKey(sideChainId string) []byte {
	return append(SideChainStorePrefixByIdKey, []byte(sideChainId)...)
}
