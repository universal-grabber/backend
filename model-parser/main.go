package main

import (
	"backend/common"
	appPackage "backend/model-parser/app"
)

func main() {
	app := new(appPackage.App)

	common.EnableGrayLog("model-processor")

	app.Addr = ":8443"

	app.CertFile = "server.crt"
	app.KeyFile = "server.key"

	app.Run()
}
