package cross_stake

import (
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
)

var (
	NewKeeper = keeper.NewKeeper
)

type (
	Keeper = keeper.Keeper
)
