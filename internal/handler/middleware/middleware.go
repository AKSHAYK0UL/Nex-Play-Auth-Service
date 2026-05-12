package middleware

import (
	"log/slog"
	"net/http"
	"nex_play_auth/github.com/pkg/response"
	"runtime/debug"
	"time"
)

type Middleware func(http.Handler) http.Handler

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {

	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

// Chain applies middlewares in order
func chain(h http.Handler, middlerwares ...Middleware) http.Handler {

	for i := len(middlerwares) - 1; i >= 0; i-- {

		h = middlerwares[i](h)
	}

	return h
}

//built in middlewares

// Logger logs method, path, status code, latency, and client IP for every request.
func Logger(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"latency", time.Since(start).String(),
			"ip", r.RemoteAddr,
		)
	})
}

// RecoverPanic logs the stack trace, and returns a 500 instead of crashing the server
func RecoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {

			if err := recover(); err != nil {

				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
				)

				response.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS
func CORS(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {

			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders
func SecurityHeaders(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")

		next.ServeHTTP(w, r)
	})
}
