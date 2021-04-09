package controller

import (
	"backend/api/const"
	"backend/api/helper"
	"backend/api/model"
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type WebsiteApiImpl struct {
}

func (receiver *WebsiteApiImpl) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/1.0/websites", receiver.List)
}

func (receiver WebsiteApiImpl) List(c *gin.Context) {
	db := helper.UgbMongoInstance
	db.Init()

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

	c.JSON(200, list)
}
