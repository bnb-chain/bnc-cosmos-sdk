package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetStakeCAoB(t *testing.T) {
	exp, err := sdk.AccAddressFromHex("91D7deA99716Cbb247E81F1cfB692009164a967E")
	if err != nil {
		t.Fatal(err)
	}
	stakeCAoB := GetStakeCAoB(exp.Bytes(), DelegateCAoBSalt)
	fmt.Println(stakeCAoB.String())
	if delAddr := GetStakeCAoB(stakeCAoB.Bytes(), DelegateCAoBSalt); delAddr.String() != exp.String() {
		t.Fatal()
	}
}
