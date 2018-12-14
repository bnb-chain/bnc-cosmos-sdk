package concurrent

import (
	"encoding/hex"
	"sync"

	"github.com/cosmos/cosmos-sdk/server/concurrent/pool"
	abcicli "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/proxy"
)

var _ abcicli.Client = (*asyncLocalClient)(nil)

// asyncLocalClient is a variant from local_client.
// It makes ABCI calling more complex:
// 1. CheckTx/DeliverTx/Query/Info can be called concurrently
// 2. Other API would block calling CheckTx/DeliverTx/Query

const (
	WorkerPoolSize  = 16
	WorkerPoolSpawn = 4
	WorkerPoolQueue = 16
)

type WorkItem struct {
	reqRes *abcicli.ReqRes
	mtx    *sync.Mutex // make sure the eventual execution sequence
}

type asyncLocalClient struct {
	cmn.BaseService
	rwLock        *sync.RWMutex
	rwDeliverLock *sync.RWMutex
	Application   ApplicationCC
	abcicli.Callback

	checkTxPool    *pool.Pool
	deliverTxPool  *pool.Pool
	commitLock     *sync.Mutex
	checkTxLowLock *sync.Mutex
	checkTxMidLock *sync.Mutex
	wgCommit       *sync.WaitGroup

	checkTxQueue   chan WorkItem
	deliverTxQueue chan WorkItem
	log            log.Logger
}

func NewAsyncLocalClient(app types.Application, log log.Logger) *asyncLocalClient {
	appcc, ok := app.(ApplicationCC)
	if !ok {
		return nil
	}
	cli := &asyncLocalClient{
		rwLock:         new(sync.RWMutex),
		rwDeliverLock:  new(sync.RWMutex),
		Application:    appcc,
		checkTxPool:    pool.NewPool(WorkerPoolSize/2, WorkerPoolQueue/2, WorkerPoolSpawn/2),
		deliverTxPool:  pool.NewPool(WorkerPoolSize, WorkerPoolQueue, WorkerPoolSpawn),
		wgCommit:       new(sync.WaitGroup),
		checkTxQueue:   make(chan WorkItem, WorkerPoolQueue*2),
		deliverTxQueue: make(chan WorkItem, WorkerPoolQueue*2),
		commitLock:     new(sync.Mutex),
		checkTxLowLock: new(sync.Mutex),
		checkTxMidLock: new(sync.Mutex),
		log:            log,
	}
	cli.BaseService = *cmn.NewBaseService(nil, "asyncLocalClient", cli)
	return cli
}

func (app *asyncLocalClient) OnStart() error {
	if err := app.BaseService.OnStart(); err != nil {
		return err
	}
	go app.checkTxWorker()
	go app.deliverTxWorker()
	return nil
}

func (app *asyncLocalClient) OnStop() {
	app.BaseService.OnStop()
	app.commitLock.Lock()
	close(app.checkTxQueue)
	close(app.deliverTxQueue)
	app.commitLock.Unlock()
}

func (app *asyncLocalClient) SetResponseCallback(cb abcicli.Callback) {
	app.rwLock.Lock()
	app.rwDeliverLock.Lock()
	defer app.rwLock.Unlock()
	defer app.rwDeliverLock.Unlock()
	app.Callback = cb
}

func (app *asyncLocalClient) checkTxWorker() {
	for i := range app.checkTxQueue {
		i.mtx.Lock() // wait the PreCheckTx finish
		i.mtx.Unlock()

		if i.reqRes.Response == nil {
			tx := i.reqRes.Request.GetCheckTx().GetTx()
			app.rwLock.Lock() // make sure not other non-CheckTx/non-DeliverTx ABCI is called
			app.log.Debug("Handle Checktx", "Tx")
			res := app.Application.CheckTx(tx)
			app.rwLock.Unlock()                              // this unlock is put after wgCommit.Done() to give commit priority
			i.reqRes.Response = types.ToResponseCheckTx(res) // Set response
		}
		i.reqRes.Done()
		app.wgCommit.Done() // enable Commit to start
		app.Callback(i.reqRes.Request, i.reqRes.Response)
	}
}

