package reqlogger

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/rs/zerolog"
	"github.com/uptrace/bunrouter"
)

func defaultLogger(
	out io.Writer, latency time.Duration, code int, method, path string,
) zerolog.Logger {
	logger := zerolog.New(out).
		Output(
			zerolog.ConsoleWriter{
				Out:     out,
				NoColor: false,
			},
		).
		With().
		Timestamp().
		Int("status", code).
		Dur("latency", latency).
		Str("method", method).
		Str("path", path).
		Logger()

	return logger
}

//Option defines how config is set.
type Option func(*config)

// Config defines the config for logger middleware.
type config struct {
	logger func(io.Writer, time.Duration, int, string, string) zerolog.Logger
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
func WithLogger(fn func(io.Writer, time.Duration, int, string, string) zerolog.Logger) Option {
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

// NewLoggingMiddleware creates a middleware instance ...
func NewLoggingMiddleware(opts ...Option) bunrouter.MiddlewareFunc {
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
	status int
	size   int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	return
}

// (rw *responseWriter) Status ...
func (rw *responseWriter) Status() int {
	return rw.status
}

// (rw *responseWriter) Size ...
func (rw *responseWriter) Size() int {
	return rw.size
}

// Hijack implements the http.Hijacker interface.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.size < 0 {
		rw.size = 0
	}
	return rw.ResponseWriter.(http.Hijacker).Hijack()
}

// Middleware ...
func (m *middleware) Middleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {

	return func(w http.ResponseWriter, req bunrouter.Request) error {

		now := time.Now()
		wrapped := wrapResponseWriter(w)
		err := next(wrapped, req)
		dur := time.Since(now)
		code := wrapped.Status()
		logger := m.c.logger(m.c.output, dur, code, req.Method, req.URL.String())
		msg := "Request"

		switch {
		case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
			logger.WithLevel(m.c.clientErrorLevel).Msg(msg)
		case code >= http.StatusInternalServerError:
			logger.WithLevel(m.c.serverErrorLevel).Msg(msg)
		default:
			logger.WithLevel(m.c.defaultLevel).Msg(msg)
		}

		return err
	}
}
