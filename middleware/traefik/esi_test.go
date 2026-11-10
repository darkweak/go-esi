package traefik_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darkweak/go-esi/middleware/traefik"
)

func TestDefaultHost(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fragment " + r.URL.Path))
	}))
	defer upstream.Close()

	cfg := traefik.CreateConfig()
	cfg.DefaultHost = upstream.Listener.Addr().String()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<esi:include src="/my-esi-content/whatever"/>`))
	})

	h, err := traefik.New(context.Background(), next, cfg, "esi")
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://domain.com/", nil))

	if got, want := rec.Body.String(), "fragment /my-esi-content/whatever"; got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestInvalidDefaultHost(t *testing.T) {
	t.Parallel()

	cfg := traefik.CreateConfig()
	cfg.DefaultHost = "ftp://svc"

	if _, err := traefik.New(context.Background(), http.NotFoundHandler(), cfg, "esi"); err == nil {
		t.Error("want an error for an invalid defaultHost")
	}
}
