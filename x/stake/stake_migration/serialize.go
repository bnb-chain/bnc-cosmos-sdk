package stake_migration

import (
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func DeserializeStakeMigrationRefundPackage(serializedPackage []byte) (*types.StakeMigrationSynPackage, error) {
	var pack types.StakeMigrationSynPackage
	err := rlp.DecodeBytes(serializedPackage, &pack)
	if err != nil {
		return nil, err
	}
	return &pack, nil
}
