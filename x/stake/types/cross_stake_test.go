package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetStakeCAoB(t *testing.T) {
	exp, err := sdk.AccAddressFromHex("0000000000000000000000000000000000001000")
	if err != nil {
		t.Fatal(err)
	}
	stakeCAoB := GetStakeCAoB(exp.Bytes(), "Delegate")
	if delAddr := GetStakeCAoB(stakeCAoB.Bytes(), "Delegate"); delAddr.String() != exp.String() {
		t.Fatal()
	}
}
