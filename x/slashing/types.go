package slashing

import "github.com/cosmos/cosmos-sdk/types"

type AddressType int

const (
	SideConsAddrType AddressType = iota + 1
	SideVoteAddrType
)

type SideSlashPackage struct {
	SideAddr      []byte        `json:"side_addr"`
	SideHeight    uint64        `json:"side_height"`
	SideChainId   types.ChainID `json:"side_chain_id"`
	SideTimestamp uint64        `json:"side_timestamp"`
	addrType      AddressType   `rlp:"-"`
}
