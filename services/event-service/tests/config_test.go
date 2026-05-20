package tests

import (
	"testing"

	"github.com/dwardyvennik/sports-event-planner-platform/services/event-service/internal/config"
)

func TestConfigLoads(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.App.Name != "event-service" {
		t.Fatalf("unexpected service name: %s", cfg.App.Name)
	}
}