func (app *asyncLocalClient) deliverTxWorker() {
	for i := range app.deliverTxQueue {
		i.mtx.Lock() // wait the PreCheckTx finish
		i.mtx.Unlock()
		//app.rwDeliverLock.Lock() // make sure not other non-CheckTx/non-DeliverTx ABCI is called
		if i.reqRes.Response == nil {
			tx := i.reqRes.Request.GetDeliverTx().GetTx()
			app.rwLock.Lock() // make sure not other non-CheckTx/non-DeliverTx ABCI is called
			app.log.Debug("Handle DeliverTx", "Tx", hex.EncodeToString(tx[:7]))
			res := app.Application.DeliverTx(tx)
			app.rwLock.Unlock()                                // this unlock is put after wgCommit.Done() to give commit priority
			i.reqRes.Response = types.ToResponseDeliverTx(res) // Set response
		}
		i.reqRes.Done()
		app.wgCommit.Done() // enable Commit to start
		//app.rwDeliverLock.Unlock() // this unlock is put after wgCommit.Done() to give commit priority
		app.Callback(i.reqRes.Request, i.reqRes.Response)
	}
}

// TODO: change types.Application to include Error()?
func (app *asyncLocalClient) Error() error {
	return nil
}

func (app *asyncLocalClient) FlushAsync() *abcicli.ReqRes {
	// Do nothing
	return newLocalReqRes(types.ToRequestFlush(), nil)
}

func (app *asyncLocalClient) EchoAsync(msg string) *abcicli.ReqRes {
	return app.callback(
		types.ToRequestEcho(msg),
		types.ToResponseEcho(msg),
	)
}

func (app *asyncLocalClient) InfoAsync(req types.RequestInfo) *abcicli.ReqRes {
	app.rwLock.RLock()
	app.rwDeliverLock.RLock()
	res := app.Application.Info(req)
	app.rwDeliverLock.RUnlock()
	app.rwLock.RUnlock()
	return app.callback(
		types.ToRequestInfo(req),
		types.ToResponseInfo(res),
	)
}

func (app *asyncLocalClient) SetOptionAsync(req types.RequestSetOption) *abcicli.ReqRes {
	app.rwLock.Lock()
	app.rwDeliverLock.Lock()
	res := app.Application.SetOption(req)
	app.rwDeliverLock.Unlock()
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestSetOption(req),
		types.ToResponseSetOption(res),
	)
}

func (app *asyncLocalClient) DeliverTxAsync(tx []byte) *abcicli.ReqRes {
	// no app level lock because the real DeliverTx would be called in the worker routine
	reqp := types.ToRequestDeliverTx(tx)
	reqres := abcicli.NewReqRes(reqp)
	mtx := new(sync.Mutex)
	mtx.Lock()
	app.deliverTxQueue <- WorkItem{reqRes: reqres, mtx: mtx}
	app.log.Debug("Enqueue DeliverTxAsync", "Tx", hex.EncodeToString(tx[:7]))
	//no need to lock commitLock because Commit and DeliverTx will not be called concurrently
	app.wgCommit.Add(1)
	app.deliverTxPool.Schedule(func() {
		app.log.Debug("Start PreDeliverTx", "Tx", hex.EncodeToString(tx[:7]))
		res := app.Application.PreDeliverTx(tx)
		if !res.IsOK() { // no need to call the real DeliverTx
			reqres.Response = types.ToResponseDeliverTx(res)
		}
		app.log.Debug("Finish PreDeliverTx", "Tx", hex.EncodeToString(tx[:7]))
		mtx.Unlock()
	})

	return reqres
}

func (app *asyncLocalClient) CheckTxAsync(tx []byte) *abcicli.ReqRes {
	// no app level lock because the real CheckTx would be called in the worker routine
	reqp := types.ToRequestCheckTx(tx)
	reqres := abcicli.NewReqRes(reqp)
	mtx := new(sync.Mutex)
	mtx.Lock()
	app.checkTxLowLock.Lock()
	app.checkTxMidLock.Lock()
	app.commitLock.Lock() // here would block further queue if commit is ready to go
	app.checkTxMidLock.Unlock()
	app.checkTxQueue <- WorkItem{reqRes: reqres, mtx: mtx}
	app.log.Debug("Enqueue CheckTxAsync", "Tx", hex.EncodeToString(tx[:7]))
	app.wgCommit.Add(1)
	app.commitLock.Unlock()
	app.checkTxLowLock.Unlock()
	app.checkTxPool.Schedule(func() {
		app.log.Debug("Start PreCheckTx", "Tx", hex.EncodeToString(tx[:7]))
		res := app.Application.PreCheckTx(tx)
		if !res.IsOK() { // no need to call the real CheckTx
			reqres.Response = types.ToResponseCheckTx(res)
		}
		app.log.Debug("Finish PreCheckTx", "Tx", hex.EncodeToString(tx[:7]))
		mtx.Unlock()
	})
	return reqres
}

