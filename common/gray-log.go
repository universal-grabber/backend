package common

import (
	log "github.com/sirupsen/logrus"
	graylog "github.com/tislib/logrus-graylog-hook/v3"
)

func EnableGrayLog(service string) {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)

	log.TraceLevel.String()

	hook := graylog.NewAsyncGraylogHook("ug.tisserv.net:12201", map[string]interface{}{"service": service})
	hook.Level = log.TraceLevel
	hook.Writer().CompressionType = 0
	hook.Writer().CompressionLevel = 9

	log.AddHook(hook)
	//log.SetFormatter(new(NullFormatter)) // Don't send logs to stdout
}
