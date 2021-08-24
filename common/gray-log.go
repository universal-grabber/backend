package common

import (
	log "github.com/sirupsen/logrus"
	graylog "github.com/tislib/logrus-graylog-hook/v3"
	"time"
)

var hook *graylog.GraylogHook

func EnableGrayLog(service string) {
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)

	hook = graylog.NewAsyncGraylogHook("ug.tisserv.net:12201", map[string]interface{}{"service": service})
	hook.Level = log.DebugLevel
	hook.Writer().CompressionType = 0
	hook.Writer().CompressionLevel = 9

	log.AddHook(hook)
	log.SetFormatter(new(NullFormatter)) // Don't send logs to stdout
}

func EnableTraceLogging(duration time.Duration) {
	go func() {
		log.Info("enabling trace logging")
		log.SetLevel(log.TraceLevel)
		hook.Level = log.TraceLevel

		time.Sleep(duration)

		log.Info("disabling trace logging")
		log.SetLevel(log.DebugLevel)
		hook.Level = log.DebugLevel
	}()
}
