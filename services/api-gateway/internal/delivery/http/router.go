package http

import (
	"encoding/json"
	"net/http"

	"github.com/university/sports-event-planner-platform/services/api-gateway/internal/usecase"
)

func RegisterRoutes(mux *http.ServeMux, gateway *usecase.GatewayUseCase) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"service": "api-gateway",
			"status":  "ok",
		})
	})

	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, gateway.Status())
	})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
