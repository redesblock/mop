package debugapi

import (
	"encoding/hex"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiformats/go-multiaddr"
	"github.com/redesblock/mop/core/crypto"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/jsonhttp"
)

type addressesResponse struct {
	Overlay      *flock.Address        `json:"overlay"`
	Underlay     []multiaddr.Multiaddr `json:"underlay"`
	Ethereum     common.Address        `json:"bsc"`
	PublicKey    string                `json:"publicKey"`
	PSSPublicKey string                `json:"pssPublicKey"`
}

func (s *Service) addressesHandler(w http.ResponseWriter, r *http.Request) {
	// initialize variable to json encode as [] instead null if p2p is nil
	underlay := make([]multiaddr.Multiaddr, 0)
	// addresses endpoint is exposed before p2p service is configured
	// to provide information about other addresses.
	if s.p2p != nil {
		u, err := s.p2p.Addresses()
		if err != nil {
			s.logger.Debugf("debug api: p2p addresses: %v", err)
			jsonhttp.InternalServerError(w, err)
			return
		}
		underlay = u
	}
	jsonhttp.OK(w, addressesResponse{
		Overlay:      s.overlay,
		Underlay:     underlay,
		Ethereum:     s.ethereumAddress,
		PublicKey:    hex.EncodeToString(crypto.EncodeSecp256k1PublicKey(&s.publicKey)),
		PSSPublicKey: hex.EncodeToString(crypto.EncodeSecp256k1PublicKey(&s.pssPublicKey)),
	})
}
