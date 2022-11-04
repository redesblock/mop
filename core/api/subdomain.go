package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/tracer"
)

func (s *Service) subdomainHandler(w http.ResponseWriter, r *http.Request) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)

	nameOrHex := mux.Vars(r)["subdomain"]
	pathVar := mux.Vars(r)["path"]
	if strings.HasSuffix(pathVar, "/") {
		pathVar = strings.TrimRight(pathVar, "/")
		// NOTE: leave one slash if there was some
		pathVar += "/"
	}

	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		logger.Debug("subdomain get: parse address string failed", "string", nameOrHex, "error", err)
		logger.Error(nil, "subdomain get: parse address string failed")
		jsonhttp.NotFound(w, nil)
		return
	}

	s.serveReference(address, pathVar, w, r)
}
