package store

import (
	"fmt"
	"io"
	"sync"

	"github.com/tendermint/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultIAVLCacheSize = 10000
)

// load the iavl store
func LoadIAVLStore(db dbm.DB, id CommitID, pruning sdk.PruningStrategy) (CommitStore, error) {
	tree := iavl.NewMutableTree(db, defaultIAVLCacheSize)
	_, err := tree.LoadVersion(id.Version)
	if err != nil {
		return nil, err
	}
	iavl := newIAVLStore(tree, sdk.PruneNothing{})
	iavl.SetPruning(pruning)
	return iavl, nil
}

//----------------------------------------

var _ KVStore = (*IavlStore)(nil)
var _ CommitStore = (*IavlStore)(nil)
var _ Queryable = (*IavlStore)(nil)

// IavlStore Implements KVStore and CommitStore.
type IavlStore struct {

	// The underlying tree.
	Tree *iavl.MutableTree

	// The strategy to prune historical versions
	pruningStrategy sdk.PruningStrategy
}

// CONTRACT: tree should be fully loaded.
// nolint: unparam
func newIAVLStore(tree *iavl.MutableTree, ps sdk.PruningStrategy) *IavlStore {
	st := &IavlStore{
		Tree:            tree,
		pruningStrategy: ps,
	}
	return st
}

// Implements Committer.
func (st *IavlStore) Commit() CommitID {
	// Save a new version.
	hash, version, err := st.Tree.SaveVersion()
	if err != nil {
		// TODO: Do we want to extend Commit to allow returning errors?
		panic(err)
	}

	// Release an old version of history, if not a sync waypoint.
	for v, _ := range st.tree.GetVersions() {
		if st.pruningStrategy.Prune(v, version) {
			st.Tree.DeleteVersion(v)
		}
	}

	return CommitID{
		Version: version,
		Hash:    hash,
	}
}

// Implements Committer.
func (st *IavlStore) CommitAt(version int64) CommitID {
	// Save a new version.
	hash, version, err := st.Tree.SaveVersionAt(version)
	if err != nil {
		// TODO: Do we want to extend Commit to allow returning errors?
		panic(err)
	}

	return CommitID{
		Version: version,
		Hash:    hash,
	}
}

// Implements Committer.
func (st *IavlStore) LastCommitID() CommitID {
	return CommitID{
		Version: st.Tree.Version(),
		Hash:    st.Tree.Hash(),
	}
}

// Implements Committer.
func (st *IavlStore) SetPruning(pruning sdk.PruningStrategy) {
	st.pruningStrategy = pruning
}

// VersionExists returns whether or not a given version is stored.
func (st *IavlStore) VersionExists(version int64) bool {
	return st.Tree.VersionExists(version)
}

// Implements Store.
func (st *IavlStore) GetStoreType() StoreType {
	return sdk.StoreTypeIAVL
}

// Implements Store.
func (st *IavlStore) CacheWrap() CacheWrap {
	return NewCacheKVStore(st)
}

// CacheWrapWithTrace implements the Store interface.
func (st *IavlStore) CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap {
	return NewCacheKVStore(NewTraceKVStore(st, w, tc))
}

// Implements KVStore.
func (st *IavlStore) Set(key, value []byte) {
	st.Tree.Set(key, value)
}

// Implements KVStore.
func (st *IavlStore) Get(key []byte) (value []byte) {
	_, v := st.Tree.Get(key)
	return v
}

// Implements KVStore.
func (st *IavlStore) Has(key []byte) (exists bool) {
	return st.Tree.Has(key)
}

// Implements KVStore.
func (st *IavlStore) Delete(key []byte) {
	st.Tree.Remove(key)
}

// Implements KVStore
func (st *IavlStore) Prefix(prefix []byte) KVStore {
	return prefixStore{st, prefix}
}

// Implements KVStore.
func (st *IavlStore) Iterator(start, end []byte) Iterator {
	return newIAVLIterator(st.Tree.ImmutableTree, start, end, true)
}

// Implements KVStore.
func (st *IavlStore) ReverseIterator(start, end []byte) Iterator {
	return newIAVLIterator(st.Tree.ImmutableTree, start, end, false)
}

// Handle gatest the latest height, if height is 0
func getHeight(tree *iavl.MutableTree, req abci.RequestQuery) int64 {
	height := req.Height
	if height == 0 {
		latest := tree.Version()
		if tree.VersionExists(latest - 1) {
			height = latest - 1
		} else {
			height = latest
		}
	}
	return height
}

