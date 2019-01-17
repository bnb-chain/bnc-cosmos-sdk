package types

import (
	"sync"
)

// block level pool to avoid frequently call ctx.With harms performance according to our test
//
// NOTE: states keep in this pool should be cleared per-block,
// an appropriate place should be in the end of Commit() with
// deliver state
type Pool struct {
	accounts sync.Map // save tx/gov related addresses (string wrapped bytes) to be published

	mesMux sync.Mutex
	// We choose slice instead of map to store msg. Since Msg can be duplicated and is not hashable,
	// and hash with msg.GetSignBytes cost a lot.
	messages []Msg
}

func (p *Pool) AddMsgs(msgs []Msg) {
	p.mesMux.Lock()
	defer p.mesMux.Unlock()
	for _, m := range msgs {
		p.messages = append(p.messages, m)
	}
}

func (p Pool) InterestMsgs(choose func(Msg) bool) []Msg {
	p.mesMux.Lock()
	defer p.mesMux.Unlock()
	msgs := make([]Msg, 0, 0)
	if p.messages == nil {
		return msgs
	}
	for _, v := range p.messages {
		if choose(v) {
			msgs = append(msgs, v)
		}
	}
	return msgs
}

func (p *Pool) AddAddrs(addrs []AccAddress) {
	for _, addr := range addrs {
		p.accounts.Store(string(addr.Bytes()), struct{}{})
	}
}

func (p Pool) TxRelatedAddrs() []string {
	addrs := make([]string, 0, 0)
	p.accounts.Range(func(key, value interface{}) bool {
		addrs = append(addrs, key.(string))
		return true
	})
	return addrs
}

func (p *Pool) Clear() {
	p.accounts = sync.Map{}

	p.mesMux.Lock()
	defer p.mesMux.Unlock()
	p.messages = []Msg{}
}
