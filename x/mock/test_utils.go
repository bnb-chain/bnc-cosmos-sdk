package mock

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// BigInterval is a representation of the interval [lo, hi), where
// lo and hi are both of type sdk.Int
type BigInterval struct {
	lo sdk.Int
	hi sdk.Int
}

// RandFromBigInterval chooses an interval uniformly from the provided list of
// BigIntervals, and then chooses an element from an interval uniformly at random.
func RandFromBigInterval(r *rand.Rand, intervals []BigInterval) sdk.Int {
	if len(intervals) == 0 {
		return sdk.ZeroInt()
	}

	interval := intervals[r.Intn(len(intervals))]

	lo := interval.lo
	hi := interval.hi

	diff := hi.Sub(lo)
	result := sdk.NewIntFromBigInt(new(big.Int).Rand(r, diff.BigInt()))
	result = result.Add(lo)

	return result
}

// CheckBalance checks the balance of an account.
func CheckBalance(t *testing.T, app *App, addr sdk.AccAddress, exp sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{})
	res := app.AccountKeeper.GetAccount(ctxCheck, addr)

	require.Equal(t, exp, res.GetCoins())
}

// CheckGenTx checks a generated signed transaction. The result of the check is
// compared against the parameter 'expPass'. A test assertion is made using the
// parameter 'expPass' against the result. A corresponding result is returned.
func CheckGenTx(
	t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg, accNums []int64,
	seq []int64, expPass bool, priv ...crypto.PrivKey,
) sdk.Result {
	tx := GenTx(msgs, accNums, seq, priv...)
	res := app.Check(tx)

	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	return res
}

// SignCheckDeliver checks a generated signed transaction and simulates a
// block commitment with the given transaction. A test assertion is made using
// the parameter 'expPass' against the result. A corresponding result is
// returned.
func SignCheckDeliver(
	t *testing.T, app *baseapp.BaseApp, msgs []sdk.Msg, accNums []int64,
	seq []int64, expSimPass, expPass bool, priv ...crypto.PrivKey,
) sdk.Result {
	tx := GenTx(msgs, accNums, seq, priv...)
	// Must simulate now as CheckTx doesn't run Msgs anymore
	res := app.Simulate(nil, tx)

	if expSimPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	// Simulate a sending a transaction and committing a block
	app.BeginBlock(abci.RequestBeginBlock{})
	res = app.Deliver(tx)

	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return res
}

func GetAccount(app *App, addr sdk.AccAddress) sdk.Account {
	ctxCheck := app.BaseApp.NewContext(sdk.RunTxModeCheck, abci.Header{})
	res := app.AccountKeeper.GetAccount(ctxCheck, addr)
	return res
}

func GenSimTxs(
	t *testing.T, app *App, msgs []sdk.Msg, expSimPass bool, privs ...crypto.PrivKey,
) (txs []auth.StdTx) {
	accSeqMap := make(map[string][2]int64)
	for i, priv := range privs {
		addr := sdk.AccAddress(priv.PubKey().Address())
		accNumSeq, found := accSeqMap[addr.String()]
		if !found {
			acc := GetAccount(app, addr)
			if acc == nil {
				panic(fmt.Sprintf("account %s not found", addr))
			}
			accNumSeq[0] = acc.GetAccountNumber()
			accNumSeq[1] = acc.GetSequence()
		}
		tx := GenTx(msgs[i:i+1], []int64{accNumSeq[0]}, []int64{accNumSeq[1]}, priv)
		res := app.Simulate(nil, tx)
		if expSimPass {
			require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
		} else {
			require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
		}
		accSeqMap[addr.String()] = [2]int64{accNumSeq[0], accNumSeq[1] + 1}
		txs = append(txs, tx)
	}
	return txs
}

func ApplyBlock(t *testing.T, app *baseapp.BaseApp, height int64, txs []auth.StdTx) (newHeight int64) {
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Height: height,
	}})
	for _, tx := range txs {
		res := app.Deliver(tx)
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	}
	app.EndBlock(abci.RequestEndBlock{Height: height})
	app.Commit()
	return height + 1
}

func ApplyEmptyBlocks(t *testing.T, app *baseapp.BaseApp, height int64, blockNum int) (newHeight int64) {
	for i := 0; i < blockNum; i++ {
		height = ApplyBlock(t, app, height, []auth.StdTx{})
	}
	return height
}
