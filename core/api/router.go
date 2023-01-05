package api

import (
	"crypto/tls"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redesblock/mop/core/api/auth"
	"github.com/redesblock/mop/core/api/httpaccess"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/cluster"
	"resenje.org/web"
)

const (
	apiVersion = "v1" // Only one api version exists, this should be configurable with more.
	rootPath   = "/" + apiVersion
)

func (s *Service) MountTechnicalDebug() {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(jsonhttp.NotFoundHandler)
	s.router = router

	s.mountTechnicalDebug()

	s.Handler = web.ChainHandlers(
		httpaccess.NewHTTPAccessLogHandler(s.logger, s.tracer, "debug api access", nil),
		handlers.CompressHandler,
		s.corsHandler,
		web.NoCacheHeadersHandler,
		web.FinalHandler(router),
	)
}

func (s *Service) MountDebug(restricted bool) {
	s.mountBusinessDebug(restricted)

	s.Handler = web.ChainHandlers(
		httpaccess.NewHTTPAccessLogHandler(s.logger, s.tracer, "debug api access", nil),
		handlers.CompressHandler,
		s.corsHandler,
		web.NoCacheHeadersHandler,
		web.FinalHandler(s.router),
	)
}

func (s *Service) MountAPI() {
	if s.router == nil {
		s.router = mux.NewRouter()
		s.router.NotFoundHandler = http.HandlerFunc(jsonhttp.NotFoundHandler)
	}

	s.mountAPI()

	skipHeadHandler := func(fn func(http.Handler) http.Handler) func(h http.Handler) http.Handler {
		return func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodHead {
					h.ServeHTTP(w, r)
				} else {
					fn(h).ServeHTTP(w, r)
				}
			})
		}
	}

	s.Handler = web.ChainHandlers(
		httpaccess.NewHTTPAccessLogHandler(s.logger, s.tracer, "api access", s.trafficHandler),
		skipHeadHandler(handlers.CompressHandler),
		s.responseCodeMetricsHandler,
		s.pageviewMetricsHandler,
		s.corsHandler,
		web.FinalHandler(s.router),
	)
}

func (s *Service) mountTechnicalDebug() {
	s.router.Handle("/node", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.nodeGetHandler),
	})

	s.router.Handle("/addresses", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.addressesHandler),
	})

	s.router.Handle("/chainstate", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.chainStateHandler),
	})

	s.router.Path("/metrics").Handler(web.ChainHandlers(
		httpaccess.NewHTTPAccessSuppressLogHandler(),
		web.FinalHandler(promhttp.InstrumentMetricHandler(
			s.metricsRegistry,
			promhttp.HandlerFor(s.metricsRegistry, promhttp.HandlerOpts{}),
		)),
	))

	s.router.Handle("/debug/pprof", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.URL
		u.Path += "/"
		http.Redirect(w, r, u.String(), http.StatusPermanentRedirect)
	}))
	s.router.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	s.router.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	s.router.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	s.router.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	s.router.PathPrefix("/debug/pprof/").Handler(http.HandlerFunc(pprof.Index))

	s.router.Handle("/debug/vars", expvar.Handler())

	s.router.Handle("/loggers", jsonhttp.MethodHandler{
		"GET": web.ChainHandlers(
			httpaccess.NewHTTPAccessSuppressLogHandler(),
			web.FinalHandlerFunc(s.loggerGetHandler),
		),
	})
	s.router.Handle("/loggers/{exp}", jsonhttp.MethodHandler{
		"GET": web.ChainHandlers(
			httpaccess.NewHTTPAccessSuppressLogHandler(),
			web.FinalHandlerFunc(s.loggerGetHandler),
		),
	})
	s.router.Handle("/loggers/{exp}/{verbosity}", jsonhttp.MethodHandler{
		"PUT": web.ChainHandlers(
			httpaccess.NewHTTPAccessSuppressLogHandler(),
			web.FinalHandlerFunc(s.loggerSetVerbosityHandler),
		),
	})

	s.router.Handle("/readiness", web.ChainHandlers(
		httpaccess.NewHTTPAccessSuppressLogHandler(),
		web.FinalHandlerFunc(s.readinessHandler),
	))

	s.router.Handle("/health", web.ChainHandlers(
		httpaccess.NewHTTPAccessSuppressLogHandler(),
		web.FinalHandlerFunc(s.healthHandler),
	))
}

