package roadrunner_esi

import (
	"net/http"

	"github.com/darkweak/go-esi/writer"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

const (
	configurationKey = "http.esi"
	defaultHostKey   = configurationKey + ".default_host"
)

type (
	// Configurer interface used to parse yaml configuration.
	// Implementation will be provided by the RoadRunner automatically via Init method.
	Configurer interface {
		// Get used to get config section
		Get(name string) any
		// Has checks if config section exists.
		Has(name string) bool
	}

	Plugin struct {
		opts []writer.Option
	}
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

	defaultHost, _ := cfg.Get(defaultHostKey).(string)

	base, err := writer.ParseBaseURL(defaultHost)
	if err != nil {
		return errors.E(op, err)
	}

	p.opts = []writer.Option{writer.WithBaseURL(base)}

	return nil
}

// Middleware is the request entrypoint to catch the response and
// process the esi tags if present.
func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cw := writer.NewWriter(rw, r, p.opts...)
		next.ServeHTTP(cw, r)
		_ = cw.Close()
	})
}
