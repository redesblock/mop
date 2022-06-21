package debugapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/redesblock/hop/core/jsonhttp"
)

func (s *Service) topologyHandler(w http.ResponseWriter, r *http.Request) {
	ms, ok := s.topologyDriver.(json.Marshaler)
	if !ok {
		s.logger.Error("topology driver cast to json marshaler")
		jsonhttp.InternalServerError(w, "topology json marshal interface error")
		return
	}

	b, err := ms.MarshalJSON()
	if err != nil {
		s.logger.Errorf("topology marshal to json: %v", err)
		jsonhttp.InternalServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", jsonhttp.DefaultContentTypeHeader)
	_, _ = io.Copy(w, bytes.NewBuffer(b))
}
