package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

// APIGateway adapts the Gin engine (with routes from internal/router) to API Gateway proxy events.
func APIGateway(engine *gin.Engine) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	adapter := ginadapter.New(engine)
	return adapter.ProxyWithContext
}
