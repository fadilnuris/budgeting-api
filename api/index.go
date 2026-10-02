// Vercel serverless entrypoint. All paths are rewritten here by vercel.json.
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

	app.ServeHTTP(w, r)
}
