package config

import (
	"github.com/ethereum/go-ethereum/common"
)

var (
	// chain ID
	testnetChainID = int64(97)
	mainnetChainID = int64(56)
	// start block
	// replace this with
	testnetStartBlock = uint64(20552968)
	mainnetStartBlock = uint64(19044111)
	// factory address
	testnetContractAddress = common.HexToAddress("0x8f60eba83B74EC2cf5d2F460DEecA8345a638878")
	mainnetContractAddress = common.HexToAddress("")
	testnetFactoryAddress  = common.HexToAddress("0x46934d6027cd3b849Dc94b1947a37a4cA6950b3F")
	mainnetFactoryAddress  = common.HexToAddress("")
	// voucher stamp
	testnetVoucherStampContractAddress = common.HexToAddress("0x179f367Cf345cE5fAB50D66E6b6F39C02dA47C85")
	mainnetVoucherStampContractAddress = common.HexToAddress("")
)

type ChainConfig struct {
	StartBlock         uint64
	LegacyFactories    []common.Address
	VoucherStamp       common.Address
	CurrentFactory     common.Address
	PriceOracleAddress common.Address
}

func GetChainConfig(chainID int64) (*ChainConfig, bool) {
	var cfg ChainConfig
	switch chainID {
	case testnetChainID:
		cfg.VoucherStamp = testnetVoucherStampContractAddress
		cfg.StartBlock = testnetStartBlock
		cfg.CurrentFactory = testnetFactoryAddress
		cfg.LegacyFactories = []common.Address{}
		cfg.PriceOracleAddress = testnetContractAddress
		return &cfg, true
	case mainnetChainID:
		cfg.VoucherStamp = mainnetVoucherStampContractAddress
		cfg.StartBlock = mainnetStartBlock
		cfg.CurrentFactory = mainnetFactoryAddress
		cfg.LegacyFactories = []common.Address{}
		cfg.PriceOracleAddress = mainnetContractAddress
		return &cfg, true
	default:
		return &cfg, false
	}
}
