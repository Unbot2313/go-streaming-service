package helpers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, statusCode int, userMessage string, err error) {
	if err != nil {
		slog.Error("request error",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Any("error", err),
		)
	}
	Error(c, statusCode, userMessage)
}