func (s *Service) mountAPI() {
	subdomainRouter := s.router.Host("{subdomain:.*}.cluster.localhost").Subrouter()

	subdomainRouter.Handle("/{path:.*}", jsonhttp.MethodHandler{
		"GET": web.ChainHandlers(
			web.FinalHandlerFunc(s.subdomainHandler),
		),
	})

	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "BNB Smart Chain Cluster Mop")
	})

	s.router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	// handle is a helper closure which simplifies the router setup.
	handle := func(path string, handler http.Handler) {
		s.router.Handle(path, handler)
		s.router.Handle(rootPath+path, handler)
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
		"HEAD": web.ChainHandlers(
			s.newTracingHandler("bytes-head"),
			web.FinalHandlerFunc(s.bytesHeadHandler),
		),
	})

	handle("/chunks", jsonhttp.MethodHandler{
		"POST": web.ChainHandlers(
			jsonhttp.NewMaxBodyBytesHandler(cluster.ChunkWithSpanSize),
			web.FinalHandlerFunc(s.chunkUploadHandler),
		),
	})

	handle("/chunks/stream", web.ChainHandlers(
		s.newTracingHandler("chunks-stream-upload"),
		web.FinalHandlerFunc(s.chunkUploadStreamHandler),
	))

	handle("/chunks/{address}", jsonhttp.MethodHandler{
		"GET":    http.HandlerFunc(s.chunkGetHandler),
		"HEAD":   http.HandlerFunc(s.hasChunkHandler),
		"DELETE": http.HandlerFunc(s.removeChunk),
	})

	handle("/soc/{owner}/{id}", jsonhttp.MethodHandler{
		"POST": web.ChainHandlers(
			jsonhttp.NewMaxBodyBytesHandler(cluster.ChunkWithSpanSize),
			web.FinalHandlerFunc(s.socUploadHandler),
		),
	})

	handle("/feeds/{owner}/{topic}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.feedGetHandler),
		"POST": web.ChainHandlers(
			jsonhttp.NewMaxBodyBytesHandler(cluster.ChunkWithSpanSize),
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
	})

	handle("/psser/send/{topic}/{targets}", web.ChainHandlers(
		web.FinalHandler(jsonhttp.MethodHandler{
			"POST": web.ChainHandlers(
				jsonhttp.NewMaxBodyBytesHandler(cluster.ChunkSize),
				web.FinalHandlerFunc(s.pssPostHandler),
			),
		})),
	)

	handle("/psser/subscribe/{topic}", web.ChainHandlers(
		web.FinalHandlerFunc(s.pssWsHandler),
	))

	handle("/tags", web.ChainHandlers(
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.listTagsHandler),
			"POST": web.ChainHandlers(
				jsonhttp.NewMaxBodyBytesHandler(1024),
				web.FinalHandlerFunc(s.createTagHandler),
			),
		})),
	)

	handle("/tags/{id}", web.ChainHandlers(
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
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.listPinnedRootHashes),
		})),
	)

	handle("/pins/{reference}", web.ChainHandlers(
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET":    http.HandlerFunc(s.getPinnedRootHash),
			"POST":   http.HandlerFunc(s.pinRootHash),
			"DELETE": http.HandlerFunc(s.unpinRootHash),
		})),
	)

	handle("/wardenship/{address}", jsonhttp.MethodHandler{
		"GET": web.ChainHandlers(
			web.FinalHandlerFunc(s.wardenshipGetHandler),
		),
		"PUT": web.ChainHandlers(
			web.FinalHandlerFunc(s.wardenshipPutHandler),
		),
	})

	handle("/readiness", web.ChainHandlers(
		httpaccess.NewHTTPAccessSuppressLogHandler(),
		web.FinalHandlerFunc(s.readinessHandler),
	))

	handle("/health", web.ChainHandlers(
		httpaccess.NewHTTPAccessSuppressLogHandler(),
		web.FinalHandlerFunc(s.healthHandler),
	))

	if s.Restricted {
		handle("/auth", jsonhttp.MethodHandler{
			"POST": web.ChainHandlers(
				s.newTracingHandler("auth"),
				jsonhttp.NewMaxBodyBytesHandler(512),
				web.FinalHandlerFunc(s.authHandler),
			),
		})
		handle("/refresh", jsonhttp.MethodHandler{
			"POST": web.ChainHandlers(
				s.newTracingHandler("auth"),
				jsonhttp.NewMaxBodyBytesHandler(512),
				web.FinalHandlerFunc(s.refreshHandler),
			),
		})
	}
}

