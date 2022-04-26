package debugapi

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/multiformats/go-multiaddr"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/swarm"
)

type peerConnectResponse struct {
	Address string `json:"address"`
}

func (s *server) peerConnectHandler(w http.ResponseWriter, r *http.Request) {
	addr, err := multiaddr.NewMultiaddr("/" + mux.Vars(r)["multi-address"])
	if err != nil {
		s.Logger.Debugf("debug api: peer connect: parse multiaddress: %v", err)
		jsonhttp.BadRequest(w, err)
		return
	}

	hopAddr, err := s.P2P.Connect(r.Context(), addr)
	if err != nil {
		s.Logger.Debugf("debug api: peer connect %s: %v", addr, err)
		s.Logger.Errorf("unable to connect to peer %s", addr)
		jsonhttp.InternalServerError(w, err)
		return
	}

	err = s.Addressbook.Put(hopAddr.Overlay, *hopAddr)
	if err != nil {
		s.Logger.Debugf("debug api: addressbook.put %s: %v", addr, err)
		s.Logger.Errorf("unable to persist peer %s", addr)
		jsonhttp.InternalServerError(w, err)
		return
	}
	if err := s.TopologyDriver.Connected(r.Context(), hopAddr.Overlay); err != nil {
		_ = s.P2P.Disconnect(hopAddr.Overlay)
		s.Logger.Debugf("debug api: topologyDriver.Connected %s: %v", addr, err)
		s.Logger.Errorf("unable to connect to peer %s", addr)
		jsonhttp.InternalServerError(w, err)
		return
	}

	jsonhttp.OK(w, peerConnectResponse{
		Address: hopAddr.Overlay.String(),
	})
}

func (s *server) peerDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["address"]
	swarmAddr, err := swarm.ParseHexAddress(addr)
	if err != nil {
		s.Logger.Debugf("debug api: parse peer address %s: %v", addr, err)
		jsonhttp.BadRequest(w, "invalid peer address")
		return
	}

	if err := s.P2P.Disconnect(swarmAddr); err != nil {
		s.Logger.Debugf("debug api: peer disconnect %s: %v", addr, err)
		if errors.Is(err, p2p.ErrPeerNotFound) {
			jsonhttp.BadRequest(w, "peer not found")
			return
		}
		s.Logger.Errorf("unable to disconnect peer %s", addr)
		jsonhttp.InternalServerError(w, err)
		return
	}

	jsonhttp.OK(w, nil)
}

type peersResponse struct {
	Peers []p2p.Peer `json:"peers"`
}

func (s *server) peersHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, peersResponse{
		Peers: s.P2P.Peers(),
	})
}
