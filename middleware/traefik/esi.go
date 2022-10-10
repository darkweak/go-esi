package traefik

import (
	"bytes"
	"context"
	"net/http"
	"sync"

	"github.com/darkweak/go-esi/writer"
)

var bufPool *sync.Pool = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

// Config the ESI plugin configuration.
type Config struct{}

// CreateConfig creates the ESI plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// ESI is a plugin that allow users to process the ESI tags.
type ESI struct {
	next http.Handler
	name string
}

// New created a new ESI plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &ESI{
		next: next,
		name: name,
	}, nil
}

func (e *ESI) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	cw := writer.NewWriter(buf, rw, req)
	go func(w *writer.Writer) {
		w.Header().Del("Content-Length")
		if w.Rq.ProtoMajor == 1 {
			w.Header().Set("Content-Encoding", "chunked")
		}
		var i = 0
		for {
			if len(cw.AsyncBuf) <= i {
				continue
			}
			rs := <-cw.AsyncBuf[i]
			if rs == nil {
				cw.Done <- true
				break
			}
			_, _ = rw.Write(rs)
			i++
		}
	}(cw)
	e.next.ServeHTTP(cw, req)
	cw.AsyncBuf = append(cw.AsyncBuf, make(chan []byte))
	go func(w *writer.Writer, iteration int) {
		w.AsyncBuf[iteration] <- nil
	}(cw, cw.Iteration)

	<-cw.Done
}
