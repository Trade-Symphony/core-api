package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckAuth is a handler to check if a user is authenticated
func CheckAuth(c *gin.Context) {
	c.String(http.StatusOK, "You are authenticated!")
}