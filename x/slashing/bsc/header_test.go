package bsc

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHeader_UnmarshalJSON(t *testing.T) {
	h := &Header{}
	jsonStr := `{"difficulty":"0x2",
        "extraData":"0xd98301090a846765746889676f312e31322e3137856c696e75780000000000005b28385ac3a02a84c391c5d90b3aa0a62365136d80892c5e6158797b394f436c70c697f8d44d6dcfa49d4871897d2ff132356496f066d0adf28ddb3b7099ff5d00",
        "gasLimit":"0x2625a00",
        "gasUsed":"0xd752e8",
        "hash":"0x14c62182b7138b45c400afccbebda3c68a78ca7a6100d3e1fe9e1e8e71ef2b66",
        "logsBloom":"0x984e3983604a6120412617213b05004281984c60680c0cd0070200181d240016810500880001002b00055a0470461c09201145010c4e0729408998810a8800400a280247010848590804014800800454317020c88041a40248321010229290e68000011a808068002a56e2c41087114422a9841921a24ea709000430069010315e2c120124080610200cc18c137a0021206170070de90887502148600090002a50440040800905028a640401210c004089c20c4000e44054100cd00907642b4040900a920126454158b64181002d10014e80820508201c0480492880008c84102b0614804474840400413818940819042084c0042a002040498000c201309400",
        "miner":"0xfffffffffffffffffffffffffffffffffffffffe",
        "mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000",
        "nonce":"0x0000000000000000",
        "number":"0x1b4",
        "parentHash":"0xbfbb0f930378e623c27c1b6888694abd63926581697cf70a268a7455497e1011",
        "receiptsRoot":"0x32a9e85c5b51c5b99ce76dac6d1a75dd603bd7406b36d62cf8e74475c2be7462",
        "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
        "size":"0xd73c",
        "stateRoot":"0xa20ff6190c7d8a7993a50a3d60bdeca460c19128ba363413f5bff57670ddccb1",
        "timestamp":"0x5e79a878",
        "totalDifficulty":"0x365",
		"transactionsRoot":"0x8f192c648a6d9035adbf72a55cab5652e3c0d7549c378be45bd3d5248d4b3ac5"
		}`
	err := h.UnmarshalJSON([]byte(jsonStr))
	require.NoError(t, err)

	signature, err := h.GetSignature()
	require.NoError(t, err)
	require.Equal(t, "5b28385ac3a02a84c391c5d90b3aa0a62365136d80892c5e6158797b394f436c70c697f8d44d6dcfa49d4871897d2ff132356496f066d0adf28ddb3b7099ff5d00", hex.EncodeToString(signature))

	signer, err := h.ExtractSignerFromHeader()
	require.NoError(t, err)
	require.Equal(t, "0xB12fA6F899a16C156B67dBcb124d3733E72A164E", signer.String())
}