//ReCheckTxAsync here still runs synchronously
func (app *asyncLocalClient) ReCheckTxAsync(tx []byte) *abcicli.ReqRes {
	app.rwLock.Lock() // wont
	app.log.Debug("Start ReCheckAsync", "Tx", hex.EncodeToString(tx[:7]))
	res := app.Application.ReCheckTx(tx)
	app.log.Debug("Finish ReCheckAsync", "Tx", hex.EncodeToString(tx[:7]))
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestCheckTx(tx),
		types.ToResponseCheckTx(res),
	)
}

// QueryAsync is supposed to run concurrently when there is no CheckTx/DeliverTx/Commit
func (app *asyncLocalClient) QueryAsync(req types.RequestQuery) *abcicli.ReqRes {
	app.rwLock.RLock()
	app.rwDeliverLock.RLock()
	res := app.Application.Query(req)
	app.rwDeliverLock.RUnlock()
	app.rwLock.RUnlock()
	return app.callback(
		types.ToRequestQuery(req),
		types.ToResponseQuery(res),
	)
}

func (app *asyncLocalClient) CommitAsync() *abcicli.ReqRes {
	app.log.Debug("Trying to get CommitAsync lock")
	app.checkTxMidLock.Lock()
	app.commitLock.Lock() // this must come before the wgCommit.Wait()
	app.checkTxMidLock.Unlock()
	app.wgCommit.Wait() // wait for all the submitted CheckTx/DeliverTx/Query finish
	app.rwLock.Lock()
	// only checkTxLock is locked here
	// because we trust deliver and commit will not call concurrently
	app.log.Debug("Start CommitAsync")
	res := app.Application.Commit()
	app.log.Debug("Finish CommitAsync")
	app.rwLock.Unlock()
	app.commitLock.Unlock()
	return app.callback(
		types.ToRequestCommit(),
		types.ToResponseCommit(res),
	)
}

func (app *asyncLocalClient) InitChainAsync(req types.RequestInitChain) *abcicli.ReqRes {
	app.rwLock.Lock()
	app.rwDeliverLock.Lock()
	res := app.Application.InitChain(req)
	reqRes := app.callback(
		types.ToRequestInitChain(req),
		types.ToResponseInitChain(res),
	)
	app.rwDeliverLock.Unlock()
	app.rwLock.Unlock()
	return reqRes
}

func (app *asyncLocalClient) BeginBlockAsync(req types.RequestBeginBlock) *abcicli.ReqRes {
	app.rwLock.Lock()
	res := app.Application.BeginBlock(req)
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestBeginBlock(req),
		types.ToResponseBeginBlock(res),
	)
}

func (app *asyncLocalClient) EndBlockAsync(req types.RequestEndBlock) *abcicli.ReqRes {
	app.log.Debug("Trying to get EndBlockAsync lock")
	app.checkTxMidLock.Lock()
	app.commitLock.Lock() // this must come before the wgCommit.Wait()
	app.checkTxMidLock.Unlock()
	app.wgCommit.Wait() // wait for all the submitted CheckTx/DeliverTx/Query finish
	app.rwLock.Lock()
	// only checkTxLock is locked here
	// because we trust deliver and commit will not call concurrently
	app.log.Debug("Starting EndBlockAsync")
	res := app.Application.EndBlock(req)
	app.log.Debug("Finish EndBlockAsync")
	app.rwLock.Unlock()
	app.commitLock.Unlock()
	return app.callback(
		types.ToRequestEndBlock(req),
		types.ToResponseEndBlock(res),
	)
}

//-------------------------------------------------------

func (app *asyncLocalClient) FlushSync() error {
	return nil
}

func (app *asyncLocalClient) EchoSync(msg string) (*types.ResponseEcho, error) {
	return &types.ResponseEcho{Message: msg}, nil
}

