package types

import (
	"crypto/rand"
	"encoding/binary"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func TestIbcValidator_Serialize(t *testing.T) {
	consAddr := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	feeAddr := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13}
	distAddr := sdk.AccAddress(tmhash.SumTruncated([]byte("dist")))
	v := IbcValidator{
		ConsAddr: consAddr,
		FeeAddr:  feeAddr,
		DistAddr: distAddr,
		Power:    1000000,
	}
	bz, err := v.Serialize()
	require.NoError(t, err)
	require.Equal(t, 68, len(bz))
	require.Equal(t, consAddr, bz[:20])
	require.Equal(t, feeAddr, bz[20:40])
	require.Equal(t, distAddr.Bytes(), bz[40:60])
	require.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x0F, 0x42, 0x40}, bz[60:])
}

func TestIbcValidatorSet_Serialize(t *testing.T) {
	ibcValidatorSet := make(IbcValidatorSet, 11)
	for i := range ibcValidatorSet {
		ibcValidatorSet[i] = IbcValidator{
			ConsAddr:randAddr(t, 20),
			FeeAddr: randAddr(t, 20),
			DistAddr: sdk.AccAddress(randAddr(t, 20)),
			Power: int64((i+1)*1000),
		}
	}
	bz, err := ibcValidatorSet.Serialize()
	require.NoError(t, err)
	require.Equal(t, 748,len(bz))
	for i:= range ibcValidatorSet {
		require.Equal(t, ibcValidatorSet[i].ConsAddr, bz[i*68:i*68+20])
		require.Equal(t, ibcValidatorSet[i].FeeAddr, bz[i*68+20:i*68+40])
		require.Equal(t, ibcValidatorSet[i].DistAddr.Bytes(), bz[i*68+40:i*68+60])
		require.Equal(t, ibcValidatorSet[i].Power, int64(binary.BigEndian.Uint64(bz[i*68+60:(i+1)*68])))
	}
}

func randAddr(t *testing.T, size int64) []byte {
	addr := make([]byte, size)
	n, err := rand.Read(addr)
	require.NoError(t, err)
	require.Equal(t, 20, n)
	require.Equal(t, 20, len(addr))
	return addr
}