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

	log.WithField("requestId", requestId).
		WithField("operation", "schedule-kafka").
		Info("starting to schedule: ", searchPageRef)

	if err != nil {
		log.WithField("requestId", requestId).
			WithField("operation", "schedule-kafka").
			Error(err)
		return
	}

	maxSize := searchPageRef.PageSize
	count := 0

	searchPageRef.PageSize = 10000
	searchPageRef.Page = 0

	go func() {
		interruptChan := make(chan bool)
		// search async

		for {
			log.WithField("requestId", requestId).
				WithField("page", searchPageRef.Page).
				WithField("operation", "schedule-kafka").
				Info("starting fetch page")
			pageChan := make(chan *model.PageRef, 100)
			go func() {
				log.WithField("requestId", requestId).
					WithField("page", searchPageRef.Page).
					WithField("operation", "schedule-kafka").
					Debug("request search")

				receiver.service.Search(searchPageRef, pageChan, interruptChan)

				log.WithField("requestId", requestId).
					WithField("page", searchPageRef.Page).
					WithField("operation", "schedule-kafka").
					Debug("end request search")
			}()

			localCount := 0
			for pageRef := range pageChan {
				localCount++

				err := kafka.SendPageRef(pageRef)

				timeCalc.Step()

				if err != nil {
					log.Error(err)
					break
				}
			}
			count += localCount

			log.WithField("requestId", requestId).
				WithField("page", searchPageRef.Page).
				WithField("operation", "schedule-kafka").
				Debug("localCount: %d; totalCount: %d", localCount, count)

			searchPageRef.Page++
			if maxSize <= count {
				log.WithField("requestId", requestId).
					WithField("page", searchPageRef.Page).
					WithField("operation", "schedule-kafka").
					Debug("interrupting as count reached max size %d / %d", localCount, count)

				interruptChan <- true
				break
			}
			if localCount == 0 {
				log.WithField("requestId", requestId).
					WithField("page", searchPageRef.Page).
					WithField("operation", "schedule-kafka").
					Debug("interrupting as data end reached", localCount, count)

				interruptChan <- true
				break
			}
		}
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
