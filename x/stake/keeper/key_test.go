package keeper

import (
	"encoding/hex"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	pk1   = ed25519.GenPrivKeyFromSecret([]byte{1}).PubKey()
	pk2   = ed25519.GenPrivKeyFromSecret([]byte{2}).PubKey()
	pk3   = ed25519.GenPrivKeyFromSecret([]byte{3}).PubKey()
	addr1 = pk1.Address()
	addr2 = pk2.Address()
	addr3 = pk3.Address()
)

func TestGetValidatorPowerRank(t *testing.T) {
	valAddr1 := sdk.ValAddress(addr1)
	emptyDesc := types.Description{}
	val1 := types.NewValidator(valAddr1, pk1, emptyDesc)
	val1.Tokens = sdk.NewDec(0)
	val2, val3, val4 := val1, val1, val1
	val2.Tokens = sdk.NewDecWithoutFra(1)
	val3.Tokens = sdk.NewDecWithoutFra(10)
	x := new(big.Int).Exp(big.NewInt(2), big.NewInt(20), big.NewInt(0))
	val4.Tokens = sdk.NewDecWithoutFra(x.Int64())

	tests := []struct {
		validator types.Validator
		wantHex   string
	}{
		{val1, "2300000000000000009c288ede7df62742fc3b7d0962045a8cef0f79f6"},
		{val2, "230000000005f5e1009c288ede7df62742fc3b7d0962045a8cef0f79f6"}, // "5f5e100" is	100000000 in base 10.
		{val3, "23000000003b9aca009c288ede7df62742fc3b7d0962045a8cef0f79f6"}, // "3b9aca00" is 1000000000 in base 10
		{val4, "2300005f5e100000009c288ede7df62742fc3b7d0962045a8cef0f79f6"}, // "5f5e10000000" is 2^20.e8 in base 10
	}
	for i, tt := range tests {
		got := hex.EncodeToString(getValidatorPowerRank(tt.validator))

		assert.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}

func TestGetREDByValDstIndexKey(t *testing.T) {
	tests := []struct {
		delAddr    sdk.AccAddress
		valSrcAddr sdk.ValAddress
		valDstAddr sdk.ValAddress
		wantHex    string
	}{
		{sdk.AccAddress(addr1), sdk.ValAddress(addr1), sdk.ValAddress(addr1),
			"3663d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f08609"},
		{sdk.AccAddress(addr1), sdk.ValAddress(addr2), sdk.ValAddress(addr3),
			"363ab62f0d93849be495e21e3e9013a517038f45bd63d771218209d8bd03c482f69dfba57310f086095ef3b5f25c54946d4a89fc0d09d2f126614540f2"},
		{sdk.AccAddress(addr2), sdk.ValAddress(addr1), sdk.ValAddress(addr3),
			"363ab62f0d93849be495e21e3e9013a517038f45bd5ef3b5f25c54946d4a89fc0d09d2f126614540f263d771218209d8bd03c482f69dfba57310f08609"},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(GetREDByValDstIndexKey(tt.delAddr, tt.valSrcAddr, tt.valDstAddr))

		assert.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}

func TestGetREDByValSrcIndexKey(t *testing.T) {
	tests := []struct {
		delAddr    sdk.AccAddress
		valSrcAddr sdk.ValAddress
		valDstAddr sdk.ValAddress
		wantHex    string
	}{
		{sdk.AccAddress(addr1), sdk.ValAddress(addr1), sdk.ValAddress(addr1),
			"3563d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f0860963d771218209d8bd03c482f69dfba57310f08609"},
		{sdk.AccAddress(addr1), sdk.ValAddress(addr2), sdk.ValAddress(addr3),
			"355ef3b5f25c54946d4a89fc0d09d2f126614540f263d771218209d8bd03c482f69dfba57310f086093ab62f0d93849be495e21e3e9013a517038f45bd"},
		{sdk.AccAddress(addr2), sdk.ValAddress(addr1), sdk.ValAddress(addr3),
			"3563d771218209d8bd03c482f69dfba57310f086095ef3b5f25c54946d4a89fc0d09d2f126614540f23ab62f0d93849be495e21e3e9013a517038f45bd"},
	}
	for i, tt := range tests {
		got := hex.EncodeToString(GetREDByValSrcIndexKey(tt.delAddr, tt.valSrcAddr, tt.valDstAddr))

		assert.Equal(t, tt.wantHex, got, "Keys did not match on test case %d", i)
	}
}
