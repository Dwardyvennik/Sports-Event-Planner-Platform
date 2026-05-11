package tests

import (
	"testing"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/config"
)

func TestConfigLoads(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.App.Name != "notification-service" {
		t.Fatalf("unexpected service name: %s", cfg.App.Name)
	}
}
