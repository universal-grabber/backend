package model

import log "github.com/sirupsen/logrus"

type Config struct {
	LogLevel        log.Level
	EnabledWebsites []string
	EnabledTasks    []string

	UgbApiUri            string
	UgbModelProcessorUri string
	UgbStorageUri        string
	ParseMongoUri        string
}
