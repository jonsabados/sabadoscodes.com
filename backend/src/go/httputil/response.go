package httputil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
)

type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type BadRequestDTO struct {
	TrackingID  string       `json:"trackingId"`
	Errors      []string     `json:"errors,omitempty"`
	FieldErrors []FieldError `json:"fieldErrors,omitempty"`
}

type ErrorTracker struct {
	errors      []string
	fieldErrors []FieldError
}

func (e ErrorTracker) WithError(error string) ErrorTracker {
	if e.errors == nil {
		e.errors = make([]string, 0)
	}
	e.errors = append(e.errors, error)
	return e
}

func (e ErrorTracker) WithFieldError(field string, error string) ErrorTracker {
	if e.fieldErrors == nil {
		e.fieldErrors = make([]FieldError, 0)
	}
	e.fieldErrors = append(e.fieldErrors, FieldError{field, error})
	return e
}

func (e ErrorTracker) InError() bool {
	return len(e.errors) > 0 || len(e.fieldErrors) > 0
}

func (e ErrorTracker) ToAPIResponse(ctx context.Context, headers map[string]string) events.APIGatewayProxyResponse {
	body := BadRequestDTO{
		Errors:      e.errors,
		FieldErrors: e.fieldErrors,
	}
	if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
		body.TrackingID = awsCtx.AwsRequestID
	}
	responseBody, err := json.Marshal(body)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("unable to marshall response body")
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusBadRequest,
		Body:       string(responseBody),
		Headers:    headers,
	}
}
