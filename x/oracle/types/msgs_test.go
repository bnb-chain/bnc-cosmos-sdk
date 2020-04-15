package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
)

func TestClaimMsg(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.Coins{})

	tests := []struct {
		claimMsg     ClaimMsg
		expectedPass bool
	}{
		{
			NewClaimMsg(sdk.ClaimType(0x8), 1, "test", addrs[0]),
			true,
		}, {
			NewClaimMsg(sdk.ClaimType(0x8), -1, "test", addrs[0]),
			false,
		}, {
			NewClaimMsg(sdk.ClaimType(0x8), 1, "", addrs[0]),
			false,
		}, {
			NewClaimMsg(sdk.ClaimType(0x8), 1, "test", sdk.AccAddress{1}),
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