// Query implements ABCI interface, allows queries
//
// by default we will return from (latest height -1),
// as we will have merkle proofs immediately (header height = data height + 1)
// If latest-1 is not present, use latest (which must be present)
// if you care to have the latest data to see a tx results, you must
// explicitly set the height you want to see
func (st *IavlStore) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(req.Data) == 0 {
		msg := "Query cannot be zero length"
		return sdk.ErrTxDecode(msg).QueryResult()
	}

	tree := st.Tree

	// store the height we chose in the response, with 0 being changed to the
	// latest height
	res.Height = getHeight(tree, req)

	switch req.Path {
	case "/store", "/key": // Get by key
		key := req.Data // Data holds the key bytes
		res.Key = key
		if !st.VersionExists(res.Height) {
			res.Log = cmn.ErrorWrap(iavl.ErrVersionDoesNotExist, "").Error()
			break
		}
		if req.Prove {
			value, proof, err := tree.GetVersionedWithProof(key, res.Height)
			if err != nil {
				res.Log = err.Error()
				break
			}
			res.Value = value
			res.Proof = &merkle.Proof{Ops: []merkle.ProofOp{iavl.NewIAVLValueOp(key, proof).ProofOp()}}
		} else {
			_, res.Value = tree.GetVersioned(key, res.Height)
		}
	case "/subspace":
		subspace := req.Data
		res.Key = subspace
		var KVs []KVPair
		iterator := sdk.KVStorePrefixIterator(st, subspace)
		for ; iterator.Valid(); iterator.Next() {
			KVs = append(KVs, KVPair{Key: iterator.Key(), Value: iterator.Value()})
		}
		iterator.Close()
		res.Value = cdc.MustMarshalBinaryLengthPrefixed(KVs)
	default:
		msg := fmt.Sprintf("Unexpected Query path: %v", req.Path)
		return sdk.ErrUnknownRequest(msg).QueryResult()
	}
	return
}

//----------------------------------------

// Implements Iterator.
type iavlIterator struct {
	// Underlying store
	tree *iavl.ImmutableTree

	// Domain
	start, end []byte

	// Iteration order
	ascending bool

	// Channel to push iteration values.
	iterCh chan cmn.KVPair

	// Close this to release goroutine.
	quitCh chan struct{}

	// Close this to signal that state is initialized.
	initCh chan struct{}

	//----------------------------------------
	// What follows are mutable state.
	mtx sync.Mutex

	invalid bool   // True once, true forever
	key     []byte // The current key
	value   []byte // The current value
}

var _ Iterator = (*iavlIterator)(nil)

// newIAVLIterator will create a new iavlIterator.
// CONTRACT: Caller must release the iavlIterator, as each one creates a new
// goroutine.
func newIAVLIterator(tree *iavl.ImmutableTree, start, end []byte, ascending bool) *iavlIterator {
	iter := &iavlIterator{
		tree:      tree,
		start:     cp(start),
		end:       cp(end),
		ascending: ascending,
		iterCh:    make(chan cmn.KVPair, 0), // Set capacity > 0?
		quitCh:    make(chan struct{}),
		initCh:    make(chan struct{}),
	}
	go iter.iterateRoutine()
	go iter.initRoutine()
	return iter
}

// Run this to funnel items from the tree to iterCh.
func (iter *iavlIterator) iterateRoutine() {
	iter.tree.IterateRange(
		iter.start, iter.end, iter.ascending,
		func(key, value []byte) bool {
			select {
			case <-iter.quitCh:
				return true // done with iteration.
			case iter.iterCh <- cmn.KVPair{Key: key, Value: value}:
				return false // yay.
			}
		},
	)
	close(iter.iterCh) // done.
}

// Run this to fetch the first item.
func (iter *iavlIterator) initRoutine() {
	iter.receiveNext()
	close(iter.initCh)
}

// Implements Iterator.
func (iter *iavlIterator) Domain() (start, end []byte) {
	return iter.start, iter.end
}

// Implements Iterator.
func (iter *iavlIterator) Valid() bool {
	iter.waitInit()
	iter.mtx.Lock()

	validity := !iter.invalid
	iter.mtx.Unlock()
	return validity
}

// Implements Iterator.
func (iter *iavlIterator) Next() {
	iter.waitInit()
	iter.mtx.Lock()
	iter.assertIsValid(true)

	iter.receiveNext()
	iter.mtx.Unlock()
}

// Implements Iterator.
func (iter *iavlIterator) Key() []byte {
	iter.waitInit()
	iter.mtx.Lock()
	iter.assertIsValid(true)

	key := iter.key
	iter.mtx.Unlock()
	return key
}

// Implements Iterator.
func (iter *iavlIterator) Value() []byte {
	iter.waitInit()
	iter.mtx.Lock()
	iter.assertIsValid(true)

	val := iter.value
	iter.mtx.Unlock()
	return val
}

// Implements Iterator.
func (iter *iavlIterator) Close() {
	close(iter.quitCh)
}

//----------------------------------------

func (iter *iavlIterator) setNext(key, value []byte) {
	iter.assertIsValid(false)

	iter.key = key
	iter.value = value
}

func (iter *iavlIterator) setInvalid() {
	iter.assertIsValid(false)

	iter.invalid = true
}

func (iter *iavlIterator) waitInit() {
	<-iter.initCh
}

func (iter *iavlIterator) receiveNext() {
	kvPair, ok := <-iter.iterCh
	if ok {
		iter.setNext(kvPair.Key, kvPair.Value)
	} else {
		iter.setInvalid()
	}
}

// assertIsValid panics if the iterator is invalid. If unlockMutex is true,
// it also unlocks the mutex before panicing, to prevent deadlocks in code that
// recovers from panics
func (iter *iavlIterator) assertIsValid(unlockMutex bool) {
	if iter.invalid {
		if unlockMutex {
			iter.mtx.Unlock()
		}
		panic("invalid iterator")
	}
}

//----------------------------------------

func cp(bz []byte) (ret []byte) {
	if bz == nil {
		return nil
	}
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}
