package bsc

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeader_UnmarshalJSON(t *testing.T) {
	chainID := big.NewInt(56)
	h := &Header{}
	jsonStr := `{"parentHash":"0xa9c482b74a276389681eabff076b19bef53cae9b5e44f02224e70e3bfc4e9142",
				"sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
				"miner":"0x72b61c6014342d914470ec7ac2975be345796c2b",
				"stateRoot":"0xacd5bca0bc33ed07cb35a635fa674e4ce06211ba201500564dd14dcdaf53e5a9",
				"transactionsRoot":"0xcb374b870584bd587dec2e82af38924a5c0d5765913568434c50448856f97a2c",
				"receiptsRoot":"0x6f2dc6ade8cf9422d62abd8e8386f20576a6b1f31cc00aa83dba140d183cff16",
				"logsBloom":"0xfcfef2ce9d78d29fcbfefbff9dfdf3afb23eff54be7efe7ffdb47abcffdff35fff7bf5de80e257fc836eb97fab43eef33dc7b7ffdfb7fcebfb3df6efff7fecfef473d6fe31dcbcebf7ffeff957b6fbbc6d1efb7ebfddbd3bddfeffded78df79ef3ff9ffd4fd67fcfdfdfff3fecdeddefddadf6ef802b5feaf6f4debffbd7ff9b1fcffff73ceb76fddcfd5f73ffff61bed4be7fb59fe77baebb746f4bcefbdebb76df77fdfb8b73bd2ffcf763b33ff7a6cfeefd7e36f6ed275ffa7fff7fbb996ff33bfbdfe76f23bfecf1ffcfceff3fefbd57b5f5dbfd7fde75cffff77ffffa7feefdf7ddef66f7db77fffd47efa6e5bf55f7fef3ebfbbdf3b1fe77f93ffacedf",
				"difficulty":"0x2",
				"number":"0x161d4e8",
				"gasLimit":"0x7355c0c",
				"gasUsed":"0x2134d60",
				"timestamp":"0x6378bdd7",
				"extraData":"0xd883010111846765746888676f312e31392e32856c696e757800000040fc9c67c61ee4a053e5ec524393cb2608e7a1b0de9a91f880095cd7bfc009b8e0ab5de96c2085a484239f38ca437bc805d2df9b88ea2a9a2a1ce9f5ab6005ae3464b17601",
				"mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000",
				"nonce":"0x0000000000000000",
				"baseFeePerGas":null,
				"hash":"0x8b6eeece6cedbb23038e7e5c2ce647fbdffa04972247d60a7564e81897e8bc30"}`
	err := h.UnmarshalJSON([]byte(jsonStr))
	require.NoError(t, err)

	signature, err := h.GetSignature()
	require.NoError(t, err)
	require.Equal(t, "c61ee4a053e5ec524393cb2608e7a1b0de9a91f880095cd7bfc009b8e0ab5de96c2085a484239f38ca437bc805d2df9b88ea2a9a2a1ce9f5ab6005ae3464b17601", hex.EncodeToString(signature))

	signer, err := h.ExtractSignerFromHeader(chainID)
	require.NoError(t, err)
	require.Equal(t, "0x72b61c6014342d914470eC7aC2975bE345796c2b", signer.String())
}
