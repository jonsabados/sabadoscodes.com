package auth

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

type Principal struct {
	UserID string
	Email  string
	Name   string
}

var Anonymous = Principal{
	UserID: "anonymous",
	Name:   "Anonymous User",
}

type PolicyBuilder func(ctx context.Context, principal Principal) (events.APIGatewayCustomAuthorizerPolicy, error)

func NewPolicyBuilder(baseResource string) PolicyBuilder {
	return func(ctx context.Context, principal Principal) (events.APIGatewayCustomAuthorizerPolicy, error) {
		return events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string {
						fmt.Sprintf("%s/%s/%s", baseResource, "GET", "v1/self"),
					},
				},
			},
		}, nil
	}
}

type Authenticator func(ctx context.Context, token string) (Principal, error)
