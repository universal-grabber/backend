package tasks

import (
	"backend/processor/client"
	"backend/processor/lib"
	"backend/processor/model"
	"bytes"
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"time"
)

type PublishTask struct {
	clients     client.Clients
	processor   lib.Processor
	mongoClient *mongo.Client
	pageDataCol *mongo.Collection
}

func (task *PublishTask) Name() string {
	return "PUBLISH"
}

func (task *PublishTask) Init(clients client.Clients) {
	task.clients = clients

	task.processor = lib.Processor{
		ApiClient:       task.clients.GetApiClient(),
		TaskProcessFunc: task.process,
		TaskName:        task.Name(),
		Parallelism:     100,
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(clients.GetConfig().ParseMongoUri)
	clientOptions.Registry = lib.MongoRegistry

	// Connect to MongoDB
	var err error
	task.mongoClient, err = mongo.Connect(context.TODO(), clientOptions)

	lib.Check(err)

	if err != nil {
		log.Panic(err)
	}

	// Check the connection
	err = task.mongoClient.Ping(context.TODO(), nil)

	task.pageDataCol = task.mongoClient.Database("ug").Collection("pageData")

	lib.Check(err)
}

func (task *PublishTask) Run() {
	task.processor.Start()

	log.Print(task.Name(), " task started processing")

	task.processor.Wait()

	log.Print(task.Name(), " task stopped processing")
}

func (task *PublishTask) process(pageRef *model.PageRef) *model.PageRef {
	lib.PageRefLogger(pageRef, "start-publish").
		Trace("starting publish process")

	cur, err := task.pageDataCol.Find(context.TODO(), bson.M{"_id": pageRef.Id})

	lib.CheckWithPageRef(err, pageRef)

	if cur.Next(context.TODO()) {
		data := new(model.PageData)

		err = cur.Decode(data)

		lib.CheckWithPageRef(err, pageRef)

		task.upload(data, pageRef)

		pageRef.Status = model.FINISHED

		return pageRef
	} else {
		lib.PageRefLogger(pageRef, "publish-record-not-found").
			Warn("record not found to publish")
	}

	pageRef.Status = model.FAILED

	lib.PageRefLogger(pageRef, "finish-publish").
		Trace("publish operation finished")

	return pageRef
}

func (task *PublishTask) upload(data *model.PageData, pageRef *model.PageRef) {
	var buffer []*model.Record

	record := data.Record
	record.PublishDate = time.Now().Truncate(1 * time.Second).UTC().Unix()

	buffer = append(buffer, record)
	content, err := json.Marshal(buffer)

	lib.CheckWithPageRef(err, pageRef)

	req, err := http.NewRequest("POST", "https://api.ug.tisserv.net/api/1.0/records", bytes.NewReader(content))

	lib.CheckWithPageRef(err, pageRef)

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "BEARER 74616c6568736d61696c40676d61696c2e636f6d:rk4VIiiW:e044f94db3aae8d914b4cc72b0b37fbec506b137")

	resp, err := http.DefaultClient.Do(req)

	lib.CheckWithPageRef(err, pageRef)

	if resp.StatusCode != 200 {
		respBytes, err := ioutil.ReadAll(resp.Body)

		lib.PageRefLogger(pageRef, "upload-error").
			Errorf("could not upload: %d / %s / %s", resp.StatusCode, string(respBytes), err)
	}
}
