package main

import (
	appPackage "backend/model-parser/app"
)

func main() {
	app := new(appPackage.App)

	app.Addr = ":8443"

	app.CertFile = "server.crt"
	app.KeyFile = "server.key"

	app.Run()
}