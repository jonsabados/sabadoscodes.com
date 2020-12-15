package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/rs/zerolog"

	"github.com/jonsabados/sabadoscodes.com/article/assets"
	"github.com/jonsabados/sabadoscodes.com/auth"
	"github.com/jonsabados/sabadoscodes.com/cors"
	"github.com/jonsabados/sabadoscodes.com/httputil"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/response"
	"github.com/jonsabados/sabadoscodes.com/s3"
)

const assetCacheDuration = time.Hour * 24 * 365

type inboundRequest struct {
	Path     string `json:"path"`
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
}

func newHandler(prepLogs logging.Preparer,
	corsHeaders cors.ResponseHeaderBuilder,
	extractPrincipal auth.PrincipalExtractor,
	targetBucket string,
	saveObject s3.PublicObjectSaver,
	baseAssetURL string) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx, _ = prepLogs(ctx)
		responseHeaders := corsHeaders(request.Headers)
		responseHeaders["content-type"] = "application/json"

		principal, err := extractPrincipal(request)
		if err != nil {
			return response.HandleError(ctx, err), nil
		}

		errors := httputil.ErrorTracker{}
		uploadRequest := new(inboundRequest)
		err = json.Unmarshal([]byte(request.Body), uploadRequest)
		if err != nil {
			zerolog.Ctx(ctx).Info().Err(err).Msg("unable to unmarshal request body")
			errors = errors.WithError("invalid request body")
			return errors.ToAPIResponse(ctx, responseHeaders), nil
		}

		if uploadRequest.Path == "" {
			errors = errors.WithFieldError("path", "path is required")
		}

		if uploadRequest.MimeType == "" {
			errors = errors.WithFieldError("mimeType", "mimeType is required")
		}

		var content []byte
		if uploadRequest.Content == "" {
			errors = errors.WithFieldError("content", "content is required")
		} else {
			content, err = base64.StdEncoding.DecodeString(uploadRequest.Content)
			if err != nil {
				zerolog.Ctx(ctx).Info().Err(err).Msg("unable to decode content")
				errors = errors.WithFieldError("content", "content must be bas64 encoded")
			}
		}

		if errors.InError() {
			return errors.ToAPIResponse(ctx, responseHeaders), nil
		}

		zerolog.Ctx(ctx).Info().Interface("user", principal).Msg("user uploading object")
		// needs to be under article-asset in the bucket. If it ever needs to become configurable will deal with it
		path := fmt.Sprintf("%s%s", assets.AssetKeyPrefix, uploadRequest.Path)
		err = saveObject(ctx, targetBucket, path, bytes.NewReader(content), uploadRequest.MimeType, assetCacheDuration)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("upload failed")
			return events.APIGatewayProxyResponse{}, err
		}
		responseHeaders["Location"] = fmt.Sprintf("%s/%s", baseAssetURL, uploadRequest.Path)

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusCreated,
			Headers:    responseHeaders,
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
	saver := s3.NewPublicObjectSaver(s3Client)

	handler := newHandler(logging.NewPreparer(), cors.NewResponseHeaderBuilder(allowedDomains), auth.NewPrincipalExtractor(), targetBucket, saver, baseAssetURL)

	lambda.Start(handler)
}
