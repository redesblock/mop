package cidv1

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/resolver"
)

// https://github.com/multiformats/multicodec/blob/master/table.csv
const (
	ClusterNsCodec       uint64 = 0xe4
	ClusterManifestCodec uint64 = 0xfa
	ClusterFeedCodec     uint64 = 0xfb
)

type Resolver struct{}

func (Resolver) Resolve(name string) (cluster.Address, error) {
	id, err := cid.Parse(name)
	if err != nil {
		return cluster.ZeroAddress, fmt.Errorf("failed parsing CID %s err %v: %w", name, err, resolver.ErrParse)
	}

	switch id.Prefix().GetCodec() {
	case ClusterNsCodec:
	case ClusterManifestCodec:
	case ClusterFeedCodec:
	default:
		return cluster.ZeroAddress, fmt.Errorf("unsupported codec for CID %d: %w", id.Prefix().GetCodec(), resolver.ErrParse)
	}

	dh, err := multihash.Decode(id.Hash())
	if err != nil {
		return cluster.ZeroAddress, fmt.Errorf("unable to decode hash %v: %w", err, resolver.ErrInvalidContentHash)
	}

	addr, err := cluster.NewAddress(dh.Digest), nil
	if err != nil {
		return cluster.ZeroAddress, fmt.Errorf("unable to parse digest hash %v: %w", err, resolver.ErrInvalidContentHash)
	}

	return addr, nil
}

func (Resolver) Close() error {
	return nil
}
