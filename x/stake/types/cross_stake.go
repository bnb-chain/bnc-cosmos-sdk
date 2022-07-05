package types

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/bsc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CrossStakePackageType uint8

const (
	SmartChainAddressLength = 20

	CrossStakeChannelID sdk.ChannelID = 16
	CrossStakeChannel                 = "crossStake"

	CrossStakeTypeDelegate         CrossStakePackageType = 1
	CrossStakeTypeUndelegate       CrossStakePackageType = 2
	CrossStakeTypeRedelegate       CrossStakePackageType = 3
	CrossStakeTypeClaimReward      CrossStakePackageType = 4
	CrossStakeTypeClaimUndelegated CrossStakePackageType = 5
)

type CrossStakeTransferOutRewardSynPackage struct {
	EventCode   CrossStakePackageType
	Amounts     []*big.Int
	Recipients  []SmartChainAddress
	RefundAddrs []sdk.AccAddress
	ExpireTime  int64
}

type CrossStakeTransferOutUndelegatedSynPackage struct {
	EventCode  CrossStakePackageType
	Amount     *big.Int
	Recipient  SmartChainAddress
	RefundAddr sdk.AccAddress
	ExpireTime int64
}

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

func GetStakeCAoB(sourceAddr []byte, salt string) (sdk.AccAddress, error) {
	saltBytes, err := hex.DecodeString("Staking" + salt + "Anchor")
	if err != nil {
		return nil, err
	}
	saltSha := bsc.Keccak256(saltBytes)
	accountBytes := make([]byte, len(sourceAddr))
	for i := 0; i < len(sourceAddr); i++ {
		accountBytes[i] = saltSha[i] ^ sourceAddr[i]
	}
	account, err := sdk.AccAddressFromHex(hex.EncodeToString(accountBytes))
	if err != nil {
		return nil, err
	}
	return account, nil
}
