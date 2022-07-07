package cross_stake

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestDeserializeDelegatePackage(t *testing.T) {
	input := "f83901945b38da6a701c568545dcfcb03fcb875f56beddc4940000000000000000000000000000000000001000880de0b6b3a76400008462c67bae"
	exp := types.CrossStakeTypeDelegate

	packageBytes, err := hex.DecodeString(input)
	if err != nil {
		t.Fatal(err)
	}
	eventType, err := DeserializeCrossStakeSynPackage(packageBytes)
	if err != nil {
		t.Fatal(err)
	}
	if eventType != exp {
		t.Error("wrong event type")
	}
}
