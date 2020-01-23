package mail

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/jonsabados/sabadoscodes.com/s3"
	"github.com/jordan-wright/email"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
)

type Attachment struct {
	Name     string
	MimeType string
	Body     io.Reader
}

type Sender func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error

func NewSender(sesClient *ses.SES) Sender {
	return func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error {
		e := &email.Email{
			To:          []string{to},
			From:        from,
			Subject:     subject,
			Text:        []byte(textBody),
			HTML:        []byte(htmlBody),
		}
		for _, a := range attachments {
			_, err := e.Attach(a.Body, a.Name, a.MimeType)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		payload, err := e.Bytes()
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("source", from).Str("to", to).Msg("sending email")
		_, err = sesClient.SendRawEmail(&ses.SendRawEmailInput{
			Source: aws.String(from),
			Destinations: []*string{
				aws.String(to),
			},
			RawMessage: &ses.RawMessage{
				Data: payload,
			},
		})
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	}
}

type Forwarder func(ctx context.Context, messageID, subjectToSend, sendFrom, forwardTo string) error

func NewForwarder(mailBucket string, fetchObject s3.ObjectFetcher, sendEmail Sender, removeObject s3.ObjectRemover) Forwarder {
	return func(ctx context.Context, messageID, subjectToSend, sendFrom, forwardTo string) error {
		originalEmail, err := fetchObject(ctx, mailBucket, messageID)
		if err != nil {
			return errors.WithStack(err)
		}
		defer func() {
			err := originalEmail.Close()
			if err != nil {
				zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error closing stream")
			}
		}()

		htmlBody := "<p>See attached email</p>"
		textBody := "See attached email"
		err = sendEmail(ctx, sendFrom, forwardTo, subjectToSend, htmlBody, textBody, Attachment{
			Name:     fmt.Sprintf("%s.eml", messageID),
			MimeType: "message/rfc822",
			Body:     originalEmail,
		})
		if err != nil {
			return errors.WithStack(err)
		}

		err = removeObject(ctx, mailBucket, messageID)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	}
}

func NewRawClient(sess *session.Session) *ses.SES {
	return ses.New(sess)
}