func (app *asyncLocalClient) InfoSync(req types.RequestInfo) (*types.ResponseInfo, error) {
	app.rwLock.RLock()
	app.rwDeliverLock.RLock()
	res := app.Application.Info(req)
	app.rwDeliverLock.RUnlock()
	app.rwLock.RUnlock()
	return &res, nil
}

func (app *asyncLocalClient) SetOptionSync(req types.RequestSetOption) (*types.ResponseSetOption, error) {
	app.rwLock.Lock()
	app.rwDeliverLock.Lock()
	res := app.Application.SetOption(req)
	app.rwDeliverLock.Unlock()
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) DeliverTxSync(tx []byte) (*types.ResponseDeliverTx, error) {
	app.rwDeliverLock.Lock()
	app.log.Debug("Start DeliverTxSync")
	res := app.Application.DeliverTx(tx)
	app.rwDeliverLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) CheckTxSync(tx []byte) (*types.ResponseCheckTx, error) {
	app.rwLock.Lock()
	app.log.Debug("Start CheckTxSync")
	res := app.Application.CheckTx(tx)
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) QuerySync(req types.RequestQuery) (*types.ResponseQuery, error) {
	app.rwLock.RLock()
	app.rwDeliverLock.RLock()
	res := app.Application.Query(req)
	app.rwDeliverLock.RUnlock()
	app.rwLock.RUnlock()
	return &res, nil
}

func (app *asyncLocalClient) CommitSync() (*types.ResponseCommit, error) {
	app.log.Debug("Trying to get CommitSync Lock")
	app.checkTxMidLock.Lock()
	app.commitLock.Lock() // this must come before the wgCommit.Wait()
	app.checkTxMidLock.Unlock()
	app.wgCommit.Wait() // wait for all the submitted CheckTx/DeliverTx/Query finish
	app.rwLock.Lock()
	// only checkTxLock is locked here
	// because we trust deliver and commit will not call concurrently
	app.log.Debug("Start CommitSync")
	res := app.Application.Commit()
	app.log.Debug("Finish CommitSync")
	app.rwLock.Unlock()
	app.commitLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) InitChainSync(req types.RequestInitChain) (*types.ResponseInitChain, error) {
	app.rwLock.Lock()
	app.rwDeliverLock.Lock()
	res := app.Application.InitChain(req)
	app.rwDeliverLock.Unlock()
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) BeginBlockSync(req types.RequestBeginBlock) (*types.ResponseBeginBlock, error) {
	app.rwLock.Lock()
	res := app.Application.BeginBlock(req)
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) EndBlockSync(req types.RequestEndBlock) (*types.ResponseEndBlock, error) {
	app.log.Debug("Trying to get EndBlockSync lock")
	app.checkTxMidLock.Lock()
	app.commitLock.Lock() // this must come before the wgCommit.Wait()
	app.checkTxMidLock.Unlock()
	app.wgCommit.Wait() // wait for all the submitted CheckTx/DeliverTx/Query finish
	app.rwLock.Lock()
	app.log.Debug("Start EndBlockSync")
	// only checkTxLock is locked here
	// because we trust deliver and commit will not call concurrently
	res := app.Application.EndBlock(req)
	app.log.Debug("Finish EndBlockSync")
	app.rwLock.Unlock()
	app.commitLock.Unlock()
	return &res, nil
}

//-------------------------------------------------------

func (app *asyncLocalClient) callback(req *types.Request, res *types.Response) *abcicli.ReqRes {
	app.Callback(req, res)
	return newLocalReqRes(req, res)
}

func newLocalReqRes(req *types.Request, res *types.Response) *abcicli.ReqRes {
	reqRes := abcicli.NewReqRes(req)
	reqRes.Response = res
	reqRes.SetDone()
	return reqRes
}

type localAsyncClientCreator struct {
	app types.Application
	log log.Logger
}

func NewAsyncLocalClientCreator(app types.Application, log log.Logger) proxy.ClientCreator {
	return &localAsyncClientCreator{
		app: app,
		log: log,
	}
}

func (l *localAsyncClientCreator) NewABCIClient() (abcicli.Client, error) {
	return NewAsyncLocalClient(l.app, l.log), nil
}
