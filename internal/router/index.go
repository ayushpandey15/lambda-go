// Package router provides Register/Setup for Gin groups.
// Subpackages (e.g. internal/router/health) call Register in init; blank-import them from main.
package router

import (
	"time"

	"github.com/gin-gonic/gin"
)

var routes []func(rg *gin.RouterGroup)

// ApplicationStartTime is set from main before the server or Lambda handler serves traffic.
var ApplicationStartTime time.Time

func Register(fn func(rg *gin.RouterGroup)) {
	routes = append(routes, fn)
}

func Setup(rg *gin.RouterGroup) {
	for _, fn := range routes {
		fn(rg)
	}
}
