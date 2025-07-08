package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sasvidu/tradesymphony/internal/handlers"
	"github.com/sasvidu/tradesymphony/internal/middleware"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

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

	r.GET("/health", handlers.HealthCheck)

	auth := r.Group("/auth")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/check", handlers.CheckAuth)

	// Start server
	if err := r.Run(":5005"); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
