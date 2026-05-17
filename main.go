package main

import (
	"budgeting-api/config"
	"budgeting-api/handlers"
	"budgeting-api/models"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()

	config.DB.AutoMigrate(&models.User{}, &models.Transaction{}, &models.BudgetPlan{}, &models.BudgetItem{})

	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API Running 🚀",
		})
	})

	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.POST("/transactions", handlers.CreateTransaction)
	r.GET("/transactions", handlers.GetTransactions)
	r.DELETE("/transactions/:id", handlers.DeleteTransaction)

	r.GET("/budget", handlers.GetBudget)
	r.POST("/budget", handlers.SaveBudget)
	r.POST("/budget/items", handlers.AddBudgetItem)
	r.PUT("/budget/items/:id", handlers.UpdateBudgetItem)
	r.DELETE("/budget/items/:id", handlers.DeleteBudgetItem)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
