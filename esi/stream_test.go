package esi_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/darkweak/go-esi/esi"
)

// offline exercises every tag that does not need the network, including
// unsupported tags and constructs straddling chunk boundaries.
const offline = `<html><head><title><esi:vars>Hello from $(HTTP_HOST)</esi:vars></title></head>
<body>
<!--esi
  <p><esi:vars>Hello, $(HTTP_COOKIE{name})!</esi:vars></p>
-->
<esi:comment text="a comment" />
<p>before</p><esi:remove><a href="http://www.example.com">www.example.com</a></esi:remove><p>after</p>
<esi:try><esi:attempt>unsupported</esi:attempt></esi:try>
<esi:vars>$(HTTP_COOKIE{missing}|'fallback') $(QUERY_STRING{q})</esi:vars>
< esi not a tag <!-- regular comment --> <es
</body></html>`

func streamInChunks(t *testing.T, input []byte, req *http.Request, chunks ...int) string {
	t.Helper()

	var out bytes.Buffer

	s := esi.NewStream(&out, req)

	for _, n := range chunks {
		n = min(n, len(input))
		if _, err := s.Write(input[:n]); err != nil {
			t.Fatal(err)
		}

		input = input[n:]
	}

	if _, err := s.Write(input); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	return out.String()
}

func Test_Stream_offline_everySplit(t *testing.T) {
	t.Parallel()

	input := []byte(offline)
	req := getRequest()
	want := string(esi.Parse(bytes.Clone(input), req))

	for i := range len(input) + 1 {
		if got := streamInChunks(t, input, req, i); got != want {
			t.Fatalf("split at %d:\nwant:\n%s\ngot:\n%s", i, want, got)
		}
	}

	ones := make([]int, len(input))
	for i := range ones {
		ones[i] = 1
	}

	if got := streamInChunks(t, input, req, ones...); got != want {
		t.Fatalf("byte by byte:\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func Test_Stream_fixtures(t *testing.T) {
	t.Parallel()

	for fixture, want := range expected {
		if fixture == "vars" {
			continue // needs a dedicated request, covered by the offline test.
		}

		input := loadFromFixtures(fixture)

		ones := make([]int, len(input))
		for i := range ones {
			ones[i] = 1
		}

		for name, chunks := range map[string][]int{"whole": nil, "half": {len(input) / 2}, "byte by byte": ones} {
			if got := streamInChunks(t, input, getRequest(), chunks...); got != want {
				t.Errorf("%s (%s):\nwant:\n%s\ngot:\n%s", fixture, name, want, got)
			}
		}
	}
}

// slowServer answers /<name>?d=<duration> with <name> after the delay.
func slowServer(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, _ := time.ParseDuration(r.URL.Query().Get("d"))
		time.Sleep(d)

		_, _ = w.Write([]byte(strings.TrimPrefix(r.URL.Path, "/")))
	}))
	t.Cleanup(srv.Close)

	return srv
}

func includes(srv *httptest.Server, n int, delay time.Duration) (doc, want string) {
	var d, w strings.Builder
	for i := range n {
		fmt.Fprintf(&d, `[%d]<esi:include src="%s/inc%d?d=%s"/>`, i, srv.URL, i, delay)
		fmt.Fprintf(&w, `[%d]inc%d`, i, i)
	}

	return d.String(), w.String()
}

func Test_Parse_parallel(t *testing.T) {
	t.Parallel()

	const (
		n     = 20
		delay = 200 * time.Millisecond
	)

	srv := slowServer(t)
	doc, want := includes(srv, n, delay)

	start := time.Now()
	got := string(esi.Parse([]byte(doc), httptest.NewRequest(http.MethodGet, srv.URL, nil)))
	elapsed := time.Since(start)

	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}

	if elapsed > 3*delay {
		t.Fatalf("%d includes of %s took %s, they are not processed in parallel", n, delay, elapsed)
	}
}

func Test_Stream_parallel(t *testing.T) {
	t.Parallel()

	const (
		n     = 20
		delay = 200 * time.Millisecond
	)

	srv := slowServer(t)
	doc, want := includes(srv, n, delay)

	start := time.Now()
	got := streamInChunks(t, []byte(doc), httptest.NewRequest(http.MethodGet, srv.URL, nil), 7, 13, 50)
	elapsed := time.Since(start)

	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}

	if elapsed > 3*delay {
		t.Fatalf("%d includes of %s took %s, they are not processed in parallel", n, delay, elapsed)
	}
}

// recorder records the writes and flushes with their timestamp.
type recorder struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	flushed []string
	first   time.Time
}

func (r *recorder) Write(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.first.IsZero() {
		r.first = time.Now()
	}

	return r.buf.Write(b)
}

func (r *recorder) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.flushed = append(r.flushed, r.buf.String())
}

func Test_Stream_flushesBeforeSlowInclude(t *testing.T) {
	t.Parallel()

	const delay = 300 * time.Millisecond

	srv := slowServer(t)
	rec := &recorder{}
	s := esi.NewStream(rec, httptest.NewRequest(http.MethodGet, srv.URL, nil))
	start := time.Now()

	_, _ = s.Write([]byte(`<header/><esi:include src="/fast"/>`))
	_, _ = s.Write([]byte(`<esi:include src="/slow?d=` + delay.String() + `"/><footer/>`))

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	if got, want := rec.buf.String(), "<header/>fastslow<footer/>"; got != want {
		t.Fatalf("want %q, got %q", want, got)
	}

	if rec.first.Sub(start) >= delay {
		t.Fatal("the content before the slow include must be written without waiting for it")
	}

	if len(rec.flushed) == 0 || !strings.HasPrefix(rec.flushed[0], "<header/>") ||
		strings.Contains(rec.flushed[0], "slow") {
		t.Fatalf("the content preceding the slow include must be flushed while waiting, flushes: %q", rec.flushed)
	}
}

type failingWriter struct{}

var errBroken = errors.New("broken pipe")

func (failingWriter) Write([]byte) (int, error) { return 0, errBroken }

func Test_Stream_writeError(t *testing.T) {
	t.Parallel()

	s := esi.NewStream(failingWriter{}, getRequest())
	if _, err := s.Write([]byte("hello")); !errors.Is(err, errBroken) {
		t.Fatalf("want %v, got %v", errBroken, err)
	}

	if err := s.Close(); !errors.Is(err, errBroken) {
		t.Fatalf("want %v, got %v", errBroken, err)
	}

	if _, err := s.Write([]byte("x")); !errors.Is(err, esi.ErrClosed) {
		t.Fatalf("want %v, got %v", esi.ErrClosed, err)
	}
}

func BenchmarkStreamFull(b *testing.B) {
	input := []byte(full)

	for b.Loop() {
		s := esi.NewStream(&bytes.Buffer{}, httptest.NewRequest(http.MethodGet, "http://domain.com:9080", nil))
		_, _ = s.Write(input[:len(input)/2])
		_, _ = s.Write(input[len(input)/2:])
		_ = s.Close()
	}
}

var offlineBig = bytes.Repeat([]byte(offline), 200)

func BenchmarkParseOffline(b *testing.B) {
	req := getRequest()

	b.SetBytes(int64(len(offlineBig)))

	for b.Loop() {
		esi.Parse(offlineBig, req)
	}
}

func BenchmarkStreamOffline(b *testing.B) {
	req := getRequest()

	b.SetBytes(int64(len(offlineBig)))

	for b.Loop() {
		s := esi.NewStream(&bytes.Buffer{}, req)
		for chunk := range slices.Chunk(offlineBig, 4096) {
			_, _ = s.Write(chunk)
		}

		_ = s.Close()
	}
}
