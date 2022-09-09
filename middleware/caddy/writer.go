package caddy_esi

import (
	"bytes"
	"net/http"
)

type writer struct {
	buf    *bytes.Buffer
	rw     http.ResponseWriter
	status int
}

func newWriter(buf *bytes.Buffer, rw http.ResponseWriter) *writer {
	return &writer{
		buf: buf,
		rw:  rw,
	}
}

// Header implements http.ResponseWriter
func (w *writer) Header() http.Header {
	return w.rw.Header()
}

// WriteHeader implements http.ResponseWriter
func (w *writer) WriteHeader(statusCode int) {
	w.status = statusCode
}

// Write will write the response body
func (w *writer) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return len(b), nil
}

var _ http.ResponseWriter = (*writer)(nil)
