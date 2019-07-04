package crypto

import (
	"github.com/spf13/viper"

	"github.com/ipfs/go-log"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	_ "github.com/binance-chain/tss-lib/ecdsa/signing"
	"github.com/binance-chain/tss/client"
	"github.com/binance-chain/tss/common"
)

func NewPrivKeyTss(home, passphrase string) (tmcrypto.PrivKey, error) {
	config, err := common.ReadConfigFromHome(viper.New(), home)
	if err != nil {
		return nil, err
	}
	config.Password = passphrase
	initLogLevel(config)
	return client.NewTssClient(config, false), nil
}

func initLogLevel(cfg common.TssConfig) {
	log.SetLogLevel("tss", cfg.P2PConfig.LogLevel)
	log.SetLogLevel("tss-lib", cfg.P2PConfig.LogLevel)
	log.SetLogLevel("srv", cfg.P2PConfig.LogLevel)
	log.SetLogLevel("trans", cfg.P2PConfig.LogLevel)
	log.SetLogLevel("p2p_utils", cfg.P2PConfig.LogLevel)

	// libp2p loggers
	log.SetLogLevel("dht", "error")
	log.SetLogLevel("discovery", "error")
	log.SetLogLevel("swarm2", "error")
}
