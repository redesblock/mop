package api

import "net/http"

type handlerJob struct {
	w          http.ResponseWriter
	r          *http.Request
	done       chan bool
	handleFunc func(w http.ResponseWriter, r *http.Request)
}

func (h *handlerJob) Do() error {
	h.handleFunc(h.w, h.r)
	h.done <- true
	return nil
}
