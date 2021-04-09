package service

import (
	"backend/api/const"
	context2 "backend/api/context"
	"backend/api/helper"
	"backend/api/model"
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

type TagsServiceImpl struct {
}

func (service *TagsServiceImpl) ApplyTagsForWebsite(websiteName string) {
	context2.GetSchedulerService().ReloadWebsites()
	pageRefCol := helper.UgbMongoInstance.GetCollection(_const.UgbMongoDb, "pageRef")
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("T")

	cursor, err := pageRefCol.Find(context.TODO(), bson.M{"websiteName": websiteName})

	check(err)

	pageRefUpdateChan := make(chan *model.PageRef, 100)
	defer cursor.Close(context.TODO())

	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		pageRefBufferedChan := helper.BufferChan(pageRefUpdateChan)

		for buffer := range pageRefBufferedChan {
			var writeData []mongo.WriteModel

			for _, pageRef := range buffer {
				writeData = append(writeData, &mongo.UpdateOneModel{
					Filter: bson.M{"_id": pageRef.Id},
					Update: bson.M{
						"$set": bson.M{
							"tags": pageRef.Tags,
						},
					},
				})
			}
			log.Print("begin")
			resp, err := pageRefCol.BulkWrite(context.TODO(), writeData)
			log.Print(resp)
			log.Print("end")

			check(err)
		}
		wg.Done()
	}()

	for cursor.Next(context.TODO()) {
		pageRef := new(model.PageRef)

		err := cursor.Decode(pageRef)

		check(err)

		oldTags := pageRef.Tags
		pageRef.Tags = &[]string{}
		context2.GetSchedulerService().ConfigurePageRef(pageRef)

		if pageRef.Tags != nil && (oldTags == nil || !Equal(*oldTags, *pageRef.Tags)) {
			check(err)
			pageRefUpdateChan <- pageRef
		}
		timeCalc.Step()
	}

	close(pageRefUpdateChan)
	wg.Wait()
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
