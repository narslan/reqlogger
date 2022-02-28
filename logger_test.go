package reqlogger

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/uptrace/bunrouter"
)

func TestMiddleware(t *testing.T) {
	router := bunrouter.New(
		bunrouter.WithMiddleware(NewLoggingMiddleware()),
	)
	router.GET("/example", func(w http.ResponseWriter, req bunrouter.Request) error {
		return nil
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/example", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Error("status not ok", w.Code)
	}
}

func TestLoggerWithLevels(t *testing.T) {
	//buffer := new(bytes.Buffer)

	router := bunrouter.New(
		bunrouter.WithMiddleware(NewLoggingMiddleware(
			WithWriter(os.Stdout),
			WithDefaultLevel(zerolog.DebugLevel),
			WithClientErrorLevel(zerolog.ErrorLevel),
			WithServerErrorLevel(zerolog.FatalLevel),
		)))

	router.GET("/example", func(w http.ResponseWriter, req bunrouter.Request) error {
		return nil
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/example", nil)
	router.ServeHTTP(w, req)
}
