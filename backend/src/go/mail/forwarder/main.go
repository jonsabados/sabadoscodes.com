package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jonsabados/sabadoscodes.com/mail"
	"github.com/jonsabados/sabadoscodes.com/s3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"os"
)

func NewRequestHandler(logger zerolog.Logger, forwardEmail mail.Forwarder, subjectToSend, sendFrom, sendTo string) func(ctx context.Context, events events.SimpleEmailEvent) error {
	return func(ctx context.Context, events events.SimpleEmailEvent) error {
		ctx = logger.WithContext(ctx)
		for _, e := range events.Records {
			logger.Info().Str("id", e.SES.Mail.MessageID).Msg("processing item")
			fullSubject := fmt.Sprintf("%s (%s)", subjectToSend, e.SES.Mail.MessageID)
			err := forwardEmail(ctx, e.SES.Mail.MessageID, fullSubject, sendFrom, sendTo)
			if err != nil {
				logger.Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error sending email")
				return errors.WithStack(err)
			}
		}
		return nil
	}
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

	logLevel, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		panic(err)
	}
	logger := zerolog.New(os.Stdout).Level(logLevel)

	s3Client := s3.RawClient(sess)
	getObject := s3.NewObjectFetcher(s3Client)
	deleteObject := s3.NewObjectRemover(s3Client)

	sesClient := mail.NewRawClient(sess)
	mailSender := mail.NewSender(sesClient)

	forwarder := mail.NewForwarder(os.Getenv("MAIL_BUCKET"), getObject, mailSender, deleteObject)
	lambda.Start(NewRequestHandler(logger, forwarder, os.Getenv("SUBJECT_TO_SEND"), os.Getenv("MAIL_FROM"), os.Getenv("MAIL_TO")))
}
