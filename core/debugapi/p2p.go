package debugapi

import (
	"encoding/hex"
	"net/http"

	"github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/crypto"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/swarm"
)

type addressesResponse struct {
	Overlay   swarm.Address         `json:"overlay"`
	Underlay  []multiaddr.Multiaddr `json:"underlay"`
	PublicKey string                `json:"public_key"`
}

func (s *server) addressesHandler(w http.ResponseWriter, r *http.Request) {
	underlay, err := s.P2P.Addresses()
	if err != nil {
		s.Logger.Debugf("debug api: p2p addresses: %v", err)
		jsonhttp.InternalServerError(w, err)
		return
	}
	jsonhttp.OK(w, addressesResponse{
		Overlay:   s.Overlay,
		Underlay:  underlay,
		PublicKey: hex.EncodeToString(crypto.EncodeSecp256k1PublicKey(&s.PublicKey)),
	})
}
