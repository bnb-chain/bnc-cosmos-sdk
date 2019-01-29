package concurrent

import "github.com/tendermint/tendermint/abci/types"

type ApplicationCC interface {
	types.Application
	PreCheckTx(tx []byte) types.ResponseCheckTx
	PreDeliverTx(tx []byte) types.ResponseDeliverTx
}
