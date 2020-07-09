package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jonsabados/sabadoscodes.com/auth"
	"github.com/jonsabados/sabadoscodes.com/httputil"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"strings"
)

func ctxLogger(ctx context.Context, baseLogger zerolog.Logger) (context.Context, zerolog.Logger, ) {
	if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
		logger := baseLogger.With().Str("requestId", awsCtx.AwsRequestID).Logger()
		return logger.WithContext(ctx), logger
	} else {
		return ctx, baseLogger
	}
}

func newHandler(baseLogger zerolog.Logger, authenticate auth.Authenticator, buildPolicy auth.PolicyBuilder) func(ctx context.Context, request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	return func(ctx context.Context, request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
		ctx, logger := ctxLogger(ctx, baseLogger)

		var principal auth.Principal
		if strings.HasPrefix(request.AuthorizationToken, "Bearer ") {
			var err error
			principal, err = authenticate(ctx, strings.Replace(request.AuthorizationToken, "Bearer ", "", 1))
			if err != nil {
				logger.Warn().Err(err).Msg("authentication failed")
				return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
			} else {
				logger.Info().Interface("principal", principal).Msg("user authenticated")
			}
		} else if request.AuthorizationToken == "anonymous" {
			principal = auth.Anonymous
		} else {
			return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
		}

		policy, err := buildPolicy(ctx, principal)
		if err != nil {
			logger.Error().Str("error", fmt.Sprintf("err")).Msg("error building policy")
			return events.APIGatewayCustomAuthorizerResponse{}, err
		}

		principalStr, err := json.Marshal(principal)
		if err != nil {
			panic(err)
		}

		return events.APIGatewayCustomAuthorizerResponse{
			PrincipalID:    principal.UserID,
			PolicyDocument: policy,
			// looks like only string values are supported in contexts - WTF?
			Context: map[string]interface{}{
				"principal": principalStr,
			},
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

	logLevel, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		panic(err)
	}
	logger := zerolog.New(os.Stdout).Level(logLevel)

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	region := os.Getenv("AWS_REGION")
	accountID := os.Getenv("ACCOUNT_ID")
	apiID := os.Getenv("API_ID")
	stage := os.Getenv("STAGE")
	clientFactory := httputil.NewXRAYAwareHTTPClientFactory(http.DefaultClient)
	certFetcher := auth.NewGoogleCertFetcher(auth.GoogleCertEndpoint, clientFactory)
	authenticator := auth.NewGoogleAuthenticator(googleClientID, certFetcher)

	lambda.Start(newHandler(logger, authenticator, auth.NewPolicyBuilder(region, accountID, apiID, stage)))
}

// so... the stuff in the SDK has actions and resources as arrays in the statement which is fucking wrong.
type APIGatewayCustomAuthorizerResponse struct {

}