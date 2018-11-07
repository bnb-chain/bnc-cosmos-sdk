package types

type AccountStoreCache interface {
	GetAccount(addr AccAddress) interface{}
	SetAccount(addr AccAddress, acc interface{})
	Delete(addr AccAddress)
}

type AccountCache interface {
	AccountStoreCache

	Write()
}
