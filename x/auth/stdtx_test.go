package auth

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestStdTx(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	sigs := []StdSignature{}

	tx := NewStdTx(msgs, sigs, "")
	require.Equal(t, msgs, tx.GetMsgs())
	require.Equal(t, sigs, tx.GetSignatures())
}

func TestStdSignBytes(t *testing.T) {
	type args struct {
		chainID  string
		accnum   int64
		sequence int64
		msgs     []sdk.Msg
		memo     string
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args{"1234", 3, 6,  []sdk.Msg{sdk.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf("{\"account_number\":\"3\",\"chain_id\":\"1234\",\"memo\":\"memo\",\"msgs\":[[\"%s\"]],\"sequence\":\"6\"}", addr),
		},
	}
	for i, tc := range tests {
		got := string(StdSignBytes(tc.args.chainID, tc.args.accnum, tc.args.sequence, tc.args.msgs, tc.args.memo))
		require.Equal(t, tc.want, got, "Got unexpected result on test case i: %d", i)
	}
}
