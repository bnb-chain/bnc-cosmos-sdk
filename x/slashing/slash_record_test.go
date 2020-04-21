package slashing

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetGetSlashRecord(t *testing.T) {
	ctx, _, _, _, keeper := createTestInput(t, DefaultParams())
	sideConsAddr := randomSideConsAddr()
	keeper.setSlashRecord(ctx, sideConsAddr, 100)
	require.NotNil(t, keeper.getSlashRecord(ctx, sideConsAddr, 100))
}

func randomSideConsAddr() []byte {
	bz := make([]byte, 20)
	rand.Read(bz)
	return bz
}
