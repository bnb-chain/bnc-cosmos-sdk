package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/sidechain"
)

func TestClaimMsg(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.Coins{})

	tests := []struct {
		claimMsg     ClaimMsg
		expectedPass bool
	}{
		{
			NewClaimMsg(1, 1, common.RandBytes(sidechain.PackageHeaderLength), addrs[0]),
			true,
		}, {
			NewClaimMsg(1, 1, []byte("test"), addrs[0]),
			false,
		}, {
			NewClaimMsg(1, 1, common.RandBytes(sidechain.PackageHeaderLength), sdk.AccAddress{1}),
			false,
		},
	}

	for i, test := range tests {
		if test.expectedPass {
			require.Nil(t, test.claimMsg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, test.claimMsg.ValidateBasic(), "test: %v", i)
		}
	}
}
