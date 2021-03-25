package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/jonsabados/sabadoscodes.com/article"
	"github.com/jonsabados/sabadoscodes.com/dynamo"
	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/s3"
)

func newHandler(prepLogs logging.Preparer,
	backupBucket string,
	assetBucket string,
	listObjects s3.ObjectLister,
	fetchObject s3.ObjectFetcher,
	saveObject s3.ObjectSaver,
	listArticles article.Lister,
	fetchArticle article.Fetcher) func(ctx context.Context) error {

	return func(ctx context.Context) error {
		ctx, logger := prepLogs(ctx)

		tempFile, err := ioutil.TempFile(os.TempDir(), "backup*.zip")
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error creating temp file")
			return err
		}
		zipOut := zip.NewWriter(tempFile)

		err = addAssets(ctx, listObjects, assetBucket, fetchObject, zipOut)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error adding asset to zip")
			return err
		}

		err = addArticles(ctx, listArticles, fetchArticle, zipOut)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error adding articles to zip")
			return err
		}

		err = zipOut.Close()
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error closing zip stream")
			return err
		}
		err = tempFile.Close()
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error closing temp file")
			return err
		}

		zipIn, err := os.Open(tempFile.Name())
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error opening created file")
			return err
		}
		err = saveObject(ctx, backupBucket, "nightlyBack.zip", zipIn, "application/zip")
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error writing backup")
			return err
		}
		zipIn.Close()

		return nil
	}
}

func addArticles(ctx context.Context, listArticles article.Lister, fetchArticle article.Fetcher, zipOut *zip.Writer) error {
	logger := zerolog.Ctx(ctx)
	articles := make([]*article.Article, 0)
	published, err := listArticles(ctx, true)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error listing published articles")
		return err
	}

	unpublished, err := listArticles(ctx, false)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error listing unpublished articles")
		return err
	}

	for _, summary := range published {
		a, err := fetchArticle(ctx, summary.Slug)
		if err != nil {
			logger.Error().Stack().Err(err).Str("slug", a.Slug).Msg("error fetching published article")
			return err
		}
		articles = append(articles, a)
	}

	for _, summary := range unpublished {
		a, err := fetchArticle(ctx, summary.Slug)
		if err != nil {
			logger.Error().Stack().Err(err).Str("slug", a.Slug).Msg("error fetching unpublished article")
			return err
		}
		articles = append(articles, a)
	}

	bytes, err := json.Marshal(articles)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error marshalling articles")
		return err
	}

	dest, err := zipOut.Create("articles.json")
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error creating zip node")
		return errors.WithStack(err)
	}

	_, err = dest.Write(bytes)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error writing articles to zip")
		return err
	}

	return nil
}

func addAssets(ctx context.Context, listObjects s3.ObjectLister, assetBucket string, fetchObject s3.ObjectFetcher, zipOut *zip.Writer) error {
	logger := zerolog.Ctx(ctx)
	assetBucketContents, err := listObjects(ctx, assetBucket, "")
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error listing objects in asset bucket")
		return err
	}
	sort.Slice(assetBucketContents, func(i, j int) bool {
		return strings.Compare(assetBucketContents[i].Path, assetBucketContents[j].Path) <= 0
	})

	for _, o := range assetBucketContents {
		err := addToZip(ctx, fetchObject, assetBucket, o, zipOut)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("error adding asset to zip")
			return err
		}
	}
	return nil
}

func addToZip(ctx context.Context, fetchObject s3.ObjectFetcher, assetBucket string, o s3.Object, zipOut *zip.Writer) error {
	zerolog.Ctx(ctx).Info().Interface("object", o).Msg("Adding object")

	contents, err := fetchObject(ctx, assetBucket, o.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer contents.Close()

	dest, err := zipOut.Create(o.Path)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Interface("asset", o).Str("bucket", assetBucket).Msg("error creating zip node")
		return errors.WithStack(err)
	}

	_, err = io.Copy(dest, contents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Interface("asset", o).Str("bucket", assetBucket).Msg("error copying asset")
		return err
	}
	return nil
}

func main() {
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		panic(err)
	}

	assetBucket := os.Getenv("ASSET_BUCKET")
	targetBucket := os.Getenv("TARGET_BUCKET")
	articleTable := os.Getenv("ARTICLE_TABLE")

	s3Client := s3.RawClient(sess)
	dynamoClient := dynamo.RawClient(sess)
	lister := s3.NewObjectLister(s3Client)
	fetcher := s3.NewObjectFetcher(s3Client)
	saver := s3.NewObjectSaver(s3Client)
	listArticle := article.NewLister(dynamoClient, articleTable)
	fetchArticle := article.NewFetcher(dynamoClient, articleTable)

	handler := newHandler(logging.NewPreparer(), targetBucket, assetBucket, lister, fetcher, saver, listArticle, fetchArticle)

	lambda.Start(handler)
}
