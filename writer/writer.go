package writer

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/darkweak/go-esi/esi"
)

// Writer is an http.ResponseWriter processing the ESI tags of the response
// body while it is streamed to the client. Close must be called once the
// wrapped handler returns.
type Writer struct {
	rw          http.ResponseWriter
	Rq          *http.Request
	stream      *esi.Stream
	wroteHeader bool
}

// ErrInvalidBaseURL is returned by ParseBaseURL when the value is not a valid
// origin.
var ErrInvalidBaseURL = errors.New("esi: the base URL must be an origin like host[:port] or scheme://host[:port]")

// ParseBaseURL parses an origin to use with WithBaseURL. The scheme is
// optional and defaults to http, e.g. "my-service.my-namespace.svc.cluster.local"
// or "https://static.example.com:8443". An empty string returns a nil URL.
func ParseBaseURL(s string) (*url.URL, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil //nolint:nilnil // no base URL configured.
	}

	if !strings.Contains(s, "://") {
		s = "http://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, errors.Join(ErrInvalidBaseURL, err)
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil ||
		(u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return nil, ErrInvalidBaseURL
	}

	return &url.URL{Scheme: u.Scheme, Host: u.Host}, nil
}

// Option configures a Writer.
type Option func(*options)

type options struct {
	baseURL *url.URL
}

// WithBaseURL resolves the relative include URLs (e.g. src="/fragment")
// against the scheme and host of base instead of the ones of the incoming
// request. The path of the incoming request is kept, so relative paths
// (e.g. src="fragment") still resolve from it. The Cookie and Authorization
// headers are forwarded to base as it becomes the same-origin target. A nil
// base is ignored.
func WithBaseURL(base *url.URL) Option {
	return func(o *options) {
		o.baseURL = base
	}
}

// NewWriter returns a Writer streaming the processed response to rw.
func NewWriter(rw http.ResponseWriter, rq *http.Request, opts ...Option) *Writer {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	if rq.URL.Scheme == "" {
		if rq.TLS != nil {
			rq.URL.Scheme = "https"
		} else {
			rq.URL.Scheme = "http"
		}
	}

	if rq.URL.Host == "" {
		rq.URL.Host = rq.Host
	}

	return &Writer{
		rw:     rw,
		Rq:     rq,
		stream: esi.NewStream(rw, includeRequest(rq, o.baseURL)),
	}
}

// includeRequest returns the request the include URLs are resolved against.
// rq is left untouched since the wrapped handler still serves it.
func includeRequest(rq *http.Request, base *url.URL) *http.Request {
	if base == nil {
		return rq
	}

	u := *rq.URL
	u.Scheme = base.Scheme
	u.Host = base.Host

	r := rq.WithContext(rq.Context())
	r.URL = &u

	return r
}

// Header implements http.ResponseWriter.
func (w *Writer) Header() http.Header {
	return w.rw.Header()
}

// WriteHeader implements http.ResponseWriter. The Content-Length header is
// dropped because processing the ESI tags changes the body length.
func (w *Writer) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	// Informational responses are forwarded, the final one is still to come.
	if statusCode >= http.StatusContinue && statusCode < http.StatusOK && statusCode != http.StatusSwitchingProtocols {
		w.rw.WriteHeader(statusCode)

		return
	}

	w.wroteHeader = true
	w.rw.Header().Del("Content-Length")
	w.rw.WriteHeader(statusCode)
}

// Write implements http.ResponseWriter.
func (w *Writer) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.stream.Write(b)
}

// Flush implements http.Flusher.
func (w *Writer) Flush() {
	w.stream.Flush()
}

// Close processes the remaining buffered content and waits for the pending
// ESI tags to be written.
func (w *Writer) Close() error {
	return w.stream.Close()
}

// Unwrap returns the underlying http.ResponseWriter (used by http.ResponseController).
func (w *Writer) Unwrap() http.ResponseWriter {
	return w.rw
}

var (
	_ http.ResponseWriter = (*Writer)(nil)
	_ http.Flusher        = (*Writer)(nil)
)
