package context_util

import (
	"context"
	log "github.com/sirupsen/logrus"
)

const loggerKey = "logger"

type LoggerContext struct {
	context.Context
}

func WithLogger(parent context.Context) context.Context {
	if parent.Value(loggerKey) == nil {
		return context.WithValue(parent, loggerKey, log.WithContext(parent))
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
