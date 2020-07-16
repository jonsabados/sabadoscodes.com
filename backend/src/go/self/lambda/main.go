package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jonsabados/sabadoscodes.com/auth"
	"github.com/jonsabados/sabadoscodes.com/cors"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/response"
	"net/http"
	"os"
	"strings"
)

func newHandler(prepLogs logging.Preparer, corsHeaders cors.ResponseHeaderBuilder, extractPrincipal auth.PrincipalExtractor) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx, _ = prepLogs(ctx)
		responseHeaders := corsHeaders(request.Headers)
		responseHeaders["content-type"] = "application/json"

		principal, err := extractPrincipal(request)
		if err != nil {
			return response.HandleError(ctx, err), nil
		}

		responseBody, err := json.Marshal(principal)
		if err != nil {
			return response.HandleError(ctx, err), nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusOK,
			Headers:         responseHeaders,
			Body:            string(responseBody),
			IsBase64Encoded: false,
		}, nil
	}
}

func main() {
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	lambda.Start(newHandler(logging.NewPreparer(), cors.NewResponseHeaderBuilder(allowedDomains), auth.NewPrincipalExtractor()))
}
