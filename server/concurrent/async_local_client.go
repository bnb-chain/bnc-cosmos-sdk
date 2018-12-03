package concurrent

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/server/concurrent/pool"
	client "github.com/tendermint/tendermint/abci/client"
	types "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var _ Client = (*asyncLocalClient)(nil)

// asyncLocalClient is a variant from local_client.
// It makes ABCI calling more complex:
// 1. CheckTx/DeliverTx/Query/Info can be called concurrently
// 2. Other API would block calling CheckTx/DeliverTx/Query
// 3. CheckTx/DeliverTx/Query implementation would implement
//    another level of synchronization to guarantee Application
//    execution correctness.

const (
	WorkerPoolSize  = 8
	WorkerPoolSpawn = 4
	WorkerPoolQueue = 8
)

type asyncLocalClient struct {
	cmn.BaseService
	rwLock *sync.RWMutex
	types.Application
	client.Callback

	checkTxPool   *pool.Pool
	deliverTxPool *pool.Pool
}

func NewAsyncLocalClient(app types.Application) *asyncLocalClient {
	cli := &asyncLocalClient{
		rwLock:        new(sync.RWMutex),
		Application:   app,
		checkTxPool:   pool.NewPool(WorkerPoolSize/2, WorkerPoolQueue/2, WorkerPoolSpawn/2),
		deliverTxPool: pool.NewPool(WorkerPoolSize, WorkerPoolQueue, WorkerPoolSpawn),
	}
	cli.BaseService = *cmn.NewBaseService(nil, "asyncLocalClient", cli)
	return cli
}

func (app *asyncLocalClient) SetResponseCallback(cb Callback) {
	app.rwLock.Lock()
	defer app.rwLock.Unlock()
	app.Callback = cb
}

// TODO: change types.Application to include Error()?
func (app *asyncLocalClient) Error() error {
	return nil
}

func (app *asyncLocalClient) FlushAsync() *ReqRes {
	// Do nothing
	return newLocalReqRes(types.ToRequestFlush(), nil)
}

func (app *asyncLocalClient) EchoAsync(msg string) *ReqRes {
	return app.callback(
		types.ToRequestEcho(msg),
		types.ToResponseEcho(msg),
	)
}

func (app *asyncLocalClient) InfoAsync(req types.RequestInfo) *ReqRes {
	app.rwLock.RLock()
	reqp := types.ToRequestInfo(req)
	reqres := NewReqRes(reqp)
	app.checkTxPool.Schedule(func() {
		res := app.Application.Info(req)
		reqres.Response = types.ToResponseInfo(res) // Set response
		reqres.Done()
		app.callback(reqp, reqres.Response)
	})
	app.rwLock.Unlock()
	return reqres
}

func (app *asyncLocalClient) SetOptionAsync(req types.RequestSetOption) *ReqRes {
	app.rwLock.Lock()
	res := app.Application.SetOption(req)
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestSetOption(req),
		types.ToResponseSetOption(res),
	)
}

func (app *asyncLocalClient) DeliverTxAsync(tx []byte) *ReqRes {
	app.rwLock.RLock()
	reqp := types.ToRequestDeliverTx(tx)
	reqres := NewReqRes(reqp)
	app.deliverTxPool.Schedule(func() {
		res := app.Application.DeliverTx(tx)
	})

	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestDeliverTx(tx),
		types.ToResponseDeliverTx(res),
	)
}

func (app *asyncLocalClient) CheckTxAsync(tx []byte) *ReqRes {
	app.rwLock.RLock()
	res := app.Application.CheckTx(tx)
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestCheckTx(tx),
		types.ToResponseCheckTx(res),
	)
}

func (app *asyncLocalClient) ReCheckTxAsync(tx []byte) *ReqRes {
	app.rwLock.RLock()
	res := app.Application.ReCheckTx(tx)
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestCheckTx(tx),
		types.ToResponseCheckTx(res),
	)
}

func (app *asyncLocalClient) QueryAsync(req types.RequestQuery) *ReqRes {
	app.rwLock.RLock()
	res := app.Application.Query(req)
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestQuery(req),
		types.ToResponseQuery(res),
	)
}

func (app *asyncLocalClient) CommitAsync() *ReqRes {
	app.rwLock.Lock()
	res := app.Application.Commit()
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestCommit(),
		types.ToResponseCommit(res),
	)
}

func (app *asyncLocalClient) InitChainAsync(req types.RequestInitChain) *ReqRes {
	app.rwLock.Lock()
	res := app.Application.InitChain(req)
	reqRes := app.callback(
		types.ToRequestInitChain(req),
		types.ToResponseInitChain(res),
	)
	app.rwLock.Unlock()
	return reqRes
}

func (app *asyncLocalClient) BeginBlockAsync(req types.RequestBeginBlock) *ReqRes {
	app.rwLock.Lock()
	res := app.Application.BeginBlock(req)
	app.rwLock.Unlock()
	return app.callback(
		types.ToRequestBeginBlock(req),
		types.ToResponseBeginBlock(res),
	)
}

func (app *asyncLocalClient) EndBlockAsync(req types.RequestEndBlock) *ReqRes {
	app.rwLock.Lock()
	res := app.Application.EndBlock(req)
	app.rwLock.Unlock()
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
	res := app.Application.Info(req)
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) SetOptionSync(req types.RequestSetOption) (*types.ResponseSetOption, error) {
	app.rwLock.Lock()
	res := app.Application.SetOption(req)
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) DeliverTxSync(tx []byte) (*types.ResponseDeliverTx, error) {
	app.rwLock.Lock()
	res := app.Application.DeliverTx(tx)
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) CheckTxSync(tx []byte) (*types.ResponseCheckTx, error) {
	app.rwLock.Lock()
	res := app.Application.CheckTx(tx)
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) QuerySync(req types.RequestQuery) (*types.ResponseQuery, error) {
	app.rwLock.Lock()
	res := app.Application.Query(req)
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) CommitSync() (*types.ResponseCommit, error) {
	app.rwLock.Lock()
	res := app.Application.Commit()
	app.rwLock.Unlock()
	return &res, nil
}

func (app *asyncLocalClient) InitChainSync(req types.RequestInitChain) (*types.ResponseInitChain, error) {
	app.rwLock.Lock()
	res := app.Application.InitChain(req)
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
	app.rwLock.Lock()
	res := app.Application.EndBlock(req)
	app.rwLock.Unlock()
	return &res, nil
}

//-------------------------------------------------------

func (app *asyncLocalClient) callback(req *types.Request, res *types.Response) *ReqRes {
	app.Callback(req, res)
	return newLocalReqRes(req, res)
}

type localAsyncClientCreator struct {
	app types.Application
}

func NewAsyncLocalClientCreator(app types.Application) ClientCreator {
	return &localAsyncClientCreator{
		app: app,
	}
}

func (l *localAsyncClientCreator) NewABCIClient() (client.Client, error) {
	return NewAsyncLocalClient(l.app), nil
}
