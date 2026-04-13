package server

import (
	"github.com/ayushpandey15/lambda-go/internal/router"
	"github.com/gin-gonic/gin"
)

// NewEngine builds the Gin app used for both local HTTP and Lambda; routes register under /lambda-go.
func NewEngine() *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/lambda-go")
	router.Setup(v1)
	return r
}
