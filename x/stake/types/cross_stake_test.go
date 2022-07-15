package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetStakeCAoB(t *testing.T) {
	exp, err := sdk.AccAddressFromHex("9fB29AAc15b9A4B7F17c3385939b007540f4d791")
	if err != nil {
		t.Fatal(err)
	}
	stakeCAoB := GetStakeCAoB(exp.Bytes(), "Delegate")
	fmt.Println(stakeCAoB.String())
	if delAddr := GetStakeCAoB(stakeCAoB.Bytes(), "Delegate"); delAddr.String() != exp.String() {
		t.Fatal()
	}
}
