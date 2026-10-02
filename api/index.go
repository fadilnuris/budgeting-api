// Vercel serverless entrypoint. All paths are rewritten here by vercel.json,
// with the original path passed in the __path query param.
// Migrations are not run here; run `go run .` locally against the same DATABASE_URL when models change.
package handler

import (
	"budgeting-api/config"
	"budgeting-api/routes"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	once sync.Once
	app  *gin.Engine
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		config.ConnectDB()
		app = routes.Setup()
	})

	q := r.URL.Query()
	if p := q.Get("__path"); p != "" {
		r.URL.Path = p
		r.URL.RawPath = ""
		q.Del("__path")
		r.URL.RawQuery = q.Encode()
	}

	app.ServeHTTP(w, r)
}
