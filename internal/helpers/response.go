package helpers

import (
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Success retorna una respuesta exitosa estandarizada
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error retorna una respuesta de error estandarizada
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    statusCode,
			Message: message,
		},
	})
}
