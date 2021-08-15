package main

import (
	"backend/api/context"
	"backend/api/controller"
	"backend/api/grpc-impl"
	"backend/api/helper"
	"backend/api/service"
	"backend/common"
	pb "backend/gen/proto/service/api"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	_ "net/http/pprof"
)

func main() {
	common.EnableGrayLog("ugb-api")

	log.Info("Started\n")

	registerContext()

	gin.SetMode(gin.ReleaseMode)

	db := helper.UgbMongoInstance
	db.Init()

	r := gin.New()

	Routes(r)

	go runGrpc()

	go context.GetSchedulerService().Run()

	log.Info("Listening on :8080\n")
	err := r.Run(":8080")

	if err != nil {
		log.Error(err)
	}
}

func runGrpc() {
	pageRefGrpcService := new(grpc_impl.PageRefGrpcService)

	pageRefGrpcService.Init()

	lis, err := net.Listen("tcp", "0.0.0.0:6565")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterPageRefServiceServer(s, pageRefGrpcService)

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func registerContext() {
	// api instances
	websiteApi := new(controller.WebsiteApiImpl)
	pageRefApi := new(controller.PageRefApiImpl)
	scheduleApi := new(controller.ScheduleApiImpl)

	pageRefApi.Init()
	scheduleApi.Init()

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
