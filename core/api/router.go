package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/logging"
	"github.com/sirupsen/logrus"
	"resenje.org/web"
)

func (s *server) setupRouting() {
	apiVersion := "v1" // only one api version exists, this should be configurable with more

	handle := func(router *mux.Router, path string, handler http.Handler) {
		router.Handle(path, handler)
		router.Handle("/"+apiVersion+path, handler)
	}

	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(jsonhttp.NotFoundHandler)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Swarm Node")
	})

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	handle(router, "/bytes", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.bytesUploadHandler),
	})
	handle(router, "/bytes/{address}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.bytesGetHandler),
	})

	handle(router, "/chunks/{addr}", jsonhttp.MethodHandler{
		"GET":  http.HandlerFunc(s.chunkGetHandler),
		"POST": http.HandlerFunc(s.chunkUploadHandler),
	})

	router.Handle("/tag/name/{name}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.CreateTag),
	})

	router.Handle("/tag/uuid/{uuid}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.getTagInfoUsingUUid),
	})

	s.Handler = web.ChainHandlers(
		logging.NewHTTPAccessLogHandler(s.Logger, logrus.InfoLevel, "api access"),
		handlers.CompressHandler,
		// todo: add recovery handler
		s.pageviewMetricsHandler,
		web.FinalHandler(router),
	)
}
