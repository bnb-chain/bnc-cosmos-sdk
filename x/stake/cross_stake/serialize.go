package cross_stake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	CrossStakeErrValidatorNotFound uint8 = 1
	CrossStakeErrValidatorJailed   uint8 = 2
	CrossStakeErrBadDelegation     uint8 = 3
)

type CrossStakePackageFromBSC struct {
	EventCode   types.CrossStakePackageType
	ParamsBytes []byte
}

func DeserializeCrossStakeSynPackage(serializedPackage []byte) (interface{}, error) {
	var pack1 CrossStakePackageFromBSC
	err := rlp.DecodeBytes(serializedPackage, &pack1)
	if err != nil {
		return nil, err
	}
	switch pack1.EventCode {
	case types.CrossStakeTypeDelegate:
		var pack2 types.CrossStakeDelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return pack2, nil
	case types.CrossStakeTypeUndelegate:
		var pack2 types.CrossStakeUndelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return pack2, nil
	case types.CrossStakeTypeRedelegate:
		var pack2 types.CrossStakeRedelegateSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return pack2, nil
	default:
		return nil, fmt.Errorf("unrecognized package type")
	}
}

func DeserializeCrossStakeAckPackage(serializedPackage []byte) (interface{}, error) {
	var pack1 CrossStakePackageFromBSC
	err := rlp.DecodeBytes(serializedPackage, &pack1)
	if err != nil {
		return nil, err
	}
	switch pack1.EventCode {
	case types.CrossStakeTypeDistributeReward:
		var pack2 types.CrossStakeDistributeRewardSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return pack2, nil
	case types.CrossStakeTypeDistributeUndelegated:
		var pack2 types.CrossStakeDistributeUndelegatedSynPackage
		err := rlp.DecodeBytes(pack1.ParamsBytes, &pack2)
		if err != nil {
			return nil, err
		}
		return pack2, nil
	default:
		return nil, fmt.Errorf("unrecognized package type")
	}
}

func DeserializeCrossStakeFailAckPackage(serializedPackage []byte) (interface{}, error) {
	deserializeIntoUndelegatedPackage := func(serializedPackage []byte) (interface{}, error) {
		var pack types.CrossStakeDistributeUndelegatedSynPackage
		err := rlp.DecodeBytes(serializedPackage, &pack)
		if err != nil {
			return nil, err
		}
		return pack, nil
	}

	deserializeIntoRewardPackage := func(serializedPackage []byte) (interface{}, error) {
		var pack types.CrossStakeDistributeRewardSynPackage
		err := rlp.DecodeBytes(serializedPackage, &pack)
		if err != nil {
			return nil, err
		}
		return pack, nil
	}

	var pack interface{}
	pack, err := deserializeIntoUndelegatedPackage(serializedPackage)
	if err != nil {
		pack, err = deserializeIntoRewardPackage(serializedPackage)
		if err != nil {
			return nil, err
		}
	}
	return pack, nil
}
