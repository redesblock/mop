package debugapi

import (
	"net/http"

	"github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/jsonhttp"
)

type addressesResponse struct {
	Addresses []multiaddr.Multiaddr `json:"addresses"`
}

func (s *server) addressesHandler(w http.ResponseWriter, r *http.Request) {
	addresses, err := s.P2P.Addresses()
	if err != nil {
		s.Logger.Debugf("debug api: p2p addresses: %v", err)
		jsonhttp.InternalServerError(w, err)
		return
	}
	jsonhttp.OK(w, addressesResponse{
		Addresses: addresses,
	})
}
