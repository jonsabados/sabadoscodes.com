package main

import (
	"archive/zip"
	"context"
	"fmt"
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

	"github.com/jonsabados/sabadoscodes.com/logging"
	"github.com/jonsabados/sabadoscodes.com/s3"
)

func newHandler(prepLogs logging.Preparer,
	backupBucket string,
	assetBucket string,
	listObjects s3.ObjectLister,
	fetchObject s3.ObjectFetcher,
	saveObject s3.ObjectSaver) func(ctx context.Context) error {

	return func(ctx context.Context) error {
		ctx, logger := prepLogs(ctx)
		assetBucketContents, err := listObjects(ctx, assetBucket, "")
		if err != nil {
			logger.Error().Str("err", fmt.Sprintf("%+v", err)).Msg("error listing objects in asset bucket")
			return err
		}
		sort.Slice(assetBucketContents, func(i, j int) bool {
			return strings.Compare(assetBucketContents[i].Path, assetBucketContents[j].Path) <= 0
		})

		tempFile, err := ioutil.TempFile(os.TempDir(), "backup*.zip")
		if err != nil {
			logger.Error().Str("err", fmt.Sprintf("%+v", err)).Msg("error creating temp file")
			return err
		}
		zipOut := zip.NewWriter(tempFile)
		for _, o := range assetBucketContents {
			err := addToZip(ctx, fetchObject, assetBucket, o, logger, zipOut)
			if err != nil {
				logger.Error().Str("err", fmt.Sprintf("%+v", err)).Msg("error adding asset to zip")
				return err
			}
		}
		err = zipOut.Close()
		if err != nil {
			logger.Error().Str("err", fmt.Sprintf("%+v", err)).Msg("error adding asset to zip")
			return err
		}
		err = tempFile.Close()
		if err != nil {
			logger.Error().Str("err", fmt.Sprintf("%+v", err)).Msg("error closing temp file")
			return err
		}

		zipIn, err := os.Open(tempFile.Name())
		if err != nil {
			logger.Error().Str("err", fmt.Sprintf("%+v", err)).Msg("error opening created file")
			return err
		}
		err = saveObject(ctx, backupBucket, "nightlyBack.zip", zipIn, "application/zip")
		if err != nil {
			logger.Error().Str("err", fmt.Sprintf("%+v", err)).Msg("error writing backup")
			return err
		}
		zipIn.Close()

		return nil
	}
}

func addToZip(ctx context.Context, fetchObject s3.ObjectFetcher, assetBucket string, o s3.Object, logger zerolog.Logger, zipOut *zip.Writer) error {
	zerolog.Ctx(ctx).Info().Interface("object", o).Msg("Adding object")

	contents, err := fetchObject(ctx, assetBucket, o.Path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer contents.Close()

	dest, err := zipOut.Create(o.Path)
	if err != nil {
		logger.Error().Str("err", fmt.Sprintf("%+v", err)).Interface("asset", o).Str("bucket", assetBucket).Msg("error creating zip node")
		return errors.WithStack(err)
	}

	_, err = io.Copy(dest, contents)
	if err != nil {
		logger.Error().Str("err", fmt.Sprintf("%+v", err)).Interface("asset", o).Str("bucket", assetBucket).Msg("error copying asset")
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

	s3Client := s3.RawClient(sess)
	lister := s3.NewObjectLister(s3Client)
	fetcher := s3.NewObjectFetcher(s3Client)
	saver := s3.NewObjectSaver(s3Client)

	handler := newHandler(logging.NewPreparer(), targetBucket, assetBucket, lister, fetcher, saver)

	lambda.Start(handler)
}
