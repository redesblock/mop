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
		fmt.Fprintln(w, "hop node")
	})

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	router.Handle("/pingpong/{peer-id}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.pingpongHandler),
	})

	router.Handle("/hop", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.hopUploadHandler),
	})

	router.Handle("/hop/{address}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.hopGetHandler),
	})

	router.Handle("/hop-chunk/{addr}", jsonhttp.MethodHandler{
		"GET":  http.HandlerFunc(s.chunkGetHandler),
		"POST": http.HandlerFunc(s.chunkUploadHandler),
	})

	router.Handle("/hop-tag/name/{name}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.CreateTag),
	})

	router.Handle("/hop-tag/addr/{addr}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.getTagInfoUsingAddress),
	})

	router.Handle("/hop-tag/uuid/{uuid}", jsonhttp.MethodHandler{
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
