package auth

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

type Principal struct {
	UserID string   `json:"userId"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
}

var Anonymous = Principal{
	UserID: "anonymous",
	Name:   "Anonymous User",
}

type PolicyBuilder func(ctx context.Context, principal Principal) (events.APIGatewayCustomAuthorizerPolicy, error)

func NewPolicyBuilder(region string, accountID string, apiID string, stage string) PolicyBuilder {
	return func(ctx context.Context, principal Principal) (events.APIGatewayCustomAuthorizerPolicy, error) {
		return events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action: []string{"execute-api:Invoke"},
					Effect: "Allow",
					Resource: []string{
						fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/%s/%s", region, accountID, apiID, stage, "GET", "self"),
					},
				},
			},
		}, nil
	}
}

type Authenticator func(ctx context.Context, token string) (Principal, error)
