package util

import (
	_const "backend/api/const"
	"backend/api/helper"
	"backend/api/model"
	"backend/common"
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RedirectChan(to chan *model.PageRef, from chan *model.PageRef) {
	for item := range from {
		to <- item
	}
}

func PrepareFilter(searchPageRef *model.SearchPageRef) bson.M {
	filters := bson.M{}

	if !searchPageRef.FairSearch && len(searchPageRef.WebsiteName) > 0 {
		filters["data.websiteName"] = searchPageRef.WebsiteName
	}

	if len(searchPageRef.State) > 0 {
		filters["data.state"] = searchPageRef.State
	}

	if len(searchPageRef.Status) > 0 {
		filters["data.status"] = searchPageRef.Status
	}

	if len(searchPageRef.Tags) > 0 {
		filters["data.tags"] = bson.M{"$in": searchPageRef.Tags}
	}
	return filters
}

func ListWebsites(db *helper.UgbMongo) []model.WebSite {
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

func SearchByFilter(ctx context.Context, db *helper.UgbMongo, filters bson.M, opts *options.FindOptions) chan *model.PageRef {
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

		cursor, err := db.GetCollection(_const.UgbMongoDb, "pageRef").Find(ctx, filters, opts)

		if err != nil {
			panic(err)
		}

		for cursor.Next(ctx) {
			pageRef := new(model.PageRef)

			err := bson.UnmarshalWithRegistry(helper.MongoRegistry, cursor.Current, pageRef)

			if err != nil {
				break
			}

			pageChan <- pageRef
		}

		err = cursor.Close(context.TODO())

		if err != nil {
			common.UseLogger(ctx).Error(err)
		}
	}()

	return pageChan
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
