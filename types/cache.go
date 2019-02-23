package types

type AccountStoreCache interface {
	GetAccount(addr AccAddress) Account
	SetAccount(addr AccAddress, acc Account)
	Delete(addr AccAddress)
	ClearCache() // only used by state sync to clear genesis status of accounts
}

type AccountCache interface {
	AccountStoreCache

	Cache() AccountCache
	Write()
}
