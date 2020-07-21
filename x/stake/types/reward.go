package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Sharer struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
}

type Reward struct {
	AccAddr sdk.AccAddress
	Shares  sdk.Dec
	Amount  int64
}
