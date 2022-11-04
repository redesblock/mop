package kademlia

var (
	TimeToRetry                 = &timeToRetry
	SaturationPeers             = &saturationPeers
	OverSaturationPeers         = &overSaturationPeers
	BootnodeOverSaturationPeers = &bootNodeOverSaturationPeers
	LowWaterMark                = &nnLowWatermark
	PruneOversaturatedBinsFunc  = func(k *Kad) func(uint8) {
		return k.pruneOversaturatedBins
	}
	GenerateCommonBinPrefixes = generateCommonBinPrefixes
	PeerPingPollTime          = &peerPingPollTime
	BitSuffixLength           = defaultBitSuffixLength
)

type PeerFilterFunc = peerFilterFunc
