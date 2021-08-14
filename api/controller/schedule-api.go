package controller

import (
	context2 "backend/api/context"
	"backend/api/helper"
	"backend/api/model"
	"backend/api/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
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

	r.GET("/api/1.0/page-refs/schedule-kafka", receiver.ScheduleKafka)
	r.GET("/api/1.0/page-refs/read-kafka", receiver.ReadKafka)
}

func (receiver *ScheduleApiImpl) ScheduleKafka(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("ScheduleKafka")

	kafka := helper.UgbKafkaInstance

	searchPageRef := new(model.SearchPageRef)
	err := helper.ParseRequestQuery(c.Request, searchPageRef)

	if err != nil {
		panic(err)
	}

	page := searchPageRef.Page

	go func() {
		interruptChan := make(chan bool)
		// search async
		for {
			pageChan := make(chan *model.PageRef, 100)

			searchPageRef.Page = 0
			go func() {
				receiver.service.Search(searchPageRef, pageChan, interruptChan)
			}()

			count := 0

			for pageRef := range pageChan {
				count++

				err := kafka.SendPageRef(pageRef)

				timeCalc.Step()

				if err != nil {
					log.Error(err)
					return
				}
			}
			log.Print(200, "SCHEDULED: "+strconv.Itoa(count))

			searchPageRef.Page++
			if searchPageRef.Page >= page {
				interruptChan <- true
				break
			}
		}
	}()

}

func (receiver *ScheduleApiImpl) ReadKafka(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("ScheduleKafka")

	kafka := helper.UgbKafkaInstance

	topic := c.Request.URL.Query().Get("topic")
	group := c.Request.URL.Query().Get("group")

	pageChan := kafka.RecvPageRef(topic, group, c.Writer.CloseNotify())

	c.String(200, "[\n")
	isFirst := true
	for pageRef := range pageChan {
		if !isFirst {
			c.String(200, ",\n")
		}
		isFirst = false

		c.JSON(200, pageRef)
	}

	c.String(200, "\n]")
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
