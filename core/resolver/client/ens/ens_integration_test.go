//go:build integration

package ens_test

import (
	"errors"
	"testing"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/resolver/client/ens"
)

func TestENSIntegration(t *testing.T) {
	// TODO: consider using a stable gateway instead of INFURA.
	defaultEndpoint := "https://bsc.infura.io/v3/59d83a5a4be74f86b9851190c802297b"
	defaultAddr := cluster.MustParseHexAddress("00cb23598c2e520b6a6aae3ddc94fed4435a2909690bdd709bf9d9e7c2aadfad")

	testCases := []struct {
		desc            string
		endpoint        string
		contractAddress string
		name            string
		wantAdr         cluster.Address
		wantErr         error
	}{
		// TODO: add a test targeting a resolver with an invalid contenthash
		// record.
		{
			desc:     "invalid resolver endpoint",
			endpoint: "example.com",
			wantErr:  ens.ErrFailedToConnect,
		},
		{
			desc:    "no domain",
			name:    "idonthaveadomain",
			wantErr: ens.ErrResolveFailed,
		},
		{
			desc:    "no eth domain",
			name:    "centralized.com",
			wantErr: ens.ErrResolveFailed,
		},
		{
			desc:    "not registered",
			name:    "unused.test.cluster.eth",
			wantErr: ens.ErrResolveFailed,
		},
		{
			desc:    "no content hash",
			name:    "nocontent.resolver.test.cluster.eth",
			wantErr: ens.ErrResolveFailed,
		},
		{
			desc:            "invalid contract address",
			contractAddress: "0xFFFFFFFF",
			name:            "example.resolver.test.cluster.eth",
			wantErr:         ens.ErrFailedToConnect,
		},
		{
			desc:    "ok",
			name:    "example.resolver.test.cluster.eth",
			wantAdr: defaultAddr,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			if tC.endpoint == "" {
				tC.endpoint = defaultEndpoint
			}

			ensClient, err := ens.NewClient(tC.endpoint, ens.WithContractAddress(tC.contractAddress))
			if err != nil {
				if !errors.Is(err, tC.wantErr) {
					t.Errorf("got %v, want %v", err, tC.wantErr)
				}
				return
			}
			defer ensClient.Close()

			addr, err := ensClient.Resolve(tC.name)
			if err != nil {
				if !errors.Is(err, tC.wantErr) {
					t.Errorf("got %v, want %v", err, tC.wantErr)
				}
				return
			}

			if !addr.Equal(defaultAddr) {
				t.Errorf("bad addr: got %s, want %s", addr, defaultAddr)
			}

			err = ensClient.Close()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
