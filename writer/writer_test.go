package writer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/darkweak/go-esi/writer"
)

func TestWriter(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)

		_, _ = w.Write([]byte("included " + r.URL.Path))
	}))
	defer upstream.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`<p><esi:vars>$(HTTP_HOST)</esi:vars></p><esi:incl`))
		_ = http.NewResponseController(w).Flush()
		_, _ = w.Write([]byte(`ude src="/a"/>|<esi:include src="/b"/><esi:remove>gone</esi:remove>`))
	})

	esiHandler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cw := writer.NewWriter(rw, r)
		handler.ServeHTTP(cw, r)

		if err := cw.Close(); err != nil {
			t.Error(err)
		}
	})

	// Point the relative includes to the upstream server.
	rq := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	rec := httptest.NewRecorder()
	esiHandler.ServeHTTP(rec, rq)

	if rec.Code != http.StatusCreated {
		t.Errorf("want status %d, got %d", http.StatusCreated, rec.Code)
	}

	if cl := rec.Header().Get("Content-Length"); cl != "" {
		t.Errorf("the Content-Length header must be removed, got %q", cl)
	}

	want := "<p>" + rq.Host + "</p>included /a|included /b"
	if got := rec.Body.String(); got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}
