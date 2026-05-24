package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const RequestIDKey = "request_id"

func RequestLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = "unknown"
			}

			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r.WithContext(ctx))

			duration := time.Since(startedAt)
			statusCode := ww.Status()

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", statusCode),
				zap.Duration("duration", duration),
				zap.String("request_id", requestID),
			}

			if statusCode >= 500 {
				logger.Error("http request completed", fields...)
			} else if statusCode >= 400 {
				logger.Warn("http request completed", fields...)
			} else {
				logger.Info("http request completed", fields...)
			}
		})
	}
}
