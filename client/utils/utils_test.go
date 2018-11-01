package utils

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryResponse(t *testing.T) {
	cdc := app.MakeCodec()
	sdkResBytes := cdc.MustMarshalBinary(sdk.Result{})
	err := parseQueryResponse(cdc, sdkResBytes)
	assert.Nil(t, err)
	err = parseQueryResponse(cdc, []byte("fuzzy"))
	assert.NotNil(t, err)
}
