package mail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"testing"
)

func Test_NewForwarder_ErrorReadingFromBucket_BubblesError(t *testing.T) {
	asserter := assert.New(t)

	bucketToUse := "somebucket"

	inputCtx := context.WithValue(context.Background(), "foo", "bar")
	inputMessageID := "12345"
	inputSubject := "testing FTW"
	inputSendFrom := "bob@testing.com"
	inputForwardTo := "mcTester@testing.com"

	expectedError := "KaBOOm!"
	objectFetcher := func(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(bucketToUse, bucket)
		asserter.Equal(inputMessageID, object)
		return nil, errors.New(expectedError)
	}

	sender := func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error {
		asserter.Fail("no mail should have been sent")
		return nil
	}

	objectRemover := func(ctx context.Context, bucket, object string) error {
		asserter.Fail("object should have been left in bucket")
		return nil
	}

	err := NewForwarder(bucketToUse, objectFetcher, sender, objectRemover)(inputCtx, inputMessageID, inputSubject, inputSendFrom, inputForwardTo)

	asserter.EqualError(err, expectedError)
}

func Test_NewForwarder_ErrorSending_BubblesError(t *testing.T) {
	asserter := assert.New(t)

	bucketToUse := "somebucket"

	inputCtx := context.WithValue(context.Background(), "foo", "bar")
	inputMessageID := "12345"
	inputSubject := "testing FTW"
	inputSendFrom := "bob@testing.com"
	inputForwardTo := "mcTester@testing.com"

	expectedAttachment := ioutil.NopCloser(bytes.NewReader([]byte("this would be an email")))
	objectFetcher := func(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(bucketToUse, bucket)
		asserter.Equal(inputMessageID, object)
		return expectedAttachment, nil
	}

	expectedError := "KaBOOm!"
	mailSent := false
	sender := func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error {
		mailSent = true
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(inputSendFrom, from)
		asserter.Equal(inputForwardTo, to)
		asserter.Equal(inputSubject, subject)
		asserter.Equal("<p>See attached email</p>", htmlBody)
		asserter.Equal("See attached email", textBody)
		asserter.Equal([]Attachment{
			{
				Name:     fmt.Sprintf("%s.eml", inputMessageID),
				MimeType: "message/rfc822",
				Body:     expectedAttachment,
			},
		}, attachments)
		return errors.New(expectedError)
	}

	objectRemover := func(ctx context.Context, bucket, object string) error {
		asserter.Fail("object should have been left in bucket")
		return nil
	}

	err := NewForwarder(bucketToUse, objectFetcher, sender, objectRemover)(inputCtx, inputMessageID, inputSubject, inputSendFrom, inputForwardTo)

	asserter.True(mailSent)
	asserter.EqualError(err, expectedError)
}

func Test_NewForwarder_ErrorCleaningBucket_BubblesError(t *testing.T) {
	asserter := assert.New(t)

	bucketToUse := "somebucket"

	inputCtx := context.WithValue(context.Background(), "foo", "bar")
	inputMessageID := "12345"
	inputSubject := "testing FTW"
	inputSendFrom := "bob@testing.com"
	inputForwardTo := "mcTester@testing.com"

	expectedAttachment := ioutil.NopCloser(bytes.NewReader([]byte("this would be an email")))
	objectFetcher := func(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(bucketToUse, bucket)
		asserter.Equal(inputMessageID, object)
		return expectedAttachment, nil
	}

	mailSent := false
	sender := func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error {
		mailSent = true
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(inputSendFrom, from)
		asserter.Equal(inputForwardTo, to)
		asserter.Equal(inputSubject, subject)
		asserter.Equal("<p>See attached email</p>", htmlBody)
		asserter.Equal("See attached email", textBody)
		asserter.Equal([]Attachment{
			{
				Name:     fmt.Sprintf("%s.eml", inputMessageID),
				MimeType: "message/rfc822",
				Body:     expectedAttachment,
			},
		}, attachments)
		return nil
	}

	expectedError := "KaBOOm!"
	objectRemoved := false
	objectRemover := func(ctx context.Context, bucket, object string) error {
		objectRemoved = true
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(bucketToUse, bucket)
		asserter.Equal(inputMessageID, object)
		return errors.New(expectedError)
	}

	err := NewForwarder(bucketToUse, objectFetcher, sender, objectRemover)(inputCtx, inputMessageID, inputSubject, inputSendFrom, inputForwardTo)

	asserter.True(mailSent)
	asserter.True(objectRemoved)
	asserter.EqualError(err, expectedError)
}

func Test_NewForwarder_HappyPath(t *testing.T) {
	asserter := assert.New(t)

	bucketToUse := "somebucket"

	inputCtx := context.WithValue(context.Background(), "foo", "bar")
	inputMessageID := "12345"
	inputSubject := "testing FTW"
	inputSendFrom := "bob@testing.com"
	inputForwardTo := "mcTester@testing.com"

	expectedAttachment := ioutil.NopCloser(bytes.NewReader([]byte("this would be an email")))
	objectFetcher := func(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(bucketToUse, bucket)
		asserter.Equal(inputMessageID, object)
		return expectedAttachment, nil
	}

	mailSent := false
	sender := func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error {
		mailSent = true
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(inputSendFrom, from)
		asserter.Equal(inputForwardTo, to)
		asserter.Equal(inputSubject, subject)
		asserter.Equal("<p>See attached email</p>", htmlBody)
		asserter.Equal("See attached email", textBody)
		asserter.Equal([]Attachment{
			{
				Name:     fmt.Sprintf("%s.eml", inputMessageID),
				MimeType: "message/rfc822",
				Body:     expectedAttachment,
			},
		}, attachments)
		return nil
	}

	objectRemoved := false
	objectRemover := func(ctx context.Context, bucket, object string) error {
		objectRemoved = true
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(bucketToUse, bucket)
		asserter.Equal(inputMessageID, object)
		return nil
	}

	err := NewForwarder(bucketToUse, objectFetcher, sender, objectRemover)(inputCtx, inputMessageID, inputSubject, inputSendFrom, inputForwardTo)

	asserter.True(mailSent)
	asserter.True(objectRemoved)
	asserter.NoError(err)
}
