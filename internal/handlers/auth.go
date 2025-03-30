package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sasvidu/tradesymphony/internal/config"
	"github.com/sasvidu/tradesymphony/internal/models"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Username        string `json:"username" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SessionRequest struct {
	Token string `json:"token" binding:"required"`
}

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[~!@#$%^&*()\-_+={}[\]|\\;:"<>,./?]`).MatchString(password)

	return hasUpper && hasLower && hasSpecial
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(append(salt, hash...)), nil
}

func verifyPassword(storedHash, password string) bool {
	decoded, err := hex.DecodeString(storedHash)
	if err != nil || len(decoded) != 64 {
		return false
	}

	salt := decoded[:32]
	hash := decoded[32:]

	newHash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return false
	}

	return string(hash) == string(newHash)
}

func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Validate username
	if len(req.Username) < 6 || len(req.Username) > 16 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Username does not match requirements"})
		return
	}

	// Check if username exists
	var existingUser models.User
	if err := config.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Username already exists"})
		return
	}

	// Check if email exists
	if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Email already exists"})
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

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error processing password"})
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error creating user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func Login(c *gin.Context) {
	var req LoginRequest
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

	if !verifyPassword(user.Password, req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid Password"})
		return
	}

	// Generate session token
	token, err := generateSessionToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error generating session"})
		return
	}

	// Create session
	session := models.Session{
		ID:        token,
		UserID:    user.ID,
		Expiry:    time.Now().Add(6 * time.Hour),
		UserAgent: c.GetHeader("User-Agent"),
		IP:        c.ClientIP(),
	}

	if err := config.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error creating session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"session_token": token,
	})
}

func VerifySession(c *gin.Context) {
	var req SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	var session models.Session
	if err := config.DB.Where("id = ?", req.Token).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error processing request"})
		return
	}

	// Check User-Agent and IP
	if session.UserAgent != c.GetHeader("User-Agent") || session.IP != c.ClientIP() {
		config.DB.Delete(&session)
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Conflicting User agent/IP"})
		return
	}

	// Check expiration
	if time.Now().After(session.Expiry) {
		config.DB.Delete(&session)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Session expired"})
		return
	}

	// Update expiration if within 3 hours of expiry
	if time.Until(session.Expiry) < 3*time.Hour {
		newExpiry := time.Now().Add(6 * time.Hour)
		session.Expiry = newExpiry
		if err := config.DB.Save(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error updating session"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"success":         true,
			"expirationTime": newExpiry,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
