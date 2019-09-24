module github.com/cosmos/cosmos-sdk

go 1.12

require (
	github.com/bartekn/go-bip39 v0.0.0-20171116152956-a05967ea095d
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.0.0-20190115013929-ed77733ec07d
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/golang-lru v0.5.0
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-isatty v0.0.6
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pelletier/go-toml v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/rakyll/statik v0.1.5
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/iavl v0.12.4
	github.com/tendermint/tendermint v0.32.3
	github.com/zondax/hid v0.9.0 // indirect
	github.com/zondax/ledger-cosmos-go v0.9.9
	github.com/zondax/ledger-go v0.9.0 // indirect
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
)

replace (
	github.com/cosmos/ledger-cosmos-go => github.com/binance-chain/ledger-cosmos-go v0.9.9-binance.2
	github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
	github.com/tendermint/iavl => github.com/binance-chain/bnc-tendermint-iavl v0.12.0-binance.1
	github.com/tendermint/tendermint => github.com/binance-chain/bnc-tendermint v0.29.1-binance.3.0.20190923114917-479a59a5dbd7
	golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
)
