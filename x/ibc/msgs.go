package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IBCPackageMsg struct {
	Sender      sdk.AccAddress   `json:"sender"`
	DestChainID sdk.CrossChainID `json:"dest_chain_id"`
	ChannelID   sdk.ChannelID    `json:"channel_id"`
	Package     []byte           `json:"package"`
}

func NewIBCPackage(srcAddr sdk.AccAddress, destChainID sdk.CrossChainID, channelID sdk.ChannelID, Package []byte) IBCPackageMsg {

	return IBCPackageMsg{
		Sender:      srcAddr,
		DestChainID: destChainID,
		ChannelID:   channelID,
		Package:     Package,
	}
}

func (msg IBCPackageMsg) Route() string                { return "ibc" }
func (msg IBCPackageMsg) Type() string                 { return "IBCPackage" }
func (msg IBCPackageMsg) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }
func (msg IBCPackageMsg) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
func (msg IBCPackageMsg) ValidateBasic() sdk.Error {
	if len(msg.Package) == 0 {
		return ErrEmptyPackage(DefaultCodespace, "empty package")
	}
	return nil
}

func (msg IBCPackageMsg) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
