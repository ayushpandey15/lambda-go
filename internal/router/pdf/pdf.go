package pdf

import (
	handlerpdf "github.com/ayushpandey15/lambda-go/internal/handler/pdf"
	"github.com/ayushpandey15/lambda-go/internal/router"
	"github.com/gin-gonic/gin"
)

func init() {
	router.Register(Routes)
}

func Routes(rg *gin.RouterGroup) {
	rg.POST("/html-to-pdf", handlerpdf.HTMLToPDF)
}
