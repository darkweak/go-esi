package caddy_esi

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/darkweak/go-esi/esi"
)

var bufPool *sync.Pool = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func init() {
	caddy.RegisterModule(ESI{})
	httpcaddyfile.RegisterGlobalOption("esi", func(h *caddyfile.Dispenser, _ interface{}) (interface{}, error) {
		return httpcaddyfile.App{
			Name:  "http.handlers.esi",
			Value: caddyconfig.JSON(ESI{}, nil),
		}, nil
	})
	httpcaddyfile.RegisterHandlerDirective("esi", func(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
		return &ESI{}, nil
	})
}

// ESI to handle, process and serve ESI tags.
type ESI struct{}

// CaddyModule returns the Caddy module information.
func (ESI) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.esi",
		New: func() caddy.Module { return new(ESI) },
	}
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (e *ESI) ServeHTTP(rw http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	cw := newWriter(bufPool.Get().(*bytes.Buffer), rw)
	next.ServeHTTP(cw, r)

	b := esi.Parse(cw.buf.Bytes(), r)

	rw.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	rw.WriteHeader(cw.status)
	_, _ = rw.Write(b)

	return nil
}

// Provision implements caddy.Provisioner
func (*ESI) Provision(caddy.Context) error {
	return nil
}

func (s ESI) Start() error { return nil }

func (s ESI) Stop() error { return nil }

// Interface guards
var (
	_ caddyhttp.MiddlewareHandler = (*ESI)(nil)
	_ caddy.Module                = (*ESI)(nil)
	_ caddy.Provisioner           = (*ESI)(nil)
	_ caddy.App                   = (*ESI)(nil)
)
