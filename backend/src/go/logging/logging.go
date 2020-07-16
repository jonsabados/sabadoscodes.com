package logging

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog"
	"os"
)

// Preparer sets up a context for logging, returning a context that has a logger established as well as the set logger
type Preparer func(ctx context.Context) (context.Context, zerolog.Logger)

func NewPreparer() Preparer {
	logLevel, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		tmpLogger := zerolog.New(os.Stdout)
		tmpLogger.Fatal().Err(err).Msg("unable to configure logger, set LOG_LEVEL to an appropriate value")
	}
	baseLogger := zerolog.New(os.Stdout).Level(logLevel)

	return func(ctx context.Context) (context.Context, zerolog.Logger) {
		if awsCtx, inLambda := lambdacontext.FromContext(ctx); inLambda {
			logger := baseLogger.With().Str("requestId", awsCtx.AwsRequestID).Logger()
			return logger.WithContext(ctx), logger
		} else {
			return ctx, baseLogger
		}
	}
}