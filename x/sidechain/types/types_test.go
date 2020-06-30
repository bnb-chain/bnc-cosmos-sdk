package types

import (
	"encoding/hex"
	"github.com/magiconair/properties/assert"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func Test_EncodePackageHeader(t *testing.T) {
	bz := EncodePackageHeader(types.SynCrossChainPackageType, *big.NewInt(10000000000000000))
	assert.Equal(t, hex.EncodeToString(bz), hex.EncodeToString(bz))
}
