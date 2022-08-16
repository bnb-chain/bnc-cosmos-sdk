package types

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/bsc/rlp"
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

func TestAckRLP(t *testing.T) {
	delAddr, _ := sdk.NewSmartChainAddress("91D7deA99716Cbb247E81F1cfB692009164a967E")

	bcAddr := "bnb1dmrarep5fawa89shw0048syv3ek4tcm28tmqp6"

	bz, _ := sdk.GetFromBech32(bcAddr, "bnb")
	valAddr := sdk.ValAddress(bz)
	synPack := CrossStakeDelegateSynPackage{
		DelAddr:   delAddr,
		Validator: valAddr,
		Amount:    big.NewInt(1000),
	}

	ackPack := NewCrossStakeDelegationAckPackage(&synPack, CrossStakeTypeDelegate, 1)
	ackBytes, _ := rlp.EncodeToBytes(ackPack)

	type AckPackage struct {
		EventType CrossStakeEventType
		DelAddr   sdk.SmartChainAddress
		Validator sdk.ValAddress
		Amount    *big.Int
		ErrorCode uint8
	}
	var pack AckPackage
	err := rlp.DecodeBytes(ackBytes, &pack)
	if err != nil {
		log.Fatal(err)
	}
}
