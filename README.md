#### reqlogger
[bunrouter](https://github.com/uptrace/bunrouter) middleware for request logging using [rs/zerolog](https://github.com/rs/zerolog).

### Example

```go
package main

import (
	"net/http"
	"os"
	"github.com/narslan/reqlogger"
	"github.com/uptrace/bunrouter"
)

func main() {
	router := bunrouter.New(
		bunrouter.Use(reqlogger.NewLoggingMiddleware()),
	)
	router.GET("/", handler)
	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func handler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route":  req.Route(),
	})
}

```

