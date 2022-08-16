package cross_stake

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestDeserializeSynPackage(t *testing.T) {
	input := "f84a03b847f845947de3642c66220e1136a42bf9897a6f8527ef9a0394beb218281ac3ebc4de700e7fcf23eea39010b8a3947de3642c66220e1136a42bf9897a6f8527ef9a038504c4b40000"

	packageBytes, err := hex.DecodeString(input)
	if err != nil {
		t.Fatal(err)
	}

	pack, err := DeserializeCrossStakeSynPackage(packageBytes)
	if err != nil {
		t.Fatal(err)
	}
	switch pack.(type) {
	case *types.CrossStakeRedelegateSynPackage:
	default:
		t.Error("wrong event type")
	}
}
