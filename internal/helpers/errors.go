package helpers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, statusCode int, userMessage string, err error) {
	if err != nil {
		log.Printf("ERROR [%s %s]: %v", c.Request.Method, c.Request.URL.Path, err)
	}
	c.JSON(statusCode, gin.H{"error": userMessage})
}
