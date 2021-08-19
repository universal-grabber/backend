package main

import (
	"backend/common"
	log "github.com/sirupsen/logrus"
)

func main() {
	common.EnableGrayLog("test")

	log.Print("print1")
	log.Trace("trace1")
	log.Debug("debug1")
	log.Info("info1")
	log.Warn("warning1")
	log.Error("error1")
	log.Fatal("fatal1")
	log.Panic("panic1")
}
