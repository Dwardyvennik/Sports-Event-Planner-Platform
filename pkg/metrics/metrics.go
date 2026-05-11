package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(service string) http.Handler {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	serviceUp := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "sports_planner",
		Name:        "service_up",
		Help:        "Static service availability marker exposed by each process.",
		ConstLabels: prometheus.Labels{"service": service},
	})
	serviceUp.Set(1)
	registry.MustRegister(serviceUp)

	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
