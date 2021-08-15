package common

import (
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	log "github.com/sirupsen/logrus"
)

func EnableGrayLog(service string) {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)

	hook := graylog.NewGraylogHook("ug.tisserv.net:12201", map[string]interface{}{"service": service})
	hook.Level = log.TraceLevel
	log.AddHook(hook)
	log.SetFormatter(new(NullFormatter)) // Don't send logs to stdout
}