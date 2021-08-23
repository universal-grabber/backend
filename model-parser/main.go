package main

import (
	"backend/common"
	appPackage "backend/model-parser/app"
)

func main() {
	common.Init("model-parser")
	app := new(appPackage.App)

	app.Addr = ":8443"

	app.CertFile = "server.crt"
	app.KeyFile = "server.key"

	app.Run()
}
