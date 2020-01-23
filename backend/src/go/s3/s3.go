package s3

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
)

type ObjectFetcher func(ctx context.Context, bucket, object string) (io.ReadCloser, error)

func NewObjectFetcher(client *s3.S3) ObjectFetcher {
	return func(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
		zerolog.Ctx(ctx).Debug().Str("bucket", bucket).Str("key", object).Msg("fetching object")
		res, err := client.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(object),
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return res.Body, nil
	}
}

type ObjectRemover func(ctx context.Context, bucket, object string) error

func NewObjectRemover(client *s3.S3) ObjectRemover {
	return func(ctx context.Context, bucket, object string) error {
		zerolog.Ctx(ctx).Info().Str("bucket", bucket).Str("key", object).Msg("removing object")
		_, err := client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(object),
		})
		return err
	}
}

func RawClient(sess *session.Session) *s3.S3 {
	return s3.New(sess)
}
