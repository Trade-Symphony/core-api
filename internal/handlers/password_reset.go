package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sasvidu/tradesymphony/internal/config"
	"github.com/sasvidu/tradesymphony/internal/models"
	"gorm.io/gorm"
)

type PasswordResetRequest struct {
	Username string `json:"username" binding:"required"`
}

type PasswordResetConfirmRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid Username"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error processing request"})
		return
	}

	// Delete any existing reset tokens for this user
	config.DB.Where("user_id = ?", user.ID).Delete(&models.PasswordReset{})

	// Generate new reset token
	token, err := generateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error generating reset token"})
		return
	}

	// Create password reset record
	reset := models.PasswordReset{
		ID:        token,
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token expires in 1 hour
	}

	if err := config.DB.Create(&reset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error creating reset token"})
		return
	}

	// TODO: Send email with reset token
	// For development, we'll just print the token
	fmt.Printf("Password reset token for %s: %s\n", user.Email, token)

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func ConfirmPasswordReset(c *gin.Context) {
	var req PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Find reset token
	var reset models.PasswordReset
	if err := config.DB.Where("token = ?", req.Token).First(&reset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid Token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error processing request"})
		return
	}

	// Check if token has expired
	if time.Now().After(reset.ExpiresAt) {
		config.DB.Delete(&reset)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Expired Token"})
		return
	}

	// Validate password
	if !validatePassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Password does not match expected criteria"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Passwords do not match"})
		return
	}

	// TODO: Check if password is compromised using the Copenhagen book method

	// Hash new password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error processing password"})
		return
	}

	// Update user's password
	var user models.User
	if err := config.DB.First(&user, reset.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error finding user"})
		return
	}

	user.Password = hashedPassword
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error updating password"})
		return
	}

	// Delete the used reset token
	config.DB.Delete(&reset)

	c.JSON(http.StatusOK, gin.H{"success": true})
}
