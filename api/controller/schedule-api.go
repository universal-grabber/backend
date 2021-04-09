package controller

import (
	context2 "backend/api/context"
	"github.com/gin-gonic/gin"
)

type ScheduleApiImpl struct {
}

func (receiver *ScheduleApiImpl) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/1.0/schedule/apply-tags", receiver.applyTags)
	r.GET("/api/1.0/schedule/website", receiver.manualScheduleWebsite)
	r.GET("/api/1.0/schedule/reload", receiver.reload)
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
