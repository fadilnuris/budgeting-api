package main

import (
	"budgeting-api/config"
	"budgeting-api/models"
	"budgeting-api/routes"
	"os"
)

func main() {
	config.ConnectDB()

	// Drop the old single-column unique index so GORM can recreate it as a composite (user_id, month) unique index
	config.DB.Exec("DROP INDEX IF EXISTS idx_user_month")

	config.DB.AutoMigrate(&models.User{}, &models.Transaction{}, &models.BudgetPlan{}, &models.BudgetItem{}, &models.BudgetCategory{})

	r := routes.Setup()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