func (s *Service) mountBusinessDebug(restricted bool) {
	handle := func(path string, handler http.Handler) {
		if restricted {
			handler = web.ChainHandlers(auth.PermissionCheckHandler(s.auth), web.FinalHandler(handler))
		}
		s.router.Handle(path, handler)
		s.router.Handle(rootPath+path, handler)
	}

	if s.transaction != nil {
		handle("/transactions", jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.transactionListHandler),
		})
		handle("/transactions/{hash}", jsonhttp.MethodHandler{
			"GET":    http.HandlerFunc(s.transactionDetailHandler),
			"POST":   http.HandlerFunc(s.transactionResendHandler),
			"DELETE": http.HandlerFunc(s.transactionCancelHandler),
		})
	}

	handle("/peers", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.peersHandler),
	})

	handle("/pingpong/{peer-id}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.pingpongHandler),
	})

	handle("/reservestate", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.reserveStateHandler),
	})

	handle("/connect/{multi-address:.+}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.peerConnectHandler),
	})

	handle("/blocklist", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.blocklistedPeersHandler),
	})

	handle("/peers/{address}", jsonhttp.MethodHandler{
		"DELETE": http.HandlerFunc(s.peerDisconnectHandler),
	})

	handle("/chunks/{address}", jsonhttp.MethodHandler{
		"GET":    http.HandlerFunc(s.hasChunkHandler),
		"DELETE": http.HandlerFunc(s.removeChunk),
	})

	handle("/topology", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.topologyHandler),
	})

	handle("/welcome-message", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.getWelcomeMessageHandler),
		"POST": web.ChainHandlers(
			jsonhttp.NewMaxBodyBytesHandler(welcomeMessageMaxRequestSize),
			web.FinalHandlerFunc(s.setWelcomeMessageHandler),
		),
	})

	handle("/balances", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.compensatedBalancesHandler),
	})

	handle("/balances/{peer}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.compensatedPeerBalanceHandler),
	})

	handle("/consumed", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.balancesHandler),
	})

	handle("/consumed/{peer}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.peerBalanceHandler),
	})

	handle("/timesettlements", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.settlementsHandlerPseudosettle),
	})

	handle("/node/balance", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.nodeBalanceHandler),
	})

	handle("/reward/balance", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.nodeRewardBalanceHandler),
	})

	handle("/reward/transactions", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.nodeRewardTransactionHandler),
	})

	handle("/reward/cash/{amount}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.nodeRewardCashHandler),
	})

	handle("/pledge/stake/{amount}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.nodePledgeStakeHandler),
	})

	handle("/pledge/unstake/{amount}", jsonhttp.MethodHandler{
		"POST": http.HandlerFunc(s.nodePledgeUnStakeHandler),
	})

	handle("/pledge/balance", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.nodePledgeHandler),
	})

	handle("/pledge/transactions", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.nodePledgeTransactionHandler),
	})

	if s.swapEnabled {
		handle("/settlements", jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.settlementsHandler),
		})

		handle("/settlements/{peer}", jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.peerSettlementsHandler),
		})

		handle("/chequebook/cheque/{peer}", jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.chequebookLastPeerHandler),
		})

		handle("/chequebook/cheque", jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.chequebookAllLastHandler),
		})

		handle("/chequebook/cashout/{peer}", jsonhttp.MethodHandler{
			"GET":  http.HandlerFunc(s.swapCashoutStatusHandler),
			"POST": http.HandlerFunc(s.swapCashoutHandler),
		})
	}

	if s.chequebookEnabled {
		handle("/chequebook/balance", jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.chequebookBalanceHandler),
		})

		handle("/chequebook/address", jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.chequebookAddressHandler),
		})

		handle("/chequebook/deposit", jsonhttp.MethodHandler{
			"POST": http.HandlerFunc(s.chequebookDepositHandler),
		})

		handle("/chequebook/withdraw", jsonhttp.MethodHandler{
			"POST": http.HandlerFunc(s.chequebookWithdrawHandler),
		})

		if s.swapEnabled {
			handle("/wallet", jsonhttp.MethodHandler{
				"GET": http.HandlerFunc(s.walletHandler),
			})
		}
	}

	handle("/stamps", web.ChainHandlers(
		s.voucherSyncStatusCheckHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.voucherGetStampsHandler),
		})),
	)

	handle("/stamps/{id}", web.ChainHandlers(
		s.voucherSyncStatusCheckHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.voucherGetStampHandler),
		})),
	)

	handle("/stamps/{id}/buckets", web.ChainHandlers(
		s.voucherSyncStatusCheckHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.voucherGetStampBucketsHandler),
		})),
	)

	handle("/stamps/{amount}/{depth}", web.ChainHandlers(
		s.voucherAccessHandler,
		s.voucherSyncStatusCheckHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"POST": http.HandlerFunc(s.voucherCreateHandler),
		})),
	)

	handle("/stamps/topup/{id}/{amount}", web.ChainHandlers(
		s.voucherAccessHandler,
		s.voucherSyncStatusCheckHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"PATCH": http.HandlerFunc(s.voucherTopUpHandler),
		})),
	)

	handle("/stamps/dilute/{id}/{depth}", web.ChainHandlers(
		s.voucherAccessHandler,
		s.voucherSyncStatusCheckHandler,
		web.FinalHandler(jsonhttp.MethodHandler{
			"PATCH": http.HandlerFunc(s.voucherDiluteHandler),
		})),
	)

	handle("/batches", web.ChainHandlers(
		web.FinalHandler(jsonhttp.MethodHandler{
			"GET": http.HandlerFunc(s.voucherGetAllStampsHandler),
		})),
	)

	handle("/tags/{id}", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.getDebugTagHandler),
	})

	handle("/bookkeeper", jsonhttp.MethodHandler{
		"GET": http.HandlerFunc(s.accountingInfoHandler),
	})

	handle("/readiness", web.ChainHandlers(
		httpaccess.NewHTTPAccessSuppressLogHandler(),
		web.FinalHandlerFunc(s.readinessHandler),
	))

	handle("/health", web.ChainHandlers(
		httpaccess.NewHTTPAccessSuppressLogHandler(),
		web.FinalHandlerFunc(s.healthHandler),
	))

}

