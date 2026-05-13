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

	config.DB.AutoMigrate(&models.User{}, &models.Transaction{})

	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API Running 🚀",
		})
	})

	r.POST("/users", handlers.CreateUser)
	r.POST("/transactions", handlers.CreateTransaction)
	r.GET("/transactions", handlers.GetTransactions)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
