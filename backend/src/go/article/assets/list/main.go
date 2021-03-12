package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/jonsabados/sabadoscodes.com/article/assets"
	"github.com/jonsabados/sabadoscodes.com/cors"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/response"
	"github.com/jonsabados/sabadoscodes.com/s3"
)

type assetDetails struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}

func newHandler(prepLogs logging.Preparer,
	corsHeaders cors.ResponseHeaderBuilder,
	targetBucket string,
	listObjects s3.ObjectLister,
	baseAssetURL string) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx, _ = prepLogs(ctx)
		responseHeaders := corsHeaders(request.Headers)

		objects, err := listObjects(ctx, targetBucket, assets.AssetKeyPrefix)
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		results := make([]interface{}, 0)
		for _, o := range objects {
			path := strings.TrimPrefix(o.Path, assets.AssetKeyPrefix)
			url := fmt.Sprintf("%s/%s", baseAssetURL, path)
			results = append(results, assetDetails{path, o.Size, url})
		}

		responseBody, err := json.Marshal(response.ListResponse{Results: results})
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		responseHeaders["content-type"] = "application/json"

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

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		panic(err)
	}

	allowedDomains := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	targetBucket := os.Getenv("ASSET_BUCKET")
	baseAssetURL := os.Getenv("BASE_ASSET_URL")

	s3Client := s3.RawClient(sess)
	lister := s3.NewObjectLister(s3Client)

	handler := newHandler(logging.NewPreparer(), cors.NewResponseHeaderBuilder(allowedDomains), targetBucket, lister, baseAssetURL)

	lambda.Start(handler)
}
