package model

import log "github.com/sirupsen/logrus"

type Config struct {
	LogLevel        log.Level
	EnabledWebsites []string
	EnabledTasks    []string

	UgbApiUri            string
	UgbApiGrpcUri            string
	UgbModelProcessorUri string
	UgbStorageUri        string
	ParseMongoUri        string
}
