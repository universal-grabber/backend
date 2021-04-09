package context

import (
	"backend/api/model"
	"github.com/gin-gonic/gin"
)

// generic api interface
type Api interface {
	RegisterRoutes(*gin.Engine)
}

// service
type PageRefService interface {
}

type TagsService interface {
	ApplyTagsForWebsite(websiteName string)
}

type SchedulerService interface {
	Run()
	ConfigurePageRef(pageRef *model.PageRef)
	ScheduleWebsiteManual(websiteName string)
	ReloadWebsites()
}

const WebsiteApiInstance = "WebsiteApiInstance"
const PageRefApiInstance = "PageRefApiInstance"
const ScheduleApiInstance = "ScheduleApiInstance"

const PageRefServiceInstance = "PageRefServiceInstance"
const SchedulerInstance = "PageRefServiceInstance"
const TagsServiceInstance = "TagsService"

// api instances
func GetPageRefApi() Api {
	return Get(PageRefApiInstance).(Api)
}

func GetWebsiteApi() Api {
	return Get(WebsiteApiInstance).(Api)
}

func GetScheduleApi() Api {
	return Get(ScheduleApiInstance).(Api)
}

// services

func GetPageRefService() PageRefService {
	return Get(PageRefServiceInstance).(PageRefService)
}

func GetSchedulerService() SchedulerService {
	return Get(SchedulerInstance).(SchedulerService)
}
func GetTagsService() TagsService {
	return Get(TagsServiceInstance).(TagsService)
}
