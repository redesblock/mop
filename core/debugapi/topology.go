package debugapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/redesblock/hop/core/jsonhttp"
)

func (s *server) topologyHandler(w http.ResponseWriter, r *http.Request) {
	ms, ok := s.TopologyDriver.(json.Marshaler)
	if !ok {
		s.Logger.Error("topology driver cast to json marshaler")
		jsonhttp.InternalServerError(w, "topology json marshal interface error")
		return
	}

	b, err := ms.MarshalJSON()
	if err != nil {
		s.Logger.Errorf("topology marshal to json: %v", err)
		jsonhttp.InternalServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", jsonhttp.DefaultContentTypeHeader)
	_, _ = io.Copy(w, bytes.NewBuffer(b))
}
