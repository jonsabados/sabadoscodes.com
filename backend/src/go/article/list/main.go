package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/jonsabados/sabadoscodes.com/article"
	"github.com/jonsabados/sabadoscodes.com/auth"
	"github.com/jonsabados/sabadoscodes.com/cors"
	"github.com/jonsabados/sabadoscodes.com/dynamo"
	"github.com/jonsabados/sabadoscodes.com/httputil"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/response"
)

func newHandler(prepLogs logging.Preparer,
	corsHeaders cors.ResponseHeaderBuilder,
	extractPrincipal auth.PrincipalExtractor,
	listArticles article.Lister) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx, _ = prepLogs(ctx)
		responseHeaders := corsHeaders(request.Headers)

		principal, err := extractPrincipal(request)
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		published := true
		if principal.HasRole(auth.RoleArticlePublish) {
			if queryParam, hasParam := request.QueryStringParameters["published"]; hasParam {
				published, err = strconv.ParseBool(queryParam)
				if err != nil {
					errors := httputil.ErrorTracker{}
					errors = errors.WithFieldError("published", "must be one of true or false")
					return errors.ToAPIResponse(ctx, responseHeaders), nil
				}
			}
		}

		articles, err := listArticles(ctx, published)
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		content, err := json.Marshal(response.ListResponse{
			Results: articles,
		})
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		responseHeaders["content-type"] = "application/json"

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    responseHeaders,
			Body:       string(content),
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
	articleTable := os.Getenv("ARTICLE_TABLE")

	dynamoClient := dynamo.RawClient(sess)
	lister := article.NewCachedLister(article.NewLister(dynamoClient, articleTable), time.Second * 15)

	handler := newHandler(logging.NewPreparer(), cors.NewResponseHeaderBuilder(allowedDomains), auth.NewPrincipalExtractor(), lister)

	lambda.Start(handler)
}
