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
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(jsonhttp.NotFoundHandler)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Swarm Node")
	})

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	router.Handle("/raw", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.rawUploadHandler),
	})

	router.Handle("/raw/{address}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.rawGetHandler),
	})

	router.Handle("/chunk/{addr}", jsonhttp.MethodHandler{
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
