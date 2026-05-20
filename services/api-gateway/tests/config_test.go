package tests

import (
	"testing"

	"github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/config"
)

func TestConfigLoads(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.App.Name != "api-gateway" {
		t.Fatalf("unexpected service name: %s", cfg.App.Name)
	}
	if len(cfg.Endpoints) != 3 {
		t.Fatalf("expected 3 upstream endpoints, got %d", len(cfg.Endpoints))
	}
}
