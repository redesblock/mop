package httpaccess

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/tracer"
)

// NewHTTPAccessSuppressLogHandler creates a
// handler that will suppress access log messages.
func NewHTTPAccessSuppressLogHandler() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rr, ok := w.(*responseRecorder); ok {
				w = rr.ResponseWriter
			}
			h.ServeHTTP(w, r)
		})
	}
}

// NewHTTPAccessLogHandler creates a handler that
// will log a message after a request has been served.
func NewHTTPAccessLogHandler(logger log.Logger, t *tracer.Tracer, message string, trafficHandler func(time time.Time, key string, upload bool, size int64)) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rr, ok := w.(*responseRecorder)
			if !ok { // No need to layer on another responseRecorder.
				rr = &responseRecorder{ResponseWriter: w}
			}

			now := time.Now()
			h.ServeHTTP(rr, r)
			if logger.Verbosity() < log.VerbosityInfo {
				return
			}
			duration := time.Since(now)

			ctx, _ := t.WithContextFromHTTPHeaders(r.Context(), r.Header)

			logger := tracer.NewLoggerWithTraceID(ctx, logger)

			status := rr.status
			if status == 0 {
				status = http.StatusOK
			}

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			fields := []interface{}{
				"ip", ip,
				"method", r.Method,
				"host", r.Host,
				"uri", r.RequestURI,
				"proto", r.Proto,
				"status", status,
				"size", rr.size,
				"duration", duration,
			}
			if v := r.Referer(); v != "" {
				fields = append(fields, "referrer", v)
			}
			if v := r.UserAgent(); v != "" {
				fields = append(fields, "user-agent", v)
			}
			if v := r.Header.Get("X-Forwarded-For"); v != "" {
				fields = append(fields, "x-forwarded-for", v)
			}
			if v := r.Header.Get("X-Real-Ip"); v != "" {
				fields = append(fields, "x-real-ip", v)
			}
			key := "Unknown"
			if reqToken := r.Header.Get("Authorization"); strings.HasPrefix(reqToken, "Bearer ") {
				keys := strings.Split(reqToken, "Bearer ")
				fields = append(fields, "key", keys[1])
				key = keys[1]
			}
			if trafficHandler != nil && r.Method != http.MethodOptions {
				//if strings.Contains(r.RequestURI, "/bytes") || strings.Contains(r.RequestURI, "/chunks") || strings.Contains(r.RequestURI, "/mop") {
				if strings.Contains(r.RequestURI, "/mop") {
					size := int64(rr.size)
					if r.Method != http.MethodGet {
						size = r.ContentLength
					}
					fields = append(fields, "traffic", size)
					trafficHandler(now, key, r.Method != http.MethodGet, size)
				}
			}

			logger.WithValues(fields...).Build().Info(message)
		})
	}
}

// responseRecorder is an implementation of
// http.ResponseWriter that records various metrics.
type responseRecorder struct {
	http.ResponseWriter

	// Metrics.
	status int
	size   int
}

// Write implements http.ResponseWriter.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	size, err := rr.ResponseWriter.Write(b)
	rr.size += size
	return size, err
}

// WriteHeader implements http.ResponseWriter.
func (rr *responseRecorder) WriteHeader(s int) {
	rr.ResponseWriter.WriteHeader(s)
	if rr.status == 0 {
		rr.status = s
	}
}

// CloseNotify implements http.CloseNotifier.
func (rr *responseRecorder) CloseNotify() <-chan bool {
	// staticcheck SA1019 CloseNotifier interface is required by gorilla compress handler.
	// nolint:staticcheck
	return rr.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Hijack implements http.Hijacker.
func (rr *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rr.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush implements http.Flusher.
func (rr *responseRecorder) Flush() {
	rr.ResponseWriter.(http.Flusher).Flush()
}

// Push implements http.Pusher.
func (rr *responseRecorder) Push(target string, opts *http.PushOptions) error {
	return rr.ResponseWriter.(http.Pusher).Push(target, opts)
}
