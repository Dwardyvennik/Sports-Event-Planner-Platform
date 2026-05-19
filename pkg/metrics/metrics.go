package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var serviceUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "sports_planner",
	Name:      "service_up",
	Help:      "Static service availability marker exposed by each process.",
}, []string{"service"})

func Handler(service string) http.Handler {
	serviceUp.WithLabelValues(service).Set(1)
	return promhttp.Handler()
}
