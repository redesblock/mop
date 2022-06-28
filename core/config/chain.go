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
	testnetContractAddress = common.HexToAddress("0x8f60eba83B74EC2cf5d2F460DEecA8345a638878")
	mainnetContractAddress = common.HexToAddress("0x0FDc5429C50e2a39066D8A94F3e2D2476fcc3b85")
	// swap factory address
	testnetFactoryAddress = common.HexToAddress("0x46934d6027cd3b849Dc94b1947a37a4cA6950b3F")
	mainnetFactoryAddress = common.HexToAddress("0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420")
	// postage stamp
	testnetPostageStampContractAddress = common.HexToAddress("0xB314052ACd38A66fBDa2a1D43f3AA593c7dd5e24")
	mainnetPostageStampContractAddress = common.HexToAddress("0x6a1a21eca3ab28be85c7ba22b2d6eae5907c900e")
	// pledge
	testnetPledgeContractAddress = common.HexToAddress("0x732abD24e1017f46559e3325600E0A08160627A7")
	mainnetPledgeContractAddress = common.HexToAddress("0x6a1a21eca3ab28be85c7ba22b2d6eae5907c900e")
	// reward
	testnetRewardContractAddress = common.HexToAddress("0x9c209Ce68Ccc900EDdFB2626bcD8Ea8C4C7F4BB8")
	mainnetRewardContractAddress = common.HexToAddress("0x6a1a21eca3ab28be85c7ba22b2d6eae5907c900e")
)

type ChainConfig struct {
	StartBlock         uint64
	LegacyFactories    []common.Address
	PostageStamp       common.Address
	CurrentFactory     common.Address
	PriceOracleAddress common.Address
	PledgeAddress      common.Address
	RewardAddress      common.Address
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
		cfg.RewardAddress = testnetRewardContractAddress
		return &cfg, true
	case mainnetChainID:
		cfg.PostageStamp = mainnetPostageStampContractAddress
		cfg.StartBlock = mainnetStartBlock
		cfg.CurrentFactory = mainnetFactoryAddress
		cfg.LegacyFactories = []common.Address{}
		cfg.PriceOracleAddress = mainnetContractAddress
		cfg.PledgeAddress = mainnetPledgeContractAddress
		cfg.RewardAddress = mainnetRewardContractAddress
		return &cfg, true
	default:
		return &cfg, false
	}
}
