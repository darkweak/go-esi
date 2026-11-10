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

func TestWriterWithBaseURL(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("included " + r.URL.Path + " cookie=" + r.Header.Get("Cookie")))
	}))
	defer upstream.Close()

	base, err := writer.ParseBaseURL(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}

	esiHandler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cw := writer.NewWriter(rw, r, writer.WithBaseURL(base))
		_, _ = cw.Write([]byte(`<esi:include src="/a"/>|<esi:include src="b"/>|<esi:vars>$(HTTP_HOST)</esi:vars>`))

		if err := cw.Close(); err != nil {
			t.Error(err)
		}

		if r.URL.Host != "public.example.com" {
			t.Errorf("the incoming request must not be modified, got host %q", r.URL.Host)
		}
	})

	// The incoming host does not serve anything: the includes must reach the
	// base URL.
	rq := httptest.NewRequest(http.MethodGet, "http://public.example.com/dir/page", nil)
	rq.Header.Set("Cookie", "session=1")

	rec := httptest.NewRecorder()
	esiHandler.ServeHTTP(rec, rq)

	want := "included /a cookie=session=1|included /dir/b cookie=session=1|public.example.com"
	if got := rec.Body.String(); got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestParseBaseURL(t *testing.T) {
	t.Parallel()

	valid := map[string]string{
		"": "",
		"static-webserver.my-app.svc.cluster.local": "http://static-webserver.my-app.svc.cluster.local",
		"svc:8080":                         "http://svc:8080",
		"https://static.example.com:8443/": "https://static.example.com:8443",
	}

	for in, want := range valid {
		u, err := writer.ParseBaseURL(in)
		if err != nil {
			t.Errorf("ParseBaseURL(%q): unexpected error %v", in, err)

			continue
		}

		got := ""
		if u != nil {
			got = u.String()
		}

		if got != want {
			t.Errorf("ParseBaseURL(%q): want %q, got %q", in, want, got)
		}
	}

	invalid := []string{"ftp://svc", "http://", "http://svc/prefix", "http://svc?a=b", "http://user@svc", "http://[::1"}

	for _, in := range invalid {
		if _, err := writer.ParseBaseURL(in); err == nil {
			t.Errorf("ParseBaseURL(%q): want an error", in)
		}
	}
}
