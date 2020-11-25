package auth

import (
	"context"
	"strings"
)

type Role string

const (
	RoleAssetPublish = "article_asset_publish"
)

type RoleOracle func(ctx context.Context, emailAddress string) []Role

func NewRoleOracle(rootUser string) RoleOracle {
	return func(ctx context.Context, emailAddress string) []Role {
		ret := make([]Role, 0)
		if strings.ToLower(emailAddress) == strings.ToLower(rootUser) {
			ret = append(ret, RoleAssetPublish)
		}
		return ret
	}
}
