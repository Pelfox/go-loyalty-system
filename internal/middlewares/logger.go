package middlewares

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// LoggerMiddleware реализует middleware для логгирования всех запросов.
func LoggerMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	logger = logger.With().Str("middleware", "logger").Logger()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			next.ServeHTTP(writer, r)
			duration := time.Since(start)

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", writer.Status()).
				Int("size", writer.BytesWritten()).
				Dur("duration", duration).
				Msg("serving HTTP request")
		})
	}
}
