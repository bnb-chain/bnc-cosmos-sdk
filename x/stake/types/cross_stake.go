package types

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/bsc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	SmartChainAddressLength = 20

	CrossStakeChannel              = "crossStake"
	TagRelayerFee                  = "relayerFee"
	TagCrossChainStakeClaimReward  = "crossCHainStakeClaimReward_%s_%s"
	TagCrossChainStakeClaimUnstake = "crossCHainStakeClaimUnstake_%s_%s"

	CrossStakeChannelID sdk.ChannelID = 16
)

// SmartChainAddress defines a standard smart chain address
type SmartChainAddress [SmartChainAddressLength]byte

// NewSmartChainAddress is a constructor function for SmartChainAddress
func NewSmartChainAddress(addr string) (SmartChainAddress, error) {
	addr = strings.ToLower(addr)
	if len(addr) >= 2 && addr[:2] == "0x" {
		addr = addr[2:]
	}
	if length := len(addr); length != 2*SmartChainAddressLength {
		return SmartChainAddress{}, fmt.Errorf("invalid address hex length: %v != %v", length, 2*SmartChainAddressLength)
	}

	bin, err := hex.DecodeString(addr)
	if err != nil {
		return SmartChainAddress{}, err
	}
	var address SmartChainAddress
	address.SetBytes(bin)
	return address, nil
}

func (addr *SmartChainAddress) SetBytes(b []byte) {
	if len(b) > len(addr) {
		b = b[len(b)-20:]
	}
	copy(addr[20-len(b):], b)
}

func (addr SmartChainAddress) IsEmpty() bool {
	addrValue := big.NewInt(0)
	addrValue.SetBytes(addr[:])

	return addrValue.Cmp(big.NewInt(0)) == 0
}

// Route should return the name of the module
func (addr SmartChainAddress) String() string {
	return sdk.HexAddress(addr[:])
}

// MarshalJSON marshals the smart chain address to JSON
func (addr SmartChainAddress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", addr.String())), nil
}

// UnmarshalJSON unmarshals an smart chain address
func (addr *SmartChainAddress) UnmarshalJSON(input []byte) error {
	hexBytes, err := sdk.HexDecode(string(input[1 : len(input)-1]))
	if err != nil {
		return err
	}
	addr.SetBytes(hexBytes)
	return nil
}

func GetStakeCAoB(delAddr []byte, salt string) (sdk.AccAddress, error) {
	saltBytes, err := hex.DecodeString("ccceStaking" + salt + "Address")
	if err != nil {
		return nil, err
	}
	saltSha := bsc.Keccak256(saltBytes)
	accountBytes := make([]byte, len(delAddr))
	for i := 0; i < len(delAddr); i++ {
		accountBytes[i] = saltSha[i] ^ delAddr[i]
	}
	account, err := sdk.AccAddressFromHex(hex.EncodeToString(accountBytes))
	if err != nil {
		return nil, err
	}
	return account, nil
}
