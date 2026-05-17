package handlers

import (
	"budgeting-api/config"
	"budgeting-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SaveBudgetInput struct {
	UserID     uint                    `json:"user_id"`
	Month      string                  `json:"month"`
	Income     int                     `json:"income"`
	Items      []models.BudgetItem     `json:"items"`
	Categories []models.BudgetCategory `json:"categories"`
}

func SaveBudget(c *gin.Context) {
	var input SaveBudgetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plan models.BudgetPlan
	err := config.DB.Where("user_id = ? AND month = ?", input.UserID, input.Month).First(&plan).Error

	if err != nil {
		plan = models.BudgetPlan{
			UserID: input.UserID,
			Month:  input.Month,
			Income: input.Income,
		}
		if err := config.DB.Create(&plan).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		plan.Income = input.Income
		config.DB.Save(&plan)
		config.DB.Where("plan_id = ?", plan.ID).Delete(&models.BudgetItem{})
		config.DB.Where("plan_id = ?", plan.ID).Delete(&models.BudgetCategory{})
	}

	for i := range input.Items {
		input.Items[i].PlanID = plan.ID
		input.Items[i].ID = 0
	}
	
	if len(input.Items) > 0 {
		if err := config.DB.Create(&input.Items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	for i := range input.Categories {
		input.Categories[i].PlanID = plan.ID
		input.Categories[i].ID = 0
	}
	
	if len(input.Categories) > 0 {
		if err := config.DB.Create(&input.Categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Budget saved successfully",
		"data":    plan,
	})
}

func GetBudget(c *gin.Context) {
	userID := c.Query("user_id")
	month := c.Query("month")

	var plan models.BudgetPlan
	err := config.DB.Preload("Items").Preload("Categories").Where("user_id = ? AND month = ?", userID, month).First(&plan).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "No budget found for this month",
			"data":    nil,
		})
		return
	}

	if len(plan.Categories) == 0 {
		defaults := []models.BudgetCategory{
			{PlanID: plan.ID, Name: "Tabungan", Color: "blue"},
			{PlanID: plan.ID, Name: "Kebutuhan", Color: "amber"},
			{PlanID: plan.ID, Name: "Keinginan", Color: "pink"},
		}
		config.DB.Create(&defaults)
		plan.Categories = defaults
	}

	c.JSON(http.StatusOK, gin.H{
		"data": plan,
	})
}

// Item Endpoints

type AddBudgetItemInput struct {
	UserID   uint   `json:"user_id"`
	Month    string `json:"month"`
	Name     string `json:"name"`
	Nominal  int    `json:"nominal"`
	Category string `json:"category"`
}

func AddBudgetItem(c *gin.Context) {
	var input AddBudgetItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plan models.BudgetPlan
	err := config.DB.Where("user_id = ? AND month = ?", input.UserID, input.Month).First(&plan).Error
	if err != nil {
		plan = models.BudgetPlan{
			UserID: input.UserID,
			Month:  input.Month,
			Income: 0,
		}
		if err := config.DB.Create(&plan).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	item := models.BudgetItem{
		PlanID:   plan.ID,
		Name:     input.Name,
		Nominal:  input.Nominal,
		Category: input.Category,
	}

	if err := config.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item added successfully",
		"data":    item,
	})
}

type UpdateBudgetItemInput struct {
	Name     string `json:"name"`
	Nominal  int    `json:"nominal"`
	Category string `json:"category"`
}

func UpdateBudgetItem(c *gin.Context) {
	id := c.Param("id")
	var input UpdateBudgetItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item models.BudgetItem
	if err := config.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	item.Name = input.Name
	item.Nominal = input.Nominal
	item.Category = input.Category

	if err := config.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item updated successfully",
		"data":    item,
	})
}

func DeleteBudgetItem(c *gin.Context) {
	id := c.Param("id")
	var item models.BudgetItem
	
	if err := config.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if err := config.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item deleted successfully",
	})
}
