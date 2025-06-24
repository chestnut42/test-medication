package httpx

import (
	"log/slog"
	"net/http"

	"github.com/felixge/httpsnoop"

	"github.com/chestnut42/test-medication/internal/utils/logx"
)

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logx.Logger(r.Context()).With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path))

		ctx := r.Context()
		ctx = logx.WithLogger(ctx, logger)
		r = r.WithContext(ctx)

		m := httpsnoop.CaptureMetrics(h, w, r)
		logx.Logger(r.Context()).Info("request served",
			slog.Int("code", m.Code),
			slog.Duration("dt", m.Duration),
			slog.Int64("written", m.Written))
	})
}
