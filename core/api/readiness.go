package api

import "net/http"

func (s *Service) readinessHandler(w http.ResponseWriter, r *http.Request) {
	if s.probe.Ready() == ProbeStatusOK {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
