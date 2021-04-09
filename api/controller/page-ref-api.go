package controller

import (
	"backend/api/const"
	context2 "backend/api/context"
	"backend/api/helper"
	"backend/api/model"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PageRefApiImpl struct {
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
		search(searchPageRef, pageChan, c.Writer.CloseNotify())
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
		search(searchPageRef, pageChan, c.Writer.CloseNotify())
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

	pageChan, updateChan := receiver.UpdateStatesBulk2(searchPageRef, toState, toStatus, c.Writer.CloseNotify())

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

func (receiver *PageRefApiImpl) UpdateStatesBulk2(searchPageRef *model.SearchPageRef, toState string, toStatus string, interruptChan <-chan bool) (chan *model.PageRef, chan *model.PageRef) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	if len(toState) == 0 {
		panic("toState is missing")
	}

	if len(toStatus) == 0 {
		panic("toStatus is missing")
	}

	pageChan := make(chan *model.PageRef, 100)

	// search async
	go func() {
		search(searchPageRef, pageChan, interruptChan)
	}()

	updateChan := make(chan *model.PageRef, 100)

	for i := 0; i < 3; i++ {
		receiver.asyncUpdateRecords(updateChan, toState, toStatus, timeCalc)
	}

	return pageChan, updateChan
}

func (receiver *PageRefApiImpl) asyncUpdateRecords(updateChan chan *model.PageRef, toState string, toStatus string, timeCalc *helper.TimeCalc) {
	col := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef")

	go func() {
		var buffer []uuid.UUID

		for pageRef := range updateChan {
			buffer = append(buffer, pageRef.Id)

			if len(buffer) > 500 {
				receiver.flushUpdate(col, buffer, toState, toStatus)
				buffer = []uuid.UUID{}
			}

			timeCalc.Step()
		}

		if len(buffer) > 0 {
			receiver.flushUpdate(col, buffer, toState, toStatus)
		}
	}()
}

func (receiver *PageRefApiImpl) flushUpdate(col *mongo.Collection, buffer []uuid.UUID, toState string, toStatus string) {
	var ids []primitive.Binary

	for index := range buffer {
		ids = append(ids, primitive.Binary{Data: buffer[index].Bytes(), Subtype: 3})
	}

	filter := bson.M{}
	update := bson.M{}

	filter["_id"] = bson.M{
		"$in": ids,
	}

	update["$set"] = bson.M{"state": toState, "status": toStatus}

	_, err := col.UpdateMany(context.Background(), filter, update)

	if err != nil {
		panic(err)
	}
}

func search(searchPageRef *model.SearchPageRef, pageChan chan *model.PageRef, interruptChan <-chan bool) {
	db := helper.UgbMongoInstance
	opts := new(options.FindOptions)
	opts.Limit = new(int64)
	*opts.Limit = 100

	if searchPageRef.PageSize > 0 {
		*opts.Limit = int64(searchPageRef.PageSize)
	}

	if searchPageRef.Page > 0 {
		opts.Skip = new(int64)
		*opts.Skip = *opts.Limit * int64(searchPageRef.Page)
	}

	if !searchPageRef.FairSearch {
		websitePageChan := searchByFilter(db, prepareFilter(searchPageRef), interruptChan, opts)
		redirectChan(pageChan, websitePageChan)
	} else {
		websites := listWebsites(db)

		var chanArr []chan *model.PageRef
		for _, website := range websites {

			filters := prepareFilter(searchPageRef)
			filters["websiteName"] = website.Name
			chanArr = append(chanArr, searchByFilter(db, filters, interruptChan, opts))
		}

		isFound := true

		for isFound {
			isFound = false

			for _, chanItem := range chanArr {
				item, ok := <-chanItem

				if ok {
					isFound = true
					pageChan <- item
				}
			}

			// if we don't found any record in
			if !isFound {
				break
			}
		}
	}

	close(pageChan)
}

func redirectChan(to chan *model.PageRef, from chan *model.PageRef) {
	for item := range from {
		to <- item
	}
}

func prepareFilter(searchPageRef *model.SearchPageRef) bson.M {
	filters := bson.M{}

	if !searchPageRef.FairSearch && len(searchPageRef.WebsiteName) > 0 {
		filters["websiteName"] = searchPageRef.WebsiteName
	}

	if len(searchPageRef.State) > 0 {
		filters["state"] = searchPageRef.State
	}

	if len(searchPageRef.Status) > 0 {
		filters["status"] = searchPageRef.Status
	}

	if len(searchPageRef.Tags) > 0 {
		filters["tags"] = bson.M{"$in": searchPageRef.Tags}
	}
	return filters
}

func listWebsites(db *helper.UgbMongo) []model.WebSite {
	websitesCursor, err := db.GetCollection(_const.UgbMongoDb, "website").Find(context.Background(), bson.M{})

	if err != nil {
		log.Panic(err)
	}

	var list []model.WebSite

	for websitesCursor.Next(context.Background()) {
		website := new(model.WebSite)

		err = websitesCursor.Decode(website)

		if err != nil {
			log.Panic(err)
		}

		list = append(list, *website)
	}

	err = websitesCursor.Close(context.TODO())

	if err != nil {
		log.Panic(err)
	}
	return list
}

func searchByFilter(db *helper.UgbMongo, filters bson.M, interruptChan <-chan bool, opts *options.FindOptions) chan *model.PageRef {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panicing searchByFilter: %s", r)
		}
	}()

	pageChan := make(chan *model.PageRef)

	go func() {
		defer func() {
			close(pageChan)
		}()

		cursor, err := db.GetCollection(_const.UgbMongoDb, "pageRef").Find(context.Background(), filters, opts)

		if err != nil {
			panic(err)
		}

		for cursor.Next(context.Background()) {
			select {
			case <-interruptChan:
				fmt.Print("Stopping receiving items as client disconnected\n")
				return
			default:
			}
			pageRef := new(model.PageRef)

			err := bson.UnmarshalWithRegistry(helper.MongoRegistry, cursor.Current, pageRef)

			if err != nil {
				break
			}

			pageChan <- pageRef
		}

		err = cursor.Close(context.TODO())

		if err != nil {
			panic(err)
		}
	}()

	return pageChan
}

