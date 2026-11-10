package traefik

import (
	"context"
	"net/http"

	"github.com/darkweak/go-esi/writer"
)

// Config the ESI plugin configuration.
type Config struct {
	// DefaultHost is the origin the relative include URLs are resolved
	// against instead of the incoming request host, e.g.
	// "static-webserver.my-app.svc.cluster.local" or "https://static.example.com".
	DefaultHost string `json:"defaultHost,omitempty" toml:"defaultHost,omitempty" yaml:"defaultHost,omitempty"`
}

// CreateConfig creates the ESI plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// ESI is a plugin that allow users to process the ESI tags.
type ESI struct {
	next http.Handler
	name string
	opts []writer.Option
}

// New created a new ESI plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	base, err := writer.ParseBaseURL(config.DefaultHost)
	if err != nil {
		return nil, err
	}

	return &ESI{
		next: next,
		name: name,
		opts: []writer.Option{writer.WithBaseURL(base)},
	}, nil
}

func (e *ESI) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	cw := writer.NewWriter(rw, req, e.opts...)
	e.next.ServeHTTP(cw, req)
	_ = cw.Close()
}
