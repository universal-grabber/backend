package service

import (
	_const "backend/api/const"
	context2 "backend/api/context"
	"backend/api/helper"
	"backend/api/model"
	"backend/api/util"
	"context"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PageRefService struct {
}

func (service *PageRefService) Search(context context.Context, searchPageRef *model.SearchPageRef, pageChan chan *model.PageRef) {
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
		websitePageChan := util.SearchByFilter(context, db, util.PrepareFilter(searchPageRef), opts)
		util.RedirectChan(pageChan, websitePageChan)
	} else {
		websites := util.ListWebsites(db)

		var chanArr []chan *model.PageRef
		for _, website := range websites {

			filters := util.PrepareFilter(searchPageRef)
			filters["data.websiteName"] = website.Name
			chanArr = append(chanArr, util.SearchByFilter(context, db, filters, opts))
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

func (service *PageRefService) asyncUpdateRecords(updateChan chan *model.PageRef, toState string, toStatus string, timeCalc *helper.TimeCalc) {
	col := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef")

	go func() {
		var buffer []uuid.UUID

		for pageRef := range updateChan {
			buffer = append(buffer, pageRef.Id)

			if len(buffer) > 500 {
				service.flushUpdate(col, buffer, toState, toStatus)
				buffer = []uuid.UUID{}
			}

			timeCalc.Step()
		}

		if len(buffer) > 0 {
			service.flushUpdate(col, buffer, toState, toStatus)
		}
	}()
}

func (service *PageRefService) flushUpdate(col *mongo.Collection, buffer []uuid.UUID, toState string, toStatus string) {
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

func (service *PageRefService) UpdateStatesBulk2(context context.Context, searchPageRef *model.SearchPageRef, toState string, toStatus string) (chan *model.PageRef, chan *model.PageRef) {
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
		service.Search(context, searchPageRef, pageChan)
	}()

	updateChan := make(chan *model.PageRef, 100)

	for i := 0; i < 3; i++ {
		service.asyncUpdateRecords(updateChan, toState, toStatus, timeCalc)
	}

	return pageChan, updateChan
}

func (service *PageRefService) pageRefExistsMultiViaUrl(urls []string) []string {
	opts := new(options.FindOptions)
	opts.Projection = bson.D{{"url", 1}}

	filter := bson.D{{"url", bson.M{"$in": urls}}}

	cursor, err := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef").Find(context.TODO(), filter, opts)

	if err != nil {
		log.Panic(err)
	}

	ctx := context.Background()

	defer cursor.Close(context.TODO())

	var existingUrls = append([]string{}, "dummy-url")

	for cursor.Next(ctx) {
		record := bson.M{}

		cursor.Decode(record)

		existingUrls = append(existingUrls, record["url"].(string))
	}

	return existingUrls
}

func (service *PageRefService) PageRefExists(id uuid.UUID) bool {
	opts := new(options.FindOptions)
	opts.Projection = bson.D{{"_id", 0}}

	cursor, err := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef").Find(context.TODO(), bson.D{{"_id", id}}, opts)

	if err != nil {
		log.Print(err)
		return false
	}

	ctx := context.Background()

	defer cursor.Close(context.TODO())

	for cursor.Next(ctx) {
		return true
	}

	return false
}

func (service *PageRefService) BulkWrite(list []model.PageRef) []model.PageRef {
	opLog := log.WithField("operation", "bulkWrite2")

	col := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef")

	opts := new(options.BulkWriteOptions)
	var models []mongo.WriteModel

	var result []model.PageRef

	for _, pageRef := range list {
		opLog.Debug("update page-ref")

		oldId := pageRef.Id
		context2.GetSchedulerService().ConfigurePageRef(&pageRef)
		opLog.Tracef("pageRef configured to state: %s; status: %s", pageRef.Data.State, pageRef.Data.Status)

		result = append(result, pageRef)

		if util.Contains(*pageRef.Data.Tags, "delete") {
			// delete dangling page-ref
			writeModel := mongo.NewDeleteOneModel()
			writeModel.Filter = bson.M{"_id": oldId}

			models = append(models, writeModel)

			opLog.Debug("deleting dangling record1")
			continue
		}
		if oldId.String() != pageRef.Id.String() {
			// delete dangling page-ref
			writeModel := mongo.NewDeleteOneModel()
			writeModel.Filter = bson.M{"_id": oldId}

			models = append(models, writeModel)
			opLog.Debug("deleting dangling record1")
		}

		writeModel := mongo.NewUpdateOneModel()
		writeModel.Upsert = new(bool)
		*writeModel.Upsert = false
		writeModel.Filter = bson.M{"_id": pageRef.Id}
		writeModel.Update = bson.M{"$set": bson.M{
			"data": bson.M{
				"state":   pageRef.Data.State,
				"status":  pageRef.Data.Status,
				"options": pageRef.Data.Options,
			},
		}}

		models = append(models, writeModel)
	}

	res, err := col.BulkWrite(context.Background(), models, opts)

	if err != nil {
		log.Error(err)
	}

	log.Infof("Bulk write result: WriteCount => %d; ModifiedCount => %d; ModifiedCount => %d; DeletedCount => %d; UpsertedCount => %d; ", len(models), res.ModifiedCount, res.InsertedCount, res.DeletedCount, res.UpsertedCount)

	return result
}

func (service *PageRefService) BulkInsert(list []model.PageRef) {
	db := helper.UgbMongoInstance

	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	var insertList []model.PageRef
	var insertUrls []string

	col := db.GetCollection(_const.UgbMongoDb, "pageRef")

	opts := new(options.BulkWriteOptions)
	var models []mongo.WriteModel

	existingItems := make(map[string]bool)

	for _, pageRef := range list {
		helper.PageRefLogger(&pageRef, "bulk-insert").Debug("inserting page-ref")
		context2.GetSchedulerService().ConfigurePageRef(&pageRef)
		context2.GetSchedulerService().ConfigurePageUrl(&pageRef)

		if util.Contains(*pageRef.Data.Tags, "delete") {
			helper.PageRefLogger(&pageRef, "bulk-insert").Debug("inserting page-ref filtered by delete tag")
			continue
		}

		if !util.Contains(*pageRef.Data.Tags, "allow-import") {
			helper.PageRefLogger(&pageRef, "bulk-insert").Debug("inserting page-ref filtered by allow import tag")
			continue
		}

		if existingItems[pageRef.Data.Url] {
			helper.PageRefLogger(&pageRef, "bulk-insert").Debug("inserting page-ref filtered by existing rule")
			continue
		}

		existingItems[pageRef.Data.Url] = true

		insertList = append(insertList, pageRef)
		insertUrls = append(insertUrls, pageRef.Data.Url)
	}

	existingUrls := service.pageRefExistsMultiViaUrl(insertUrls)

	for _, pageRef := range insertList {
		if util.Contains(existingUrls, pageRef.Data.Url) {
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
