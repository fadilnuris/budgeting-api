package handlers

import (
	"budgeting-api/config"
	"budgeting-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateTransactionInput struct {
	UserID uint   `json:"user_id"`
	Type   string `json:"type"` // income / expense
	Amount int    `json:"amount"`
	Note   string `json:"note"`
	Date   string `json:"date"`
}

func CreateTransaction(c *gin.Context) {
	var input CreateTransactionInput

	// bind JSON dari request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// mapping ke model
	transaction := models.Transaction{
		UserID: input.UserID,
		Type:   input.Type,
		Amount: input.Amount,
		Note:   input.Note,
		Date:   input.Date,
	}

	// simpan ke database
	result := config.DB.Create(&transaction)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	// response sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction created",
		"data":    transaction,
	})
}

func GetTransactions(c *gin.Context) {
	userID := c.Query("user_id")
	var transactions []models.Transaction

	query := config.DB
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	result := query.Find(&transactions)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transactions,
	})
}

func DeleteTransaction(c *gin.Context) {
	id := c.Param("id")
	var transaction models.Transaction

	// Cek apakah transaksi ada
	if err := config.DB.First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Transaction not found",
		})
		return
	}

	// Hapus transaksi
	if err := config.DB.Delete(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction deleted successfully",
	})
}
