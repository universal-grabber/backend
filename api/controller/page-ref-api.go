package controller

import (
	"backend/api/helper"
	"backend/api/model"
	"backend/api/service"
	"github.com/gin-gonic/gin"
)

type PageRefApiImpl struct {
	service *service.PageRefService
}

func (receiver *PageRefApiImpl) Init() {
	receiver.service = new(service.PageRefService)
}

func (receiver *PageRefApiImpl) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/1.0/page-refs", receiver.List)
	r.GET("/api/1.0/page-refs/urls", receiver.ListUrls)
}

func (receiver *PageRefApiImpl) List(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	searchPageRef := new(model.SearchPageRef)
	err := helper.ParseRequestQuery(c.Request, searchPageRef)

	if err != nil {
		panic(err)
	}

	pageChan := make(chan *model.PageRef, 100)

	// search async
	go func() {
		receiver.service.Search(searchPageRef, pageChan, c.Writer.CloseNotify())
	}()

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

func (receiver *PageRefApiImpl) ListUrls(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	searchPageRef := new(model.SearchPageRef)
	err := helper.ParseRequestQuery(c.Request, searchPageRef)

	if err != nil {
		panic(err)
	}

	pageChan := make(chan *model.PageRef, 100)

	// search async
	go func() {
		receiver.service.Search(searchPageRef, pageChan, c.Writer.CloseNotify())
	}()

	for pageRef := range pageChan {
		c.String(200, pageRef.Data.Url+"\n")
	}

}

func (receiver *PageRefApiImpl) UpdateStatesBulk(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	searchPageRef := new(model.SearchPageRef)
	err := helper.ParseRequestQuery(c.Request, searchPageRef)

	if err != nil {
		panic(err)
	}

	toState := c.Request.URL.Query().Get("toState")
	toStatus := c.Request.URL.Query().Get("toStatus")

	if len(toState) == 0 {
		panic("toState is missing")
	}

	if len(toStatus) == 0 {
		panic("toStatus is missing")
	}

	pageChan, updateChan := receiver.service.UpdateStatesBulk2(searchPageRef, toState, toStatus, c.Writer.CloseNotify())

	defer close(updateChan)

	c.String(200, "[\n")
	isFirst := true
	for pageRef := range pageChan {
		if !isFirst {
			c.String(200, ",\n")
		}
		isFirst = false

		helper.PageRefLogger(pageRef, "request-update-state").Debug("pageRef state updated")

		c.JSON(200, pageRef)
		updateChan <- pageRef
	}

	c.String(200, "\n]")
}

func (receiver *PageRefApiImpl) BulkUpsert(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	var list []model.PageRef

	c.BindJSON(&list)

	receiver.service.BulkWrite2(list)
}

func (receiver *PageRefApiImpl) BulkInsert(c *gin.Context) {
	var list []model.PageRef
	c.BindJSON(&list)

	receiver.service.BulkInsert(list)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
