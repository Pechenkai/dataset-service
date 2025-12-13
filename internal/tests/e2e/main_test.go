//go:build e2e

package e2e

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ppo/internal/di"
)

var (
	e2eServer *httptest.Server
	e2eApp    *di.App
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	if os.Getenv("AUTH_TWOFA_BLOCK_DURATION") == "" {
		_ = os.Setenv("AUTH_TWOFA_BLOCK_DURATION", "3s")
	}
	if os.Getenv("AUTH_TWOFA_CODE_TTL") == "" {
		_ = os.Setenv("AUTH_TWOFA_CODE_TTL", "5m")
	}
	if os.Getenv("AUTH_TWOFA_DEBUG_SECRET") == "" {
		_ = os.Setenv("AUTH_TWOFA_DEBUG_SECRET", "e2e-debug-secret")
	}
	_ = os.MkdirAll("logs", 0o755)

	app, err := di.Build(ctx)
	if err != nil {
		if os.Getenv("E2E_REQUIRE_INFRA") == "1" {
			log.Fatalf("failed to bootstrap app for e2e (E2E_REQUIRE_INFRA=1): %v", err)
		}
		log.Printf("skipping e2e suite: infrastructure unavailable: %v", err)
		os.Exit(0)
	}
	e2eApp = app

	srv := httptest.NewServer(app.HTTPHandler)
	e2eServer = srv
	base := srv.URL + "/api/v2"
	_ = os.Setenv("E2E_BASE_URL", base)

	code := m.Run()

	srv.Close()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.Shutdown(shutdownCtx)

	os.Exit(code)
}
