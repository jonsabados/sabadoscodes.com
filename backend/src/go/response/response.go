package response

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
	"net/http"
)

type ListResponse struct {
	Results []interface{} `json:"results"`
}

type ErrorResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func HandleError(ctx context.Context, responseHeaders map[string]string, err error) events.APIGatewayProxyResponse {
	zerolog.Ctx(ctx).Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error encountered")

	responseBody := ErrorResponse{
		Message: "an error has occurred",
	}

	if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
		responseBody.RequestID = awsCtx.AwsRequestID
	}

	content, err := json.Marshal(responseBody)
	if err != nil {
		panic(err)
	}

	responseHeaders["content-type"] = "application/json"

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       string(content),
		Headers:    responseHeaders,
	}
}

func HandleNtFound(ctx context.Context, responseHeaders map[string]string) events.APIGatewayProxyResponse {
	responseBody := ErrorResponse{
		Message: "requested entity not found",
	}

	if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
		responseBody.RequestID = awsCtx.AwsRequestID
	}

	content, err := json.Marshal(responseBody)
	if err != nil {
		panic(err)
	}

	responseHeaders["content-type"] = "application/json"

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Body:       string(content),
		Headers:    responseHeaders,
	}
}
