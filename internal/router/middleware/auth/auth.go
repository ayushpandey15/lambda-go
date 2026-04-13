package auth

import (
	"strings"

	"github.com/ayushpandey15/lambda-go/internal/pkg/constant"
	"github.com/ayushpandey15/lambda-go/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func AuthCheck(c *gin.Context) {
	ip := strings.TrimSpace(c.ClientIP())

	var isIpWhiteListed bool

	for _, whitelistIp := range constant.IPWhitelist {
		if strings.TrimSpace(whitelistIp) == ip {
			isIpWhiteListed = true
			break
		}
	}

	if !isIpWhiteListed {
		response.WriteErrorResponse(c, response.ErrUnauthorized.WithMessage("Unauthorized access"))
		return
	}

	c.Next()
}
