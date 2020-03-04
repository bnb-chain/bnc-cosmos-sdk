package store

import (
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//----------------------------------------

var _ KVStore = (*IavlStoreMock)(nil)
var _ CommitStore = (*IavlStoreMock)(nil)
var _ Queryable = (*IavlStoreMock)(nil)

// IavlStoreMock Implements KVStore and CommitStore.
type IavlStoreMock struct {
	db *dbm.MemDB
}

func LoadIAVLStoreMock() (CommitStore, error) {

	return &IavlStoreMock{db: dbm.NewMemDB()}, nil
}

func (st *IavlStoreMock) SetVersion(version int64) {
	return
}

// Implements Committer.
func (st *IavlStoreMock) Commit() CommitID {

	return CommitID{
		Version: 0,
		Hash:    nil,
	}
}

// Implements Committer.
func (st *IavlStoreMock) LastCommitID() CommitID {
	return CommitID{
		Version: 0,
		Hash:    nil,
	}
}

// Implements Committer.
func (st *IavlStoreMock) SetPruning(pruning sdk.PruningStrategy) {
	return
}

// VersionExists returns whether or not a given version is stored.
func (st *IavlStoreMock) VersionExists(version int64) bool {
	return false
}

// Implements Store.
func (st *IavlStoreMock) GetStoreType() StoreType {
	return sdk.StoreTypeIAVLMock
}

// Implements Store.
func (st *IavlStoreMock) CacheWrap() CacheWrap {
	return NewCacheKVStore(st)
}

// CacheWrapWithTrace implements the Store interface.
func (st *IavlStoreMock) CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap {
	return NewCacheKVStore(NewTraceKVStore(st, w, tc))
}

// Implements KVStore.
func (st *IavlStoreMock) Set(key, value []byte) {
	st.db.Set(key, value)
}

// Implements KVStore.
func (st *IavlStoreMock) Get(key []byte) (value []byte) {
	return st.db.Get(key)
}

// Implements KVStore.
func (st *IavlStoreMock) Has(key []byte) (exists bool) {
	return st.db.Has(key)
}

// Implements KVStore.
func (st *IavlStoreMock) Delete(key []byte) {
	st.db.Delete(key)
}

// Implements KVStore
func (st *IavlStoreMock) Prefix(prefix []byte) KVStore {
	return prefixStore{st, prefix}
}

// Implements KVStore.
func (st *IavlStoreMock) Iterator(start, end []byte) Iterator {
	return st.db.Iterator(start, end)
}

// Implements KVStore.
func (st *IavlStoreMock) ReverseIterator(start, end []byte) Iterator {
	return st.db.ReverseIterator(start, end)
}

// Query implements ABCI interface, allows queries
//
// by default we will return from (latest height -1),
// as we will have merkle proofs immediately (header height = data height + 1)
// If latest-1 is not present, use latest (which must be present)
// if you care to have the latest data to see a tx results, you must
// explicitly set the height you want to see
func (st *IavlStoreMock) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	return
}
