package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func ConfigureCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("Incoming request: %s %s\n", c.Request.Method, c.Request.URL.Path)

		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			fmt.Println("Handling OPTIONS request")
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
