package main

import (
	"backend/api/helper"
	"backend/common/model"
	"bufio"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

import uuid "github.com/satori/go.uuid"

func main() {
	path := os.Args[1]

	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		os.Exit(1)
	}
	r := bufio.NewReader(f)
	Readln(r)
}

func Readln(r *bufio.Reader) {
	ugbMongo := UgbMongo{}
	ugbMongo.Init()

	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	coll := ugbMongo.GetCollection("ug", "pageRef")

	var buffer []*model.PageRefResource

	for {
		line, _, err := r.ReadLine()
		domain := string(line)
		pageRef := makePageRefResource(domain)

		if err != nil {
			log.Error(err)
			break
		}

		buffer = append(buffer, pageRef)

		if len(buffer) >= 10000 {
			insertFlush(coll, buffer)
			buffer = []*model.PageRefResource{}
		}

		timeCalc.Step()
	}

	if len(buffer) >= 0 {
		insertFlush(coll, buffer)
	}
}

func makePageRefResource(domain string) *model.PageRefResource {
	return &model.PageRefResource{
		Id:    uuid.NewV4(),
		Title: domain,
		Data: model.PageRefData{
			Source: "all-domains",
			Url:    "http://" + domain,
			State:  "DOWNLOAD",
			Status: "PENDING",
		},
	}
}

func insertFlush(coll *mongo.Collection, insertList []*model.PageRefResource) {

	opts := new(options.BulkWriteOptions)
	var models []mongo.WriteModel

	for _, pageRef := range insertList {
		writeModel := mongo.NewInsertOneModel()
		writeModel.SetDocument(pageRef)

		models = append(models, writeModel)
	}

	if len(models) == 0 {
		return
	}

	resp, err := coll.BulkWrite(context.Background(), models, opts)
	log.Printf("insert records %d of %d; real insert count: %d", len(insertList), len(models), resp.InsertedCount)

	if err != nil {
		panic(err)
	}
}
