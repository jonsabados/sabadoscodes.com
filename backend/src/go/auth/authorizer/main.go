package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jonsabados/sabadoscodes.com/auth"
	"github.com/jonsabados/sabadoscodes.com/httputil"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"net/http"
	"os"
	"strings"
)

func newHandler(prepLogs logging.Preparer, authenticate auth.Authenticator, buildPolicy auth.PolicyBuilder) func(ctx context.Context, request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	return func(ctx context.Context, request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
		ctx, logger := prepLogs(ctx)

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
			// looks like only string values are supported in contexts (or at least not complex objects)
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

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	rootUser := os.Getenv("ROOT_USER")
	region := os.Getenv("AWS_REGION")
	accountID := os.Getenv("ACCOUNT_ID")
	apiID := os.Getenv("API_ID")
	stage := os.Getenv("STAGE")
	clientFactory := httputil.NewXRAYAwareHTTPClientFactory(http.DefaultClient)
	certFetcher := auth.NewGoogleCertFetcher(auth.GoogleCertEndpoint, clientFactory)
	roleOracle := auth.NewRoleOracle(rootUser)
	authenticator := auth.NewGoogleAuthenticator(googleClientID, certFetcher, roleOracle)
	policyBuilder := auth.NewPolicyBuilder(region, accountID, apiID, stage)

	lambda.Start(newHandler(logging.NewPreparer(), authenticator, policyBuilder))
}
