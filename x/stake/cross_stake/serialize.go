package cross_stake

import (
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

type CrossStakePackageType uint8

const (
	CrossStakeTypeDelegate         CrossStakePackageType = 1
	CrossStakeTypeUndelegate       CrossStakePackageType = 2
	CrossStakeTypeClaimReward      CrossStakePackageType = 3
	CrossStakeTypeClaimUndelegated CrossStakePackageType = 4
	CrossStakeTypeReinvest         CrossStakePackageType = 5
	CrossStakeTypeRedelegate       CrossStakePackageType = 6
)

type CrossStakeSynPackage struct {
	PackageType CrossStakePackageType
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
