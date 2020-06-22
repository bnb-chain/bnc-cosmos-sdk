package sidechain

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MaxSideChainIdLength = 20
)

const (
	CrossChainFeeLength = 32
	PackageTypeLength   = 1
	PackageHeaderLength = CrossChainFeeLength + PackageTypeLength
)

func EncodePackageHeader(packageType sdk.CrossChainPackageType, relayerFee big.Int) []byte {
	packageHeader := make([]byte, PackageHeaderLength)
	packageHeader[0] = uint8(packageType)
	length := len(relayerFee.Bytes())
	copy(packageHeader[PackageHeaderLength-length:PackageHeaderLength], relayerFee.Bytes())
	return packageHeader
}

func DecodePackageHeader(packageHeader []byte) (packageType sdk.CrossChainPackageType, relayFee big.Int, err error) {
	if len(packageHeader) < PackageHeaderLength {
		err = fmt.Errorf("length of packageHeader is less than %d", PackageHeaderLength)
		return
	}
	packageType = sdk.CrossChainPackageType(packageHeader[0])
	relayFee.SetBytes(packageHeader[PackageTypeLength : CrossChainFeeLength+PackageTypeLength])
	return
}

type CommonAckPackage struct {
	Code uint32
}

func (p CommonAckPackage) IsOk() bool {
	return p.Code == 0
}

func GenCommonAckPackage(code uint32) ([]byte, error) {
	return rlp.EncodeToBytes(&CommonAckPackage{Code: code})
}
