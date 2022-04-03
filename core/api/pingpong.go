package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/p2p"
	"github.com/redesblock/hop/core/swarm"
)

type pingpongResponse struct {
	RTT string `json:"rtt"`
}

func (s *server) pingpongHandler(w http.ResponseWriter, r *http.Request) {
	peerID := mux.Vars(r)["peer-id"]
	ctx := r.Context()

	span, logger, ctx := s.Tracer.StartSpanFromContext(ctx, "pingpong-api", s.Logger)
	defer span.Finish()

	address, err := swarm.ParseHexAddress(peerID)
	if err != nil {
		logger.Debugf("pingpong: parse peer address %s: %v", peerID, err)
		jsonhttp.BadRequest(w, "invalid peer address")
		return
	}

	rtt, err := s.Pingpong.Ping(ctx, address, "hey", "there", ",", "how are", "you", "?")
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
	s.metrics.PingRequestCount.Inc()

	logger.Infof("pingpong succeeded to peer %s", peerID)
	jsonhttp.OK(w, pingpongResponse{
		RTT: rtt.String(),
	})
}
