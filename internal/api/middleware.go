package api

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// responseWriterInterceptor captures the HTTP status code for logging.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// MiddlewareStack chains multiple middleware functions around an http.Handler.
type Middleware func(http.Handler) http.Handler

// CreateMiddlewareStack wraps a handler with a sequence of middleware.
func CreateMiddlewareStack(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// CORSMiddleware enables cross-origin resource sharing for the frontend SPA.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware catches any panics and returns an RFC 7807 500 error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC RECOVERED] %v\nStack: %s", rec, debug.Stack())
				WriteError(
					w, r, http.StatusInternalServerError,
					"ERR_INTERNAL_SERVER_ERROR", "Internal Server Error",
					"An unexpected internal error occurred while processing the request.",
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs incoming requests with status code and elapsed time.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		interceptor := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(interceptor, r)

		elapsed := time.Since(start)
		log.Printf("[HTTP] %s %s -> %d (%s)", r.Method, r.URL.Path, interceptor.statusCode, elapsed)
	})
}
