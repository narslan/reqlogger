package reqlogger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bunrouter"
)

func TestRequestWithContext(t *testing.T) {
	router := bunrouter.New()
	router.GET("/user/:param", func(w http.ResponseWriter, req bunrouter.Request) error {
		value1 := req.Param("param")
		require.Equal(t, "hello", value1)

		value2 := req.WithContext(context.TODO()).Param("param")
		require.Equal(t, value1, value2)

		return nil
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/hello", nil)
	router.ServeHTTP(w, req)
}
