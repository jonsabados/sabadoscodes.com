package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"net/http"
	"os"
	"strings"
)

func newHandler(allowedDomains []string) func (ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		headers := make(map[string]string)
		origin, hasOrigin := request.Headers["Origin"]
		if hasOrigin && isOriginAllowed(origin, allowedDomains) {
			headers["Access-Control-Allow-Origin"] = origin
			headers["Access-Control-Allow-Headers"] = "Authorization"
			headers["Access-Control-Allow-Methods"] = "OPTIONS,HEAD,GET,POST,PUT,DELETE"
		}
		return events.APIGatewayProxyResponse{
			StatusCode:        http.StatusOK,
			Headers:           headers,
			Body:              "",
			IsBase64Encoded:   false,
		}, nil
	}
}

func isOriginAllowed(origin string, allowedDomains []string) bool {
	for _, o := range allowedDomains {
		if origin == o {
			return true
		}
	}
	return false
}

func main() {
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	lambda.Start(newHandler(allowedDomains))
}