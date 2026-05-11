package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Checker func(context.Context) error

type Response struct {
	Service string            `json:"service"`
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks,omitempty"`
	Time    string            `json:"time"`
}

func Liveness(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, Response{
			Service: service,
			Status:  "ok",
			Time:    time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func Readiness(service string, checks map[string]Checker, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "ok"
		code := http.StatusOK
		results := make(map[string]string, len(checks))

		for name, check := range checks {
			if err := check(ctx); err != nil {
				status = "unavailable"
				code = http.StatusServiceUnavailable
				results[name] = err.Error()
				log.WarnContext(ctx, "readiness check failed", "check", name, "error", err)
				continue
			}
			results[name] = "ok"
		}

		writeJSON(w, code, Response{
			Service: service,
			Status:  status,
			Checks:  results,
			Time:    time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func writeJSON(w http.ResponseWriter, code int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
