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
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	config.InitDB()

	// Create router
	r := gin.Default()

	// Enable CORS for all origins
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, User-Agent, X-Forwarded-For, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Public endpoints (no API key required)
	r.GET("/health", handlers.HealthCheck)

	// Protected endpoints
	protected := r.Group("/")
	protected.Use(middleware.APIKeyAuth())
	{
		// Auth routes
		auth := protected.Group("/auth")
		{
			auth.POST("/register", rateLimiter.RateLimit(), handlers.Register)
			auth.POST("/login", rateLimiter.RateLimit(), handlers.Login)
			auth.POST("/session", rateLimiter.RateLimit(), handlers.VerifySession)
			auth.POST("/password-reset", rateLimiter.RateLimit(), handlers.RequestPasswordReset)
			auth.PATCH("/password-reset", rateLimiter.RateLimit(), handlers.ConfirmPasswordReset)
		}
	}

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
