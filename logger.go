package reqlogger

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/rs/zerolog"
	"github.com/uptrace/bunrouter"
)

func defaultLogger(out io.Writer, latency time.Duration) zerolog.Logger {
	logger := zerolog.New(out).
		Output(
			zerolog.ConsoleWriter{
				Out:     out,
				NoColor: false,
			},
		).
		With().
		Timestamp().
		Dur("latency", latency).
		Logger()

	return logger
}

//Option defines how config is set.
type Option func(*config)

// Config defines the config for logger middleware.
type config struct {
	logger func(io.Writer, time.Duration) zerolog.Logger
	// UTC a boolean stating whether to use UTC time zone or local.
	utc            bool
	skipPath       []string
	skipPathRegexp *regexp.Regexp
	// Output is a writer where logs are written.
	// Optional. Default value is os.StdOut.
	output io.Writer
	// the log level used for request with status code < 400
	defaultLevel zerolog.Level
	// the log level used for request with status code between 400 and 499
	clientErrorLevel zerolog.Level
	// the log level used for request with status code >= 500
	serverErrorLevel zerolog.Level
}

// WithLogger set custom logger func
func WithLogger(fn func(io.Writer, time.Duration) zerolog.Logger) Option {
	return func(c *config) {
		c.logger = fn
	}
}

// WithSkipPathRegexp skip URL path by regexp pattern
func WithSkipPathRegexp(s *regexp.Regexp) Option {
	return func(c *config) {
		c.skipPathRegexp = s
	}
}

// WithUTC returns t with the location set to UTC.
func WithUTC(s bool) Option {
	return func(c *config) {
		c.utc = s
	}
}

// WithSkipPath skip URL path by specfic pattern
func WithSkipPath(s []string) Option {
	return func(c *config) {
		c.skipPath = s
	}
}

// WithWriter change the default output writer.
// Default is os.StdOut.
func WithWriter(s io.Writer) Option {
	return func(c *config) {
		c.output = s
	}
}

func WithDefaultLevel(lvl zerolog.Level) Option {
	return func(c *config) {
		c.defaultLevel = lvl
	}
}

func WithClientErrorLevel(lvl zerolog.Level) Option {
	return func(c *config) {
		c.clientErrorLevel = lvl
	}
}

func WithServerErrorLevel(lvl zerolog.Level) Option {
	return func(c *config) {
		c.serverErrorLevel = lvl
	}
}

type middleware struct {
	c *config
}

// NewMiddleware creates a middleware instance ...
func NewMiddleware(opts ...Option) bunrouter.MiddlewareFunc {
	m := &middleware{}
	c := &config{
		logger:           defaultLogger,
		defaultLevel:     zerolog.InfoLevel,
		clientErrorLevel: zerolog.WarnLevel,
		serverErrorLevel: zerolog.ErrorLevel,
		output:           os.Stdout,
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(c)
	}
	m.c = c
	return m.Middleware
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	Status int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, Status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {

	rw.Status = code
	rw.ResponseWriter.WriteHeader(code)

	return
}

// Middleware ...
func (m *middleware) Middleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {

	return func(w http.ResponseWriter, req bunrouter.Request) error {

		now := time.Now()
		wrapped := wrapResponseWriter(w)
		err := next(wrapped, req)
		dur := time.Since(now)
		logger := m.c.logger(m.c.output, dur)
		msg := "Request"
		logger.WithLevel(m.c.defaultLevel).
			Int("status", wrapped.Status).
			Msg(msg)

		return err
	}
}
