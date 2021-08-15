package main

import (
	"backend/common"
	appPackage "backend/storage/app"
)

func main() {
	common.EnableGrayLog("ugb-storage")

	app := new(appPackage.App)

	app.Addr = ":8443"

	app.CertFile = "server.crt"
	app.KeyFile = "server.key"

	app.Run()
}
