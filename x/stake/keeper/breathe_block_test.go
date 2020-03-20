package keeper

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSetGetBreatheBlockHeight(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	height1 := int64(1000)
	height2 := int64(2000)
	now := time.Now()
	keeper.SetBreatheBlockHeight(ctx,height1,now)
	keeper.SetBreatheBlockHeight(ctx,height2,now.Add(time.Hour * 24))

	resHeight1,found := keeper.GetBreatheBlockHeight(ctx,1)
	require.True(t,found)
	require.Equal(t, height2, resHeight1)

	resHeight2,found := keeper.GetBreatheBlockHeight(ctx,2)
	require.True(t,found)
	require.Equal(t, height1, resHeight2)

	_,found = keeper.GetBreatheBlockHeight(ctx,3)
	require.False(t,found)
}
