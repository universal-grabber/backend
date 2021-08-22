package common

import (
	"context"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

const loggerKey = "logger"
const meterKey = "meter"

type LoggerContext struct {
	context.Context
}

func WithLogger(parent context.Context) context.Context {
	if parent.Value(loggerKey) == nil {
		return context.WithValue(parent, loggerKey, log.WithContext(parent).WithField("trackId", rand.Intn(1000000)))
	}

	return parent
}

func WithMeter(parent context.Context, meter Meter) context.Context {
	if parent.Value(meterKey) == nil {
		return context.WithValue(parent, meterKey, meter)
	}

	return parent
}

func WithLoggerEntry(parent context.Context, entry *log.Entry) context.Context {
	return context.WithValue(parent, loggerKey, entry)
}

func WithLoggerField(parent context.Context, key string, value string) context.Context {
	loggerContext := WithLogger(parent)

	logger := loggerContext.Value(loggerKey).(*log.Entry)

	return WithLoggerEntry(parent, logger.WithField(key, value))
}

func WithOperation(parent context.Context, operation string) context.Context {
	return WithLoggerField(parent, "operation", operation)
}

func UseLogger(parent context.Context) *log.Entry {
	return parent.Value(loggerKey).(*log.Entry)
}

func UseMeter(parent context.Context) Meter {
	return parent.Value(meterKey).(Meter)
}
