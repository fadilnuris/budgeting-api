package handlers

import (
	"budgeting-api/config"
	"budgeting-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateUserInput struct {
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	ConfirmationPassword string `json:"confirmation_password"`
}

func CreateUser(c *gin.Context) {
	var input CreateUserInput

	// bind JSON dari request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// mapping ke model
	user := models.User{
		Name:                 input.Name,
		Email:                input.Email,
		Password:             input.Password,
		ConfirmationPassword: input.ConfirmationPassword,
	}

	if input.Password != input.ConfirmationPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password and confirmation password do not match",
		})
		return
	}

	// simpan ke database
	result := config.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	// response sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "User created",
		"data":    user,
	})
}
