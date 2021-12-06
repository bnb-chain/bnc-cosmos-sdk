package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Sharer struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
}

// reward model before bgc upgrade
type PreReward struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
	Amount  int64
}

type Reward struct {
	Validator sdk.ValAddress // Validator will be published for downstreams
	AccAddr   sdk.AccAddress
	Tokens    sdk.Dec // Shares in PreReward does not publish
	Amount    int64
}

type StoredValDistAddr struct {
	Validator      sdk.ValAddress
	DistributeAddr sdk.AccAddress
}

func MustMarshalRewards(cdc *codec.Codec, rewards []Reward) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(rewards)
}

func MustUnmarshalRewards(cdc *codec.Codec, value []byte) (rewards []Reward) {
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
