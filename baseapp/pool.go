package baseapp

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// block level pool to avoid frequently call ctx.With harms performance according to our test
//
// NOTE: states keep in this pool should be cleared per-block,
// an appropriate place should be in the end of Commit() with
// deliver state
type pool struct {
	accounts sync.Map // save tx related addresses (string wrapped bytes) to be published
}

func (p *pool) AddAddrs(addrs []sdk.AccAddress) {
	for _, addr := range addrs {
		p.accounts.Store(string(addr.Bytes()), struct{}{})
	}
}

func (p pool) TxRelatedAddrs() []string {
	addrs := make([]string, 0, 0)
	p.accounts.Range(func(key, value interface{}) bool {
		addrs = append(addrs, key.(string))
		return true
	})
	return addrs
}

func (p *pool) ClearTxRelatedAddrs() {
	p.accounts = sync.Map{}
}
