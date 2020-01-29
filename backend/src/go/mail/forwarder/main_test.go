package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_NewRequestHandler_BubblesErrors(t *testing.T) {
	asserter := assert.New(t)

	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	inputCtx := context.WithValue(context.Background(), "foo", "bar")
	inputSubject := "testing for fun and profit"
	inputSendFrom := "mcTester@foo.com"
	inputSendTo := "someone@testing.com"

	messageIDOne := "1234"
	messageIDTwo := "4321"

	input := events.SimpleEmailEvent{
		Records: []events.SimpleEmailRecord{
			{
				SES: events.SimpleEmailService{
					Mail: events.SimpleEmailMessage{
						MessageID: messageIDOne,
					},
					Receipt: events.SimpleEmailReceipt{},
				},
			},
			{
				SES: events.SimpleEmailService{
					Mail: events.SimpleEmailMessage{
						MessageID: messageIDTwo,
					},
					Receipt: events.SimpleEmailReceipt{},
				},
			},
		},
	}

	expectedError := "KaPow!"
	forwarder := func(ctx context.Context, messageID, subjectToSend, sendFrom, forwardTo string) error {
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(messageIDOne, messageID)
		asserter.Equal(fmt.Sprintf("%s (%s)", inputSubject, messageID), subjectToSend)
		asserter.Equal(inputSendFrom, sendFrom)
		asserter.Equal(inputSendTo, forwardTo)

		return errors.New(expectedError)
	}

	err := NewRequestHandler(logger, forwarder, inputSubject, inputSendFrom, inputSendTo)(inputCtx, input)
	asserter.EqualError(err, expectedError)
}

func Test_NewRequestHandler_HappyPath(t *testing.T) {
	asserter := assert.New(t)

	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	inputCtx := context.WithValue(context.Background(), "foo", "bar")
	inputSubject := "testing for fun and profit"
	inputSendFrom := "mcTester@foo.com"
	inputSendTo := "someone@testing.com"

	messageIDOne := "1234"
	messageIDTwo := "4321"

	input := events.SimpleEmailEvent{
		Records: []events.SimpleEmailRecord{
			{
				SES: events.SimpleEmailService{
					Mail: events.SimpleEmailMessage{
						MessageID: messageIDOne,
					},
					Receipt: events.SimpleEmailReceipt{},
				},
			},
			{
				SES: events.SimpleEmailService{
					Mail: events.SimpleEmailMessage{
						MessageID: messageIDTwo,
					},
					Receipt: events.SimpleEmailReceipt{},
				},
			},
		},
	}

	processedMessages := make([]string, 0)
	forwarder := func(ctx context.Context, messageID, subjectToSend, sendFrom, forwardTo string) error {
		asserter.Equal(inputCtx, ctx)
		asserter.Equal(fmt.Sprintf("%s (%s)", inputSubject, messageID), subjectToSend)
		asserter.Equal(inputSendFrom, sendFrom)
		asserter.Equal(inputSendTo, forwardTo)

		processedMessages = append(processedMessages, messageID)

		return nil
	}

	err := NewRequestHandler(logger, forwarder, inputSubject, inputSendFrom, inputSendTo)(inputCtx, input)
	asserter.NoError(err)
	asserter.Equal([]string{messageIDOne, messageIDTwo}, processedMessages)
}
