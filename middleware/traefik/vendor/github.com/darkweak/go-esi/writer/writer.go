package writer

import (
	"net/http"

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

// NewWriter returns a Writer streaming the processed response to rw.
func NewWriter(rw http.ResponseWriter, rq *http.Request) *Writer {
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
		stream: esi.NewStream(rw, rq),
	}
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
