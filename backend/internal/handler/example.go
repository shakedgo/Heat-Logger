package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "List of users"})
}

func CreateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func HealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
