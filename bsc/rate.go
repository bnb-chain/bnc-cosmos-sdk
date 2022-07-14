package bsc

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	BNBDecimalOnBC  = 8
	BNBDecimalOnBSC = 18
)

func ConvertBCAmountToBSCAmount(bcAmount int64) *big.Int {
	decimals := sdk.NewIntWithDecimal(1, int(BNBDecimalOnBSC-BNBDecimalOnBC))
	bscAmount := sdk.NewInt(bcAmount).Mul(decimals)
	return bscAmount.BigInt()
}

func ConvertBSCAmountToBCAmount(bscAmount *big.Int) int64 {
	decimals := sdk.NewIntWithDecimal(1, int(BNBDecimalOnBSC-BNBDecimalOnBC))
	bscAmountInt := sdk.NewIntFromBigInt(bscAmount)
	bcAmount := bscAmountInt.Div(decimals)
	return bcAmount.Int64()
}
