package caddy_esi

import (
	"errors"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/darkweak/go-esi/writer"
)

func init() {
	caddy.RegisterModule(ESI{})
	httpcaddyfile.RegisterGlobalOption("esi", func(h *caddyfile.Dispenser, _ interface{}) (interface{}, error) {
		return &ESI{}, nil
	})
	httpcaddyfile.RegisterHandlerDirective("esi", func(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
		e := &ESI{}
		err := e.UnmarshalCaddyfile(h.Dispenser)

		return e, err
	})
}

// ESI to handle, process and serve ESI tags.
type ESI struct {
	// DefaultHost is the origin the relative include URLs are resolved
	// against instead of the incoming request host, e.g. "backend:8080" or
	// "https://static.example.com".
	DefaultHost string `json:"default_host,omitempty"`

	opts []writer.Option
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens. Syntax:
//
//	esi {
//	    default_host <origin>
//	}
func (e *ESI) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}

		for d.NextBlock(0) {
			switch d.Val() {
			case "default_host":
				if !d.NextArg() {
					return d.ArgErr()
				}

				e.DefaultHost = d.Val()

				if d.NextArg() {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized esi subdirective %q", d.Val())
			}
		}
	}

	return nil
}

// CaddyModule returns the Caddy module information.
func (ESI) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.esi",
		New: func() caddy.Module { return new(ESI) },
	}
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (e *ESI) ServeHTTP(rw http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	cw := writer.NewWriter(rw, r, e.opts...)
	err := next.ServeHTTP(cw, r)

	return errors.Join(err, cw.Close())
}

// Provision implements caddy.Provisioner
func (e *ESI) Provision(caddy.Context) error {
	base, err := writer.ParseBaseURL(e.DefaultHost)
	if err != nil {
		return err
	}

	e.opts = []writer.Option{writer.WithBaseURL(base)}

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
	_ caddyfile.Unmarshaler       = (*ESI)(nil)
)
