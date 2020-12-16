package article

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

type Article struct {
	Slug        string
	PublishDate time.Time
	Content     string
}

type Saver func(ctx context.Context, article Article) error

func NewSaver(db *dynamodb.DynamoDB, articleTable string) Saver {
	return func(ctx context.Context, article Article) error {
		toPut := &dynamodb.PutItemInput{
			TableName: aws.String(articleTable),
			Item: map[string]*dynamodb.AttributeValue{
				"Slug":        {S: aws.String(article.Slug)},
				"PublishDate": {N: aws.String(strconv.FormatInt(article.PublishDate.Unix(), 10))},
				"Content":     {S: aws.String(article.Content)},
			},
		}

		_, err := db.PutItemWithContext(ctx, toPut)
		return errors.WithStack(err)
	}
}
