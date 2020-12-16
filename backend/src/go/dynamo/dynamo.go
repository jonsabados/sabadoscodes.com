package dynamo

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-xray-sdk-go/xray"
)

func RawClient(sess *session.Session) *dynamodb.DynamoDB {
	ret := dynamodb.New(sess)
	xray.AWS(ret.Client)
	return ret
}