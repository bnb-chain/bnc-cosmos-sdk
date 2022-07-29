package cross_stake

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestDeserializeDelegatePackage(t *testing.T) {
	input := "f701b5f401945b38da6a701c568545dcfcb03fcb875f56beddc4940000000000000000000000000000000000001000880de0b6b3a7640000"

	packageBytes, err := hex.DecodeString(input)
	if err != nil {
		t.Fatal(err)
	}

	pack, err := DeserializeCrossStakeSynPackage(packageBytes)
	if err != nil {
		t.Fatal(err)
	}
	switch pack.(type) {
	case types.CrossStakeDelegateSynPackage:
		break
	default:
		t.Error("wrong event type")
	}
}
