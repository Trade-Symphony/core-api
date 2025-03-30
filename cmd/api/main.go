package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sasvidu/tradesymphony/internal/config"
	"github.com/sasvidu/tradesymphony/internal/handlers"
	"github.com/sasvidu/tradesymphony/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	config.InitDB()

	// Create router
	r := gin.Default()

	// Enable CORS for all origins
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, User-Agent, X-Forwarded-For")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", rateLimiter.RateLimit(), handlers.Register)
		auth.POST("/login", rateLimiter.RateLimit(), handlers.Login)
		auth.POST("/session", rateLimiter.RateLimit(), handlers.VerifySession)
	}

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
