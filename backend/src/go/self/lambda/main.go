package main

import (
	"context"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"net/http"
)

func newHandler() func (ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		encodedPrincipal := request.RequestContext.Authorizer["principal"]
		principalString, err := base64.StdEncoding.DecodeString(encodedPrincipal.(string))
		if err != nil {
			panic(err)
		}

		return events.APIGatewayProxyResponse{
			StatusCode:        http.StatusOK,
			Headers:           map[string]string {
				"content-type": "application/json",
			},
			Body:              string(principalString),
			IsBase64Encoded:   false,
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

	lambda.Start(newHandler())
}