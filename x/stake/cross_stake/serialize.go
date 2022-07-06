package cross_stake

import (
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	CrossStakeErrCodeExpired       uint8 = 1
	CrossStakeErrValidatorNotFound uint8 = 2
	CrossStakeErrValidatorJailed   uint8 = 3
	CrossStakeErrBadDelegation     uint8 = 4
)

type CrossStakeSynPackage struct {
	PackageType types.CrossStakePackageType
	params      []byte
}

func DeserializeCrossStakeSynPackage(serializedPackage []byte) (*CrossStakeSynPackage, error) {
	var pack CrossStakeSynPackage
	err := rlp.DecodeBytes(serializedPackage, &pack)
	if err != nil {
		return nil, types.ErrDeserializePackageFailed("deserialize cross stake syn package failed")
	}
	return &pack, nil
}