func (receiver *PageRefApiImpl) BulkUpsert(c *gin.Context) {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	var list []model.PageRef

	c.BindJSON(&list)

	receiver.BulkWrite2(list)
}

func (receiver *PageRefApiImpl) BulkWrite2(list []model.PageRef) {
	col := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef")

	opts := new(options.BulkWriteOptions)
	var models []mongo.WriteModel

	for _, pageRef := range list {
		oldId := pageRef.Id
		context2.GetSchedulerService().ConfigurePageRef(&pageRef)
		if contains(*pageRef.Tags, "delete") {
			// delete dangling page-ref
			writeModel := mongo.NewDeleteOneModel()
			writeModel.Filter = bson.M{"_id": oldId}

			models = append(models, writeModel)
			continue
		}
		if oldId.String() != pageRef.Id.String() {
			// delete dangling page-ref
			writeModel := mongo.NewDeleteOneModel()
			writeModel.Filter = bson.M{"_id": oldId}

			models = append(models, writeModel)
		}

		writeModel := mongo.NewUpdateOneModel()
		writeModel.Upsert = new(bool)
		*writeModel.Upsert = true
		writeModel.Filter = bson.M{"_id": pageRef.Id}
		writeModel.Update = bson.M{"$set": pageRef}

		models = append(models, writeModel)
	}

	_, err := col.BulkWrite(context.Background(), models, opts)

	if err != nil {
		panic(err)
	}
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

	existingUrls := receiver.PageRefExistsMultiViaUrl(insertUrls)

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

func (receiver *PageRefApiImpl) PageRefExistsMultiViaUrl(urls []string) []string {
	opts := new(options.FindOptions)
	opts.Projection = bson.D{{"url", 1}}

	filter := bson.D{{"url", bson.M{"$in": urls}}}

	cursor, err := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef").Find(context.TODO(), filter, opts)

	if err != nil {
		log.Panic(err)
	}

	ctx := context.Background()

	defer cursor.Close(ctx)

	var existingUrls = append([]string{}, "dummy-url")

	for cursor.Next(ctx) {
		record := bson.M{}

		cursor.Decode(record)

		existingUrls = append(existingUrls, record["url"].(string))
	}

	return existingUrls
}

func (receiver *PageRefApiImpl) PageRefExists(id uuid.UUID) bool {
	opts := new(options.FindOptions)
	opts.Projection = bson.D{{"_id", 0}}

	cursor, err := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef").Find(context.TODO(), bson.D{{"_id", id}}, opts)

	if err != nil {
		log.Print(err)
		return false
	}

	ctx := context.Background()

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		return true
	}

	return false
}
