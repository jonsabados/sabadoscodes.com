package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/rs/zerolog"

	"github.com/jonsabados/sabadoscodes.com/article"
	"github.com/jonsabados/sabadoscodes.com/auth"
	"github.com/jonsabados/sabadoscodes.com/cors"
	"github.com/jonsabados/sabadoscodes.com/dynamo"
	"github.com/jonsabados/sabadoscodes.com/httputil"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/response"
)

type inboundRequest struct {
	Slug        string     `json:"slug"`
	PublishDate *time.Time `json:"publishDate"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
}

func newHandler(prepLogs logging.Preparer,
	corsHeaders cors.ResponseHeaderBuilder,
	extractPrincipal auth.PrincipalExtractor,
	fetchArticle article.Fetcher,
	saveArticle article.Saver,
	baseArticleURL string) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		ctx, _ = prepLogs(ctx)
		responseHeaders := corsHeaders(request.Headers)
		responseHeaders["content-type"] = "application/json"

		principal, err := extractPrincipal(request)
		if err != nil {
			return response.HandleError(ctx, err), nil
		}

		errors := httputil.ErrorTracker{}
		putRequest := new(inboundRequest)
		err = json.Unmarshal([]byte(request.Body), putRequest)
		if err != nil {
			zerolog.Ctx(ctx).Info().Err(err).Msg("unable to unmarshal request body")
			errors = errors.WithError("invalid request body")
			return errors.ToAPIResponse(ctx, responseHeaders), nil
		}

		if putRequest.Title == "" {
			errors = errors.WithFieldError("title", "title is required")
		}

		if putRequest.Content == "" {
			errors = errors.WithFieldError("content", "content is required")
		}

		if errors.InError() {
			return errors.ToAPIResponse(ctx, responseHeaders), nil
		}

		slug, err := url.PathUnescape(request.PathParameters["slug"])
		if err != nil {
			return response.HandleError(ctx, err), nil
		}

		existing, err := fetchArticle(ctx, slug)
		if err != nil {
			return response.HandleError(ctx, err), nil
		}

		var responseCode int
		if existing == nil {
			zerolog.Ctx(ctx).Info().Interface("user", principal).Msg("user putting new article")
			responseCode = http.StatusCreated
			responseHeaders["Location"] = fmt.Sprintf("%s/%s", baseArticleURL, slug)
		} else {
			zerolog.Ctx(ctx).Info().Interface("user", principal).Interface("original", existing).Msg("user over-writing article")
			responseCode = http.StatusNoContent
		}

		err = saveArticle(ctx, article.Article{
			Slug:        slug,
			PublishDate: putRequest.PublishDate,
			Title:       putRequest.Title,
			Content:     putRequest.Content,
		})
		if err != nil {
			return response.HandleError(ctx, err), nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: responseCode,
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
	baseArticleURL := os.Getenv("BASE_ARTICLE_URL")
	articleTable := os.Getenv("ARTICLE_TABLE")

	dynamoClient := dynamo.RawClient(sess)
	fetcher := article.NewFetcher(dynamoClient, articleTable)
	saver := article.NewSaver(dynamoClient, articleTable)

	handler := newHandler(logging.NewPreparer(), cors.NewResponseHeaderBuilder(allowedDomains), auth.NewPrincipalExtractor(), fetcher, saver, baseArticleURL)

	lambda.Start(handler)
}
