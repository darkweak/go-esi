package traefik

import (
	"context"
	"net/http"

	"github.com/darkweak/go-esi/writer"
)

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
	cw := writer.NewWriter(rw, req)
	e.next.ServeHTTP(cw, req)
	_ = cw.Close()
}
