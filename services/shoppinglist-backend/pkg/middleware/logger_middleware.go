package middleware

import (
	"ShoppingList-Backend/pkg/server"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Config struct {
	// Next defines a function to skip this middleware when returned true
	// Optional. Default: nil
	// Next func(c *fiber.Ctx) bool

	// TimeFormat https://pkg.go.dev/time#Time.Format
	//
	// Optional. Default: 2006-01-02 15:04:05
	TimeFormat          string
	RedactedQueryParams []string
}

// Minimal wrapper around http.ResponseWriter to capture http status code
type responseWriter struct {
	http.ResponseWriter
	flusher       http.Flusher
	hijacker      http.Hijacker
	closeNotifier http.CloseNotifier

	status      int
	wroteHeader bool
}

var ConfigDefault = Config{
	// Next:       nil,
	TimeFormat: "2006-01-02 15:04:05",
}

var (
	headerXRequestId = http.CanonicalHeaderKey("X-Request-ID")
)

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	hijacker, _ := w.(http.Hijacker)
	flusher, _ := w.(http.Flusher)
	closeNotifier, _ := w.(http.CloseNotifier)
	return &responseWriter{ResponseWriter: w, hijacker: hijacker, flusher: flusher, closeNotifier: closeNotifier}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// Hijack was implemented to support websockets
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.hijacker == nil {
		return nil, nil, errors.New("http.Hijacker not implemenetd by underlying http.ResponseWriter")
	}
	return rw.hijacker.Hijack()
}

func (rw *responseWriter) Flush() {
	if rw.flusher != nil {
		rw.flusher.Flush()
	}
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	if rw.closeNotifier != nil {
		return rw.closeNotifier.CloseNotify()
	}
	return nil
}

func configDefault(config ...Config) Config {
	if len(config) < 1 {
		return ConfigDefault
	}

	cfg := config[0]
	// if cfg.Next == nil {
	// 	cfg.Next = ConfigDefault.Next
	// }
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = ConfigDefault.TimeFormat
	}
	return cfg
}

func NewZapLogger(config ...Config) func(http.Handler) http.Handler {

	cfg := configDefault(config...)

	var (
	// once       sync.Once
	// errHandler fiber.ErrorHandler
	)

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get/Set request id
			requestId := ""
			requestIds, ok := r.Header[headerXRequestId]
			if !ok || len(requestIds) == 0 {
				requestId = uuid.New().String()
				// TODO: should logging middleware be responsible for creating request id?
				r.Header[headerXRequestId] = []string{requestId}
			} else {
				requestId = requestIds[0]
			}

			// responseWriter is wrapped, so status can be inspected after it has been sent
			wrapped := wrapResponseWriter(w)

			defer func() {
				// Panic recover
				rvr := recover()
				var errorMsg error = nil
				errorStack := ""

				if rvr != nil {
					err, ok := rvr.(error)
					if !ok {
						err = fmt.Errorf("%v", rvr)
					}
					errorMsg = err
					errorStack = string(debug.Stack())
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(server.HTTPError{
						Status: http.StatusInternalServerError,
						// TODO: Disable/enable this via config switch
						Error: fmt.Sprintf("Unhandled error: %v. Stacktrace: %v", errorMsg, errorStack),
					})
				}

				message := ""
				status := wrapped.status
				switch {
				case rvr != nil:
					message = "panic recover"
				case status >= 500:
					message = "server error"
				case status >= 400:
					message = "client error"
				case status >= 300:
					message = "redirect"
				case status >= 200:
					message = "success"
				case status >= 100:
					message = "informative"
				default:
					message = "unknown status"
				}

				logger := zap.L()
				defer logger.Sync()
				// TODO: make fields configurable

				url := r.URL
				query := url.Query()
				for _, param := range cfg.RedactedQueryParams {
					if query.Has(param) {
						query.Set(param, "<REDACTED>")
					}
				}
				url.RawQuery = query.Encode()

				logger.Info(message,
					zap.String("Timestamp", start.Format(cfg.TimeFormat)),
					zap.Int("Status", status),
					zap.Duration("Duration", time.Since(start)),
					zap.String("IP", r.RemoteAddr),
					zap.String("RequestID", requestId),
					zap.String("Method", r.Method),
					zap.String("Path", url.String()),
					zap.String("Stacktrace", errorStack),
					zap.NamedError("Error", errorMsg),
				)
			}()

			next.ServeHTTP(wrapped, r)
		}
		return http.HandlerFunc(fn)
	}

}
