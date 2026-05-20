package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serviceUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sports_planner",
		Name:      "service_up",
		Help:      "Static service availability marker exposed by each process.",
	}, []string{"service"})

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests.",
	}, []string{"service", "method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "sports_planner",
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "method", "path"})

	httpErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Name:      "http_errors_total",
		Help:      "Total number of HTTP requests completed with 4xx or 5xx statuses.",
	}, []string{"service", "method", "path", "status"})

	EventsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Name:      "events_created_total",
		Help:      "Total number of events created successfully.",
	})

	EventJoinTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Name:      "event_join_total",
		Help:      "Total number of successful event joins.",
	})

	ActiveUsersTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Name:      "active_users_total",
		Help:      "Total number of successful auth sessions issued.",
	})

	NotificationsSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Name:      "notifications_sent_total",
		Help:      "Total number of notifications sent successfully.",
	})

	NotificationsFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Name:      "notifications_failed_total",
		Help:      "Total number of notification send failures.",
	})

	NATSPublishedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "nats",
		Name:      "published_total",
		Help:      "Total number of NATS messages published successfully.",
	}, []string{"subject"})

	NATSPublishFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "nats",
		Name:      "publish_failed_total",
		Help:      "Total number of failed NATS publish attempts.",
	}, []string{"subject"})

	NATSConsumedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "nats",
		Name:      "consumed_total",
		Help:      "Total number of NATS messages consumed successfully.",
	}, []string{"subject"})

	NATSConsumerRetryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "nats",
		Name:      "consumer_retry_total",
		Help:      "Total number of NATS consumer retry attempts.",
	}, []string{"subject"})

	NATSConsumerFailedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "nats",
		Name:      "consumer_failed_total",
		Help:      "Total number of NATS messages that failed after all retry attempts.",
	}, []string{"subject"})
)

func Handler(service string) http.Handler {
	serviceUp.WithLabelValues(service).Set(1)
	return promhttp.Handler()
}

func InstrumentHTTP(service string, next http.Handler) http.Handler {
	serviceUp.WithLabelValues(service).Set(1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		ObserveHTTP(service, r.Method, r.URL.Path, recorder.status, time.Since(started))
	})
}

func ObserveHTTP(service string, method string, path string, status int, duration time.Duration) {
	if path == "" {
		path = "unknown"
	}
	statusCode := strconv.Itoa(status)
	httpRequestsTotal.WithLabelValues(service, method, path, statusCode).Inc()
	httpRequestDuration.WithLabelValues(service, method, path).Observe(duration.Seconds())
	if status >= 400 {
		httpErrorsTotal.WithLabelValues(service, method, path, statusCode).Inc()
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
