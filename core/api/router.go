package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/auth"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/jsonhttp"
	"github.com/redesblock/mop/core/logging/httpaccess"
	"github.com/sirupsen/logrus"
	"resenje.org/web"
)

func (s *server) setupRouting() {
	const (
		apiVersion = "v1" // Only one api version exists, this should be configurable with more.
		rootPath   = "/" + apiVersion
	)

	router := mux.NewRouter()

	// handle is a helper closure which simplifies the router setup.
	handle := func(path string, handler http.Handler) {
		if s.Restricted {
			handler = web.ChainHandlers(auth.PermissionCheckHandler(s.auth), web.FinalHandler(handler))
		}
		router.Handle(path, handler)
		router.Handle(rootPath+path, handler)
	}

	router.NotFoundHandler = http.HandlerFunc(jsonhttp.NotFoundHandler)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "MOP Flock")
	})

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	if s.Restricted {
		router.Handle("/auth", jsonhttp.MethodHandler{
			"POST": web.ChainHandlers(
				s.newTracingHandler("auth"),
				jsonhttp.NewMaxBodyBytesHandler(512),
				web.FinalHandlerFunc(s.authHandler),
			),
		})
		router.Handle("/refresh", jsonhttp.MethodHandler{
			"POST": web.ChainHandlers(
				s.newTracingHandler("auth"),
				jsonhttp.NewMaxBodyBytesHandler(512),
				web.FinalHandlerFunc(s.refreshHandler),
			),
		})
	}

	handle("/bytes", jsonhttp.MethodHandler{
		"POST": web.ChainHandlers(
			s.contentLengthMetricMiddleware(),
			s.newTracingHandler("bytes-upload"),
			web.FinalHandlerFunc(s.bytesUploadHandler),
		),
	})
	handle("/bytes/{address}", jsonhttp.MethodHandler{
		"GET": web.ChainHandlers(
			s.contentLengthMetricMiddleware(),
			s.newTracingHandler("bytes-download"),
			web.FinalHandlerFunc(s.bytesGetHandler),
		),
	})

	handle("/chunks", jsonhttp.MethodHandler{
		"POST": web.ChainHandlers(
			jsonhttp.NewMaxBodyBytesHandler(flock.ChunkWithSpanSize),
			web.FinalHandlerFunc(s.chunkUploadHandler),
		),
	})

	handle("/chunks/stream", web.ChainHandlers(
		s.newTracingHandler("chunks-stream-upload"),
		web.FinalHandlerFunc(s.chunkUploadStreamHandler),
	))

	handle("/chunks/{addr}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.chunkGetHandler),
	})

	handle("/soc/{owner}/{id}", jsonhttp.MethodHandler{
		"POST": web.ChainHandlers(
			jsonhttp.NewMaxBodyBytesHandler(flock.ChunkWithSpanSize),
			web.FinalHandlerFunc(s.socUploadHandler),
		),
	})

	handle("/feeds/{owner}/{topic}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.feedGetHandler),
		"POST": web.ChainHandlers(
			jsonhttp.NewMaxBodyBytesHandler(flock.ChunkWithSpanSize),
			web.FinalHandlerFunc(s.feedPostHandler),
		),
	})

	handle("/mop", jsonhttp.MethodHandler{
		"POST": web.ChainHandlers(
			s.contentLengthMetricMiddleware(),
			s.newTracingHandler("mop-upload"),
			web.FinalHandlerFunc(s.mopUploadHandler),
		),
	})
	handle("/mop/{address}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.URL
		u.Path += "/"
		http.Redirect(w, r, u.String(), http.StatusPermanentRedirect)
	}))
	handle("/mop/{address}/{path:.*}", jsonhttp.MethodHandler{
		"GET": web.ChainHandlers(
			s.contentLengthMetricMiddleware(),
			s.newTracingHandler("mop-download"),
			web.FinalHandlerFunc(s.mopDownloadHandler),
		),
		"PATCH": web.ChainHandlers(
			s.newTracingHandler("mop-patch"),
			web.FinalHandlerFunc(s.mopPatchHandler),
		),
	})

	handle("/pss/send/{topic}/{targets}", web.ChainHandlers(
		s.gatewayModeForbidEndpointHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"POST": web.ChainHandlers(
				jsonhttp.NewMaxBodyBytesHandler(flock.ChunkSize),
				web.FinalHandlerFunc(s.pssPostHandler),
			),
		})),
	)

	handle("/pss/subscribe/{topic}", web.ChainHandlers(
		s.gatewayModeForbidEndpointHandler,
		web.FinalHandlerFunc(s.pssWsHandler),
	))

	handle("/tags", web.ChainHandlers(
		s.gatewayModeForbidEndpointHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.listTagsHandler),
			"POST": web.ChainHandlers(
				jsonhttp.NewMaxBodyBytesHandler(1024),
				web.FinalHandlerFunc(s.createTagHandler),
			),
		})),
	)
	handle("/tags/{id}", web.ChainHandlers(
		s.gatewayModeForbidEndpointHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET":    http.HandlerFunc(s.getTagHandler),
			"DELETE": http.HandlerFunc(s.deleteTagHandler),
			"PATCH": web.ChainHandlers(
				jsonhttp.NewMaxBodyBytesHandler(1024),
				web.FinalHandlerFunc(s.doneSplitHandler),
			),
		})),
	)

	handle("/pins", web.ChainHandlers(
		s.gatewayModeForbidEndpointHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.listPinnedRootHashes),
		})),
	)
	handle("/pins/{reference}", web.ChainHandlers(
		s.gatewayModeForbidEndpointHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET":    http.HandlerFunc(s.getPinnedRootHash),
			"POST":   http.HandlerFunc(s.pinRootHash),
			"DELETE": http.HandlerFunc(s.unpinRootHash),
		})),
	)

	handle("/stewardship/{address}", jsonhttp.MethodHandler{
		"GET": web.ChainHandlers(
			s.gatewayModeForbidEndpointHandler,
			web.FinalHandlerFunc(s.stewardshipGetHandler),
		),
		"PUT": web.ChainHandlers(
			s.gatewayModeForbidEndpointHandler,
			web.FinalHandlerFunc(s.stewardshipPutHandler),
		),
	})

	s.Handler = web.ChainHandlers(
		httpaccess.NewHTTPAccessLogHandler(s.logger, logrus.InfoLevel, s.tracer, "api access"),
		handlers.CompressHandler,
		// todo: add recovery handler
		s.responseCodeMetricsHandler,
		s.pageviewMetricsHandler,
		func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if o := r.Header.Get("Origin"); o != "" && s.checkOrigin(r) {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Origin", o)
					w.Header().Set("Access-Control-Allow-Headers", "User-Agent, Origin, Accept, Authorization, Content-Type, X-Requested-With, Access-Control-Request-Headers, Access-Control-Request-Method, Flock-Tag, Flock-Pin, Flock-Encrypt, Flock-Index-Document, Flock-Error-Document, Flock-Collection, Flock-Postage-Batch-Id, Gas-Price")
					w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, POST, PUT, DELETE")
					w.Header().Set("Access-Control-Max-Age", "3600")
				}
				h.ServeHTTP(w, r)
			})
		},
		s.gatewayModeForbidHeadersHandler,
		web.FinalHandler(router),
	)
}

func (s *server) gatewayModeForbidEndpointHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.GatewayMode {
			s.logger.Tracef("gateway mode: forbidden %s", r.URL.String())
			jsonhttp.Forbidden(w, nil)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (s *server) gatewayModeForbidHeadersHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.GatewayMode {
			if strings.ToLower(r.Header.Get(FlockPinHeader)) == "true" {
				s.logger.Tracef("gateway mode: forbidden pinning %s", r.URL.String())
				jsonhttp.Forbidden(w, "pinning is disabled")
				return
			}
			if strings.ToLower(r.Header.Get(FlockEncryptHeader)) == "true" {
				s.logger.Tracef("gateway mode: forbidden encryption %s", r.URL.String())
				jsonhttp.Forbidden(w, "encryption is disabled")
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
