package reqlogger

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bunrouter"
)

type middleware struct{}

// NewMiddleware creates a middleware instance ...
func NewMiddleware() bunrouter.MiddlewareFunc {
	m := &middleware{}
	return m.Middleware
}

// Middleware ...
func (m *middleware) Middleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {

	return func(w http.ResponseWriter, req bunrouter.Request) error {
		now := time.Now()
		err := next(w, req)
		dur := time.Since(now)

		log.Info().Dur("latency", dur).Msg("")
		return err
	}
}
