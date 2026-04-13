package health

import (
	"net/http"

	"github.com/ayushpandey15/lambda-go/internal/router"
	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(Health)
}

func Health(rg *gin.RouterGroup) {
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
