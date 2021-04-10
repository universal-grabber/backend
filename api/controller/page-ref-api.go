package controller

import (
	"backend/api/const"
	context2 "backend/api/context"
	"backend/api/helper"
	"backend/api/model"
	"backend/api/service"
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	r.PATCH("/api/1.0/page-refs/bulk", receiver.BulkUpsert)
	r.POST("/api/1.0/page-refs/bulk", receiver.BulkInsert)
	r.POST("/api/1.0/page-refs/update-state", receiver.UpdateStatesBulk)
	r.PUT("/api/1.0/page-refs/update-state", receiver.UpdateStatesBulk)
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
		c.String(200, pageRef.Url+"\n")
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
	db := helper.UgbMongoInstance

	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	var list []model.PageRef
	var insertList []model.PageRef
	var insertUrls []string

	c.BindJSON(&list)

	col := db.GetCollection(_const.UgbMongoDb, "pageRef")

	opts := new(options.BulkWriteOptions)
	var models []mongo.WriteModel

	existingItems := make(map[string]bool)

	for _, pageRef := range list {
		context2.GetSchedulerService().ConfigurePageRef(&pageRef)

		if contains(*pageRef.Tags, "delete") {
			continue
		}

		if !contains(*pageRef.Tags, "allow-import") {
			continue
		}

		if existingItems[pageRef.Url] {
			continue
		}

		existingItems[pageRef.Url] = true

		insertList = append(insertList, pageRef)
		insertUrls = append(insertUrls, pageRef.Url)
	}

	existingUrls := receiver.service.PageRefExistsMultiViaUrl(insertUrls)

	for _, pageRef := range insertList {
		if contains(existingUrls, pageRef.Url) {
			continue
		}
		writeModel := mongo.NewInsertOneModel()
		writeModel.SetDocument(pageRef)

		models = append(models, writeModel)
	}

	if len(models) == 0 {
		return
	}

	resp, err := col.BulkWrite(context.Background(), models, opts)
	log.Printf("insert records %d of %d; real insert count: %d", len(list), len(models), resp.InsertedCount)

	if err != nil {
		panic(err)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
