package controller

import (
	context2 "backend/api/context"
	"backend/api/helper"
	"backend/api/model"
	"backend/api/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

type ScheduleApiImpl struct {
	service *service.PageRefService
}

func (receiver *ScheduleApiImpl) Init() {
	receiver.service = new(service.PageRefService)
}

func (receiver *ScheduleApiImpl) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/1.0/schedule/apply-tags", receiver.applyTags)
	r.GET("/api/1.0/schedule/website", receiver.manualScheduleWebsite)
	r.GET("/api/1.0/schedule/reload", receiver.reload)

	r.GET("/api/1.0/schedule/kafka", receiver.ScheduleKafka)
}

func (receiver *ScheduleApiImpl) ScheduleKafka(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("ScheduleKafka")

	kafka := helper.UgbKafkaInstance

	searchPageRef := new(model.SearchPageRef)
	err := helper.ParseRequestQuery(c.Request, searchPageRef)

	requestId := rand.Intn(1000000)
	opLog := log.WithField("requestId", requestId).
		WithField("operation", "schedule-kafka")

	opLog.Info("starting to schedule: ", searchPageRef.State, searchPageRef.Status, searchPageRef.Tags)

	timeCalc.Logger(opLog)

	if err != nil {
		opLog.Error(err)
		return
	}

	count := 0

	go func() {
		interruptChan := make(chan bool)
		// search async

		pageLog := opLog.WithField("page", searchPageRef.Page)

		pageLog.Info("starting fetch page")
		pageChan := make(chan *model.PageRef, 100)
		go func() {
			pageLog.Debug("request search")

			receiver.service.Search(searchPageRef, pageChan, interruptChan)

			pageLog.Debug("end request search")
		}()

		var buffer []*model.PageRef

		for pageRef := range pageChan {
			count++

			buffer = append(buffer, pageRef)

			if len(buffer) >= 100000 {
				log.Info("starting flush %d out of %d", buffer, count)
				err := kafka.SendPageRef(buffer)
				if err != nil {
					pageLog.Error(err)
					break
				}
				log.Info("end flush %d out of %d", buffer, count)

				buffer = nil
			}

			timeCalc.Step()
		}

		if buffer != nil && len(buffer) > 0 {
			log.Info("starting flush %d out of %d (tail)", buffer, count)

			err := kafka.SendPageRef(buffer)
			if err != nil {
				pageLog.Error(err)
			}

			log.Info("end flush %d out of %d (tail)", buffer, count)
		}

		pageLog.Debugf("message sent kafka count: %d", count)
	}()

}

func (receiver *ScheduleApiImpl) applyTags(context *gin.Context) {
	websiteName, ok := context.GetQuery("websiteName")

	if !ok {
		context.String(400, "websiteName is required")
		return
	}

	context2.GetTagsService().ApplyTagsForWebsite(websiteName)

	context.String(200, "done")
}

func (receiver *ScheduleApiImpl) manualScheduleWebsite(context *gin.Context) {
	websiteName, ok := context.GetQuery("websiteName")

	if !ok {
		context.String(400, "websiteName is required")
		return
	}

	context2.GetSchedulerService().ScheduleWebsiteManual(websiteName)

	context.String(200, "done")
}

func (receiver *ScheduleApiImpl) reload(context *gin.Context) {
	context2.GetSchedulerService().ReloadWebsites()

	context.String(200, "done")
}
