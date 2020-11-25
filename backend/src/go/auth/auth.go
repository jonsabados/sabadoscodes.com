package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
)

type Principal struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Roles  []Role `json:"roles"`
}

var Anonymous = Principal{
	UserID: "anonymous",
	Name:   "Anonymous User",
}

type PolicyBuilder func(ctx context.Context, principal Principal) (events.APIGatewayCustomAuthorizerPolicy, error)

func NewPolicyBuilder(region string, accountID string, apiID string, stage string) PolicyBuilder {
	return func(ctx context.Context, principal Principal) (events.APIGatewayCustomAuthorizerPolicy, error) {
		statement := []events.IAMPolicyStatement{
			{
				Action: []string{"execute-api:Invoke"},
				Effect: "Allow",
				Resource: []string{
					fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/%s/%s", region, accountID, apiID, stage, "GET", "self"),
				},
			},
		}
		for _, r := range principal.Roles {
			switch r {
			case RoleAssetPublish:
				statement = append(statement, events.IAMPolicyStatement{
					Action: []string{"execute-api:Invoke"},
					Effect: "Allow",
					Resource: []string{
						fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/%s/%s", region, accountID, apiID, stage, "POST", "article/asset"),
					},
				})
			}
		}
		return events.APIGatewayCustomAuthorizerPolicy{
			Version:   "2012-10-17",
			Statement: statement,
		}, nil
	}
}

type Authenticator func(ctx context.Context, token string) (Principal, error)

type PrincipalExtractor func(request events.APIGatewayProxyRequest) (Principal, error)

func NewPrincipalExtractor() PrincipalExtractor {
	return func(request events.APIGatewayProxyRequest) (Principal, error) {
		encodedPrincipal := request.RequestContext.Authorizer["principal"]
		principal, err := base64.StdEncoding.DecodeString(encodedPrincipal.(string))
		if err != nil {
			return Principal{}, errors.WithStack(err)
		}
		ret := Principal{}
		err = json.Unmarshal(principal, &ret)
		if err != nil {
			return Principal{}, errors.WithStack(err)
		}
		return ret, nil
	}
}
