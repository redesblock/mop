package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/redesblock/mop/core/api/jsonhttp"
)

func (s *Service) topologyHandler(w http.ResponseWriter, r *http.Request) {
	params := s.topologyDriver.Snapshot()

	params.LightNodes = s.lightNodes.PeerInfo()

	b, err := json.Marshal(params)
	if err != nil {
		s.logger.Error(err, "topology get: marshal to json failed")
		jsonhttp.InternalServerError(w, err)
		return
	}
	w.Header().Set(contentTypeHeader, jsonhttp.DefaultContentTypeHeader)
	_, _ = io.Copy(w, bytes.NewBuffer(b))
}
