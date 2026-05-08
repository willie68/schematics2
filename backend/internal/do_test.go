package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/config"
)

func TestNewRouterServesEmbeddedClient(t *testing.T) {
	cfg := config.Config{
		JWTSecret: "test-secret",
		AdminUser: "admin",
		AdminPass: "admin",
	}
	inj := do.New()
	if err := InitServices(inj, cfg); err != nil {
		t.Skipf("skip test because service init failed: %v", err)
	}

	router, err := NewRouter(inj)
	if err != nil {
		t.Skipf("skip test because router requires mongodb: %v", err)
	}

	t.Run("root redirects to client", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTemporaryRedirect {
			t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rr.Code)
		}
		if got := rr.Header().Get("Location"); got != "/client" {
			t.Fatalf("expected Location /client, got %q", got)
		}
	})

	t.Run("client route serves index", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/client", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); ct == "" {
			t.Fatal("expected Content-Type header")
		}
	})
}
