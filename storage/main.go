package main

import (
	appPackage "backend/storage/app"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)

	app := new(appPackage.App)

	app.Addr = ":8080"

	app.CertFile = "server.crt"
	app.KeyFile = "server.key"

	app.Run()
}
