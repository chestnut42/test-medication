package metrics

import (
	"log"
	"net/http"

	clientprom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

var _registerer = clientprom.DefaultRegisterer
var _gatherer = clientprom.DefaultGatherer

func init() {
	exporter, err := prometheus.New(prometheus.WithRegisterer(_registerer))
	if err != nil {
		log.Fatalf("failed to initialize prometheus exporter: %v", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)
}

func NewHandler() http.Handler {
	return promhttp.HandlerFor(_gatherer, promhttp.HandlerOpts{})
}
