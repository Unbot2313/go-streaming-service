package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/helpers"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

var authService = services.NewAuthService()

func AuthMiddleware(c *gin.Context) {
	rawToken := c.GetHeader("Authorization")

	if rawToken == "" {
		c.JSON(401, gin.H{"error": "Authorization header not provided"})
		c.Abort()
		return
	}

	if !strings.HasPrefix(rawToken, "Bearer ") {
		c.JSON(401, gin.H{"error": "Invalid authorization format. Use: Bearer <token>"})
		c.Abort()
		return
	}

	token := strings.TrimPrefix(rawToken, "Bearer ")

	user, err := authService.ValidateToken(token)

	if err != nil {
		helpers.HandleError(c, 401, "Invalid or expired token", err)
		c.Abort()
		return
	}

	// Guardar el usuario autenticado en el contexto de Gin
	c.Set("user", user)


	c.Next()
}