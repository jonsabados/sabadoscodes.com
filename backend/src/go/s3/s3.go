package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"time"
)

type Object struct {
	Path string
	Size int64
}

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

type ObjectLister func(ctx context.Context, bucket string, filter string) ([]Object, error)

func NewObjectLister(client *s3.S3) ObjectLister {
	return func(ctx context.Context, bucket string, prefix string) ([]Object, error) {
		res, err := client.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			Prefix: aws.String(prefix),
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		ret := make([]Object, 0)
		for _, o := range res.Contents {
			ret = append(ret, Object{
				Path: *o.Key,
				Size: *o.Size,
			})
		}
		return ret, nil
	}
}

type ObjectSaver func(ctx context.Context, bucket string, objectKey string, object io.ReadSeeker, mimeType string) error

func NewObjectSaver(client *s3.S3) ObjectSaver {
	return func(ctx context.Context, bucket string, objectKey string, object io.ReadSeeker, mimeType string) error {
		zerolog.Ctx(ctx).Info().Str("bucket", bucket).Str("key", objectKey).Msg("saving object")
		_, err := client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(bucket),
			Key:          aws.String(objectKey),
			Body:         object,
			ContentType:  aws.String(mimeType),
			ACL:          aws.String("private"),
		})
		return err
	}
}

type PublicObjectSaver func(ctx context.Context, bucket string, objectKey string, object io.ReadSeeker, mimeType string, cacheDuration time.Duration) error

func NewPublicObjectSaver(client *s3.S3) PublicObjectSaver {
	return func(ctx context.Context, bucket string, objectKey string, object io.ReadSeeker, mimeType string, cacheDuration time.Duration) error {
		zerolog.Ctx(ctx).Info().Str("bucket", bucket).Str("key", objectKey).Msg("saving object")
		_, err := client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(bucket),
			Key:          aws.String(objectKey),
			Body:         object,
			ContentType:  aws.String(mimeType),
			CacheControl: aws.String(fmt.Sprintf("max-age=%d", int(cacheDuration.Seconds()))),
			ACL:          aws.String("public-read"),
		})
		return err
	}
}

func RawClient(sess *session.Session) *s3.S3 {
	ret := s3.New(sess)
	xray.AWS(ret.Client)
	return ret
}
