package httpx

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

func WithTelemetry(h http.Handler) http.Handler {
	return otelhttp.NewHandler(h, "medication", otelhttp.WithTracerProvider(otel.GetTracerProvider()))
}
