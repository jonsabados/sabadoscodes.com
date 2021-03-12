package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/jonsabados/sabadoscodes.com/article"
	"github.com/jonsabados/sabadoscodes.com/auth"
	"github.com/jonsabados/sabadoscodes.com/cors"
	"github.com/jonsabados/sabadoscodes.com/dynamo"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/response"
)

func newHandler(prepLogs logging.Preparer,
	corsHeaders cors.ResponseHeaderBuilder,
	extractPrincipal auth.PrincipalExtractor,
	fetchArticle article.Fetcher) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx, _ = prepLogs(ctx)
		responseHeaders := corsHeaders(request.Headers)

		principal, err := extractPrincipal(request)
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		slug, err := url.PathUnescape(request.PathParameters["slug"])
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		a, err := fetchArticle(ctx, slug)
		if err != nil {
			return response.HandleError(ctx, responseHeaders, err), nil
		}

		if a == nil {
			return response.HandleNtFound(ctx, responseHeaders), nil
		}

		// only folks with article publish can see unpublished articles
		if a.PublishDate == nil && !principal.HasRole(auth.RoleArticlePublish) {
			return response.HandleNtFound(ctx, responseHeaders), nil
		}

		content, err := json.Marshal(a)
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
	fetcher := article.NewFetcher(dynamoClient, articleTable)

	handler := newHandler(logging.NewPreparer(), cors.NewResponseHeaderBuilder(allowedDomains), auth.NewPrincipalExtractor(), fetcher)

	lambda.Start(handler)
}