type trafficObject struct {
	Timestamp  int64            `json:"timestamp"`
	Address    string           `json:"address"`
	Uploaded   map[string]int64 `json:"uploaded"`
	Downloaded map[string]int64 `json:"downloaded"`
	Signed     string           `json:"signed"`
}

func (s *Service) trafficHandler(t time.Time, key string, upload bool, size int) {
	duration := 10 * time.Minute
	if s.lru == nil {
		s.lru, _ = lru.NewWithEvict(1, func(key, value interface{}) {
			go func() {
				ticker := time.NewTicker(duration)
				for {
					<-ticker.C
					d := int64(duration / time.Second)
					timestamp := (time.Now().Unix() / d) * d
					if _, ok := s.lru.Get(timestamp); !ok {
						s.lru.Add(timestamp, &trafficObject{
							Timestamp:  timestamp,
							Uploaded:   make(map[string]int64),
							Downloaded: make(map[string]int64),
						})
					}
				}
			}()
			if len(s.Options.RemoteEndPoint) > 0 {
				traffic := value.(*trafficObject)
				if len(traffic.Downloaded) > 0 || len(traffic.Uploaded) > 0 {
					traffic.Address = s.bscAddress.String()
					bts, _ := json.Marshal(traffic)
					client := &http.Client{
						Transport: &http.Transport{
							TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
						},
					}
					resp, err := client.Post(s.Options.RemoteEndPoint+"/api/v1/traffic", "application/json", strings.NewReader(string(bts)))
					if err != nil {
						s.logger.Error(err, "traffic handler", "key", key, "val", string(bts))
					} else {
						s.logger.Debug("traffic handler", "key", key, "val", string(bts))
						resp.Body.Close()
					}
				}
			}
		})
	}

	d := int64(duration / time.Second)
	timestamp := (t.Unix() / d) * d

	traffic := &trafficObject{
		Timestamp:  timestamp,
		Uploaded:   make(map[string]int64),
		Downloaded: make(map[string]int64),
	}

	if val, ok := s.lru.Get(timestamp); ok {
		traffic = val.(*trafficObject)
	}
	if upload {
		traffic.Uploaded[key] += int64(size)
	} else {
		traffic.Downloaded[key] += int64(size)
	}
	s.lru.Add(timestamp, traffic)
}
