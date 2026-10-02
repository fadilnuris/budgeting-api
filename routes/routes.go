package routes

import (
	"budgeting-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API Running 🚀",
		})
	})

	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.GET("/profile", handlers.GetProfile)
	r.POST("/transactions", handlers.CreateTransaction)
	r.GET("/transactions", handlers.GetTransactions)
	r.DELETE("/transactions/:id", handlers.DeleteTransaction)

	r.GET("/budget", handlers.GetBudget)
	r.POST("/budget", handlers.SaveBudget)
	r.POST("/budget/items", handlers.AddBudgetItem)
	r.PUT("/budget/items/:id", handlers.UpdateBudgetItem)
	r.DELETE("/budget/items/:id", handlers.DeleteBudgetItem)

	return r
}
