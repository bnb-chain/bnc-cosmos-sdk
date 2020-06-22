package slashing

type SideDowntimeSlashEvent struct {
	SideConsAddr  []byte `json:"side_cons_addr"`
	SideHeight    uint64 `json:"side_height"`
	SideChainId   uint16 `json:"side_chain_id"`
	SideTimestamp uint64 `json:"side_timestamp"`
}
