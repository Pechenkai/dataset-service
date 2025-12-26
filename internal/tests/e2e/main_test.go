//go:build e2e

package e2e

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"ppo/internal/di"
	"ppo/internal/integrations/catfacts"
	"ppo/internal/integrations/openai"
)

var (
	e2eServer    *httptest.Server
	e2eApp       *di.App
	catfactsSrv  *httptest.Server
	catfactsProc *exec.Cmd
	openaiSrv    *httptest.Server
	openaiProc   *exec.Cmd
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

	mode := strings.ToLower(os.Getenv("EXTERNAL_CATFACTS_MODE"))
	if mode == "" {
		mode = "mock"
		_ = os.Setenv("EXTERNAL_CATFACTS_MODE", mode)
	}
	if mode == "mock" && os.Getenv("EXTERNAL_CATFACTS_MOCK_BASE_URL") == "" {
		startCatfactsMock()
	}

	openaiMode := strings.ToLower(os.Getenv("EXTERNAL_OPENAI_MODE"))
	if openaiMode == "" {
		openaiMode = "mock"
		_ = os.Setenv("EXTERNAL_OPENAI_MODE", openaiMode)
	}
	if openaiMode == "mock" && os.Getenv("EXTERNAL_OPENAI_BASE_URL") == "" {
		startOpenAIMock()
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

	if catfactsSrv != nil {
		catfactsSrv.Close()
	}
	if catfactsProc != nil && catfactsProc.Process != nil {
		_ = catfactsProc.Process.Kill()
	}
	if openaiSrv != nil {
		openaiSrv.Close()
	}
	if openaiProc != nil && openaiProc.Process != nil {
		_ = openaiProc.Process.Kill()
	}

	os.Exit(code)
}

func startCatfactsMock() {
	if bin := strings.TrimSpace(os.Getenv("EXTERNAL_CATFACTS_BIN")); bin != "" {
		addr := strings.TrimSpace(os.Getenv("EXTERNAL_CATFACTS_ADDR"))
		if addr == "" {
			addr = ":9099"
		}
		baseURL := strings.TrimSpace(os.Getenv("EXTERNAL_CATFACTS_MOCK_BASE_URL"))
		if baseURL == "" {
			baseURL = normalizeMockBaseURL(addr)
		}
		catfactsProc = exec.Command(bin, "-addr", addr) // #nosec G204 test helper
		_ = catfactsProc.Start()
		time.Sleep(200 * time.Millisecond)
		_ = os.Setenv("EXTERNAL_CATFACTS_MOCK_BASE_URL", baseURL)
		return
	}

	catfactsSrv = httptest.NewServer(catfacts.DefaultMockHandler())
	_ = os.Setenv("EXTERNAL_CATFACTS_MOCK_BASE_URL", catfactsSrv.URL)
}

func normalizeMockBaseURL(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return strings.TrimRight(addr, "/")
	}
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr
	}
	return "http://" + addr
}

func startOpenAIMock() {
	if bin := strings.TrimSpace(os.Getenv("EXTERNAL_OPENAI_BIN")); bin != "" {
		addr := strings.TrimSpace(os.Getenv("EXTERNAL_OPENAI_ADDR"))
		if addr == "" {
			addr = ":8088"
		}
		baseURL := strings.TrimSpace(os.Getenv("EXTERNAL_OPENAI_BASE_URL"))
		if baseURL == "" {
			baseURL = normalizeMockBaseURL(addr)
		}
		openaiProc = exec.Command(bin, "-addr", addr) // #nosec G204 test helper
		_ = openaiProc.Start()
		time.Sleep(200 * time.Millisecond)
		_ = os.Setenv("EXTERNAL_OPENAI_BASE_URL", baseURL)
		return
	}
	openaiSrv = httptest.NewServer(openai.DefaultMockHandler())
	_ = os.Setenv("EXTERNAL_OPENAI_BASE_URL", openaiSrv.URL)
}
