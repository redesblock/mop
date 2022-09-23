package debugapi

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/jsonhttp"
	"github.com/redesblock/mop/core/p2p"
	"github.com/redesblock/mop/core/swarm"
)

type pingpongResponse struct {
	RTT string `json:"rtt"`
}

func (s *Service) pingpongHandler(w http.ResponseWriter, r *http.Request) {
	peerID := mux.Vars(r)["peer-id"]
	ctx := r.Context()

	span, logger, ctx := s.tracer.StartSpanFromContext(ctx, "pingpong-api", s.logger)
	defer span.Finish()

	address, err := swarm.ParseHexAddress(peerID)
	if err != nil {
		logger.Debugf("pingpong: parse peer address %s: %v", peerID, err)
		jsonhttp.BadRequest(w, "invalid peer address")
		return
	}

	rtt, err := s.pingpong.Ping(ctx, address, "ping")
	if err != nil {
		logger.Debugf("pingpong: ping %s: %v", peerID, err)
		if errors.Is(err, p2p.ErrPeerNotFound) {
			jsonhttp.NotFound(w, "peer not found")
			return
		}

		logger.Errorf("pingpong failed to peer %s", peerID)
		jsonhttp.InternalServerError(w, nil)
		return
	}

	logger.Infof("pingpong succeeded to peer %s", peerID)
	jsonhttp.OK(w, pingpongResponse{
		RTT: rtt.String(),
	})
}
