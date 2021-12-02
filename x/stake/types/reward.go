package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Sharer struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
}

type Reward struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
	Amount  int64
}

type StoredReward struct {
	Validator sdk.ValAddress
	AccAddr   sdk.AccAddress
	Shares    sdk.Dec
	Amount    int64
}

type StoredValDistAddr struct {
	Validator      sdk.ValAddress
	DistributeAddr sdk.AccAddress
}

func MustMarshalRewards(cdc *codec.Codec, rewards []StoredReward) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(rewards)
}

func MustUnmarshalRewards(cdc *codec.Codec, value []byte) (rewards []StoredReward) {
	err := cdc.UnmarshalBinaryLengthPrefixed(value, &rewards)
	if err != nil {
		panic(err)
	}
	return rewards
}

func MustMarshalValDistAddrs(cdc *codec.Codec, valDistAddrs []StoredValDistAddr) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(valDistAddrs)
}

func MustUnmarshalValDistAddrs(cdc *codec.Codec, value []byte) (valDistAddrs []StoredValDistAddr) {
	err := cdc.UnmarshalBinaryLengthPrefixed(value, &valDistAddrs)
	if err != nil {
		panic(err)
	}
	return valDistAddrs
}
