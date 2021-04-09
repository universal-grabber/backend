package main

import (
	"backend/api/context"
	"backend/api/controller"
	"backend/api/helper"
	"backend/api/service"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	_ "net/http/pprof"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)

	fmt.Print("Started\n")

	registerContext()

	gin.SetMode(gin.ReleaseMode)

	db := helper.UgbMongoInstance
	db.Init()

	r := gin.New()

	Routes(r)

	go context.GetSchedulerService().Run()

	fmt.Print("Listening on :8080\n")
	err := r.Run(":8080")

	if err != nil {
		log.Print(err)
	}
}

func registerContext() {
	// api instances
	websiteApi := new(controller.WebsiteApiImpl)
	pageRefApi := new(controller.PageRefApiImpl)
	scheduleApi := new(controller.ScheduleApiImpl)

	context.Register(context.WebsiteApiInstance, websiteApi)
	context.Register(context.PageRefApiInstance, pageRefApi)
	context.Register(context.ScheduleApiInstance, scheduleApi)

	// service instances
	pageRefService := new(service.PageRefService)
	schedulerService := new(service.SchedulerServiceImpl)
	tagsService := new(service.TagsServiceImpl)

	context.Register(context.PageRefServiceInstance, pageRefService)
	context.Register(context.SchedulerInstance, schedulerService)
	context.Register(context.TagsServiceInstance, tagsService)
}
