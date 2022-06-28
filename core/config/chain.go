package config

import (
	"github.com/ethereum/go-ethereum/common"
)

var (
	// chain ID
	testnetChainID = int64(97)
	mainnetChainID = int64(56)
	// start block
	testnetStartBlock = uint64(20552968)
	mainnetStartBlock = uint64(19044111)
	// price oracle
	testnetContractAddress = common.HexToAddress("0x8a78fc3ef8b75ff5b3983bdc37b01278e3eaaa01")
	mainnetContractAddress = common.HexToAddress("0x0FDc5429C50e2a39066D8A94F3e2D2476fcc3b85")
	// swap factory address
	testnetFactoryAddress = common.HexToAddress("0x780594f00f4eb0d6dd55b6063a1cb294085e3893")
	mainnetFactoryAddress = common.HexToAddress("0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420")
	// postage stamp
	testnetPostageStampContractAddress = common.HexToAddress("0xb06d1bcaaeee431db99c7e4b4ff2e4313b6d38cc")
	mainnetPostageStampContractAddress = common.HexToAddress("0x6a1a21eca3ab28be85c7ba22b2d6eae5907c900e")
	// pledge
	testnetPledgeContractAddress = common.HexToAddress("0x61846DA3318FdE90426F2D46F2540CfaF915815A")
	mainnetPledgeContractAddress = common.HexToAddress("0x6a1a21eca3ab28be85c7ba22b2d6eae5907c900e")
)

type ChainConfig struct {
	StartBlock         uint64
	LegacyFactories    []common.Address
	PostageStamp       common.Address
	CurrentFactory     common.Address
	PriceOracleAddress common.Address
	PledgeAddress      common.Address
}

func GetChainConfig(chainID int64) (*ChainConfig, bool) {
	var cfg ChainConfig
	switch chainID {
	case testnetChainID:
		cfg.PostageStamp = testnetPostageStampContractAddress
		cfg.StartBlock = testnetStartBlock
		cfg.CurrentFactory = testnetFactoryAddress
		cfg.LegacyFactories = []common.Address{}
		cfg.PriceOracleAddress = testnetContractAddress
		cfg.PledgeAddress = testnetPledgeContractAddress
		return &cfg, true
	case mainnetChainID:
		cfg.PostageStamp = mainnetPostageStampContractAddress
		cfg.StartBlock = mainnetStartBlock
		cfg.CurrentFactory = mainnetFactoryAddress
		cfg.LegacyFactories = []common.Address{}
		cfg.PriceOracleAddress = mainnetContractAddress
		cfg.PledgeAddress = mainnetPledgeContractAddress
		return &cfg, true
	default:
		return &cfg, false
	}
}
