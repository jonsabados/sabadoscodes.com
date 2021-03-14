package article

import (
	"context"
	"fmt"
	"strconv"
	"sync"
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
	fieldPublished   = "Published"
)

type Article struct {
	Summary
	Content string `json:"content"`
}

type Summary struct {
	Slug        string     `json:"slug"`
	PublishDate *time.Time `json:"publishDate,omitempty"`
	Title       string     `json:"title"`
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
			item[fieldPublished] = &dynamodb.AttributeValue{S: aws.String("true")}
		} else {
			item[fieldPublished] = &dynamodb.AttributeValue{S: aws.String("false")}
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
		res, err := db.GetItemWithContext(ctx, &dynamodb.GetItemInput{
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
		publishDate, err := publishedDate(res.Item)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		ret := &Article{
			Summary: Summary{
				Slug:        *res.Item[fieldSlug].S,
				Title:       *res.Item[fieldTitle].S,
				PublishDate: publishDate,
			},
			Content: *res.Item[fieldContent].S,
		}
		return ret, nil
	}
}

func NewCachedFetcher(base Fetcher, cacheDuration time.Duration) Fetcher {
	mutex := sync.Mutex{}
	cache := make(map[string]struct {
		cachedTime time.Time
		article    *Article
	})
	return func(ctx context.Context, slug string) (*Article, error) {
		mutex.Lock()
		defer mutex.Unlock()
		if rec, inCache := cache[slug]; !inCache || time.Now().After(rec.cachedTime.Add(cacheDuration)) {
			fresh, err := base(ctx, slug)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			rec.cachedTime = time.Now()
			rec.article = fresh
			cache[slug] = rec
		}
		return cache[slug].article, nil
	}
}

type Lister func(ctx context.Context, published bool) ([]Summary, error)

func NewLister(db *dynamodb.DynamoDB, articleTable string) Lister {
	return func(ctx context.Context, published bool) ([]Summary, error) {
		res, err := db.ScanWithContext(ctx, &dynamodb.ScanInput{
			TableName:        aws.String(articleTable),
			FilterExpression: aws.String(fmt.Sprintf("%s = :published", fieldPublished)),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":published": {S: aws.String(strconv.FormatBool(published))},
			},
			ProjectionExpression: aws.String(fmt.Sprintf("%s, %s, %s", fieldSlug, fieldPublishDate, fieldTitle)),
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		ret := make([]Summary, *res.Count)
		for i := int64(0); i < *res.Count; i++ {
			rec := res.Items[i]
			publishDate, err := publishedDate(rec)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			ret[i] = Summary{
				Slug:        *rec[fieldSlug].S,
				Title:       *rec[fieldTitle].S,
				PublishDate: publishDate,
			}
		}
		return ret, nil
	}
}

func NewCachedLister(base Lister, cacheDuration time.Duration) Lister {
	var mutex sync.Mutex
	cache := make(map[bool]struct {
		cachedTime time.Time
		articles   []Summary
	})
	return func(ctx context.Context, published bool) ([]Summary, error) {
		mutex.Lock()
		defer mutex.Unlock()
		if rec, inCache := cache[published]; !inCache || time.Now().After(rec.cachedTime.Add(cacheDuration))  {
			fresh, err := base(ctx, published)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			rec.articles = fresh
			rec.cachedTime = time.Now()
			cache[published] = rec
		}
		return cache[published].articles, nil
	}
}

func publishedDate(item map[string]*dynamodb.AttributeValue) (*time.Time, error) {
	if item[fieldPublishDate] == nil {
		return nil, nil
	}
	unixTime, err := strconv.ParseInt(*item[fieldPublishDate].N, 10, 64)
	if err != nil {
		return nil, errors.Errorf("invalid timestamp %s for article %s", *item[fieldPublishDate].N, *item[fieldSlug].S)
	}
	publishDate := time.Unix(unixTime, 0)
	return &publishDate, nil
}
