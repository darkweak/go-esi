package roadrunner

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/darkweak/go-esi/writer"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

var bufPool *sync.Pool = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

const configurationKey = "http.esi"

type (
	// Configurer interface used to parse yaml configuration.
	// Implementation will be provided by the RoadRunner automatically via Init method.
	Configurer interface {
		// Get used to get config section
		Get(name string) any
		// Has checks if config section exists.
		Has(name string) bool
	}

	Plugin struct{}
)

// Name is the plugin name
func (p *Plugin) Name() string {
	return "esi"
}

// Init allows the user to set up an efficient esi processor.
func (p *Plugin) Init(cfg Configurer, log *zap.Logger) error {
	const op = errors.Op("esi_middleware_init")
	if !cfg.Has(configurationKey) {
		return errors.E(op, errors.Disabled)
	}

	return nil
}

// Middleware is the request entrypoint to catch the response and
// process the esi tags if present.
func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufPool.Put(buf)
		cw := writer.NewWriter(buf, rw, r)
		go func(w *writer.Writer) {
			var i = 0
			for {
				if len(w.AsyncBuf) <= i {
					continue
				}
				rs := <-w.AsyncBuf[i]
				if rs == nil {
					w.Done <- true
					break
				}
				_, _ = rw.Write(rs)
				i++
			}
		}(cw)
		next.ServeHTTP(cw, r)
		cw.Header().Del("Content-Length")
		if cw.Rq.ProtoMajor == 1 {
			cw.Header().Set("Content-Encoding", "chunked")
		}
		cw.AsyncBuf = append(cw.AsyncBuf, make(chan []byte))
		go func(w *writer.Writer, iteration int) {
			w.AsyncBuf[iteration] <- nil
		}(cw, cw.Iteration)

		<-cw.Done
	})
}
