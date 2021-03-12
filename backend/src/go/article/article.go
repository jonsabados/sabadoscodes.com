package article

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

const (
	articleSortKey   = "Article"
	fieldSlug        = "Slug"
	fieldSortKey     = "SortKey"
	fieldTitle       = "Title"
	fieldContent     = "Content"
	fieldPublishDate = "PublishDate"
)

type Article struct {
	Slug        string     `json:"slug"`
	PublishDate *time.Time `json:"publishDate,omitempty"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
}

type Saver func(ctx context.Context, article Article) error

func NewSaver(db *dynamodb.DynamoDB, articleTable string) Saver {
	return func(ctx context.Context, article Article) error {
		item := map[string]*dynamodb.AttributeValue{
			fieldSlug:    {S: aws.String(article.Slug)},
			fieldSortKey: {S: aws.String(articleSortKey)},
			fieldTitle:   {S: aws.String(article.Title)},
			fieldContent: {S: aws.String(article.Content)},
		}

		if article.PublishDate != nil {
			item[fieldPublishDate] = &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(article.PublishDate.Unix(), 10))}
		}

		toPut := &dynamodb.PutItemInput{
			TableName: aws.String(articleTable),
			Item:      item,
		}

		_, err := db.PutItemWithContext(ctx, toPut)
		return errors.WithStack(err)
	}
}

type Fetcher func(ctx context.Context, slug string) (*Article, error)

func NewFetcher(db *dynamodb.DynamoDB, articleTable string) Fetcher {
	return func(ctx context.Context, slug string) (*Article, error) {
		res, err := db.GetItem(&dynamodb.GetItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				fieldSlug:    {S: aws.String(slug)},
				fieldSortKey: {S: aws.String(articleSortKey)},
			},
			TableName: aws.String(articleTable),
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if res.Item == nil {
			return nil, nil
		}
		ret := &Article{
			Slug:    *res.Item[fieldSlug].S,
			Title:   *res.Item[fieldTitle].S,
			Content: *res.Item[fieldContent].S,
		}
		if res.Item[fieldPublishDate] != nil {
			unixTime, err := strconv.ParseInt(*res.Item[fieldPublishDate].N, 10, 64)
			if err != nil {
				return nil, errors.Errorf("invalid timestamp %s for article %s", *res.Item[fieldPublishDate].N, slug)
			}
			publishDate := time.Unix(unixTime, 0)
			ret.PublishDate = &publishDate
		}
		return ret, nil
	}
}
