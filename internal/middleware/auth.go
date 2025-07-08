package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

// AuthMiddleware is a middleware to verify Firebase ID tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		credentialsJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")
		if credentialsJSON == "" {
			log.Println("FIREBASE_CREDENTIALS_JSON environment variable not set")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Firebase credentials not configured"})
			return
		}
		opt := option.WithCredentialsJSON([]byte(credentialsJSON))

		projectID := os.Getenv("FIREBASE_PROJECT_ID")
		if projectID == "" {
			log.Println("FIREBASE_PROJECT_ID environment variable not set")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Firebase project ID not configured"})
			return
		}
		config := &firebase.Config{
			ProjectID: projectID,
		}

		app, err := firebase.NewApp(context.Background(), config, opt)
		if err != nil {
			log.Printf("error initializing app: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error initializing app"})
			return
		}

		client, err := app.Auth(context.Background())
		if err != nil {
			log.Printf("error getting Auth client: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error getting Auth client"})
			return
		}


		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		token := strings.Replace(authHeader, "Bearer ", "", 1)
		decodedToken, err := client.VerifyIDToken(context.Background(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Add user to context
		c.Set("user", decodedToken)
		c.Next()
	}
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(c *gin.Context) *auth.Token {
	if user, exists := c.Get("user"); exists {
		if token, ok := user.(*auth.Token); ok {
			return token
		}
	}
	return nil
}