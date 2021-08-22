package tasks

import (
	"backend/common"
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/storage"
	"backend/processor/client"
	"backend/processor/lib"
	"backend/processor/model"
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	parseTaskMetricsRegistry = common.NewMeter("ugb-parse-task")
)

type ParseTask struct {
	clients     client.Clients
	processor   lib.Processor
	mongoClient *mongo.Client
	pageDataCol *mongo.Collection
}

func (task *ParseTask) Name() string {
	return "PARSE"
}

func (task *ParseTask) Init(clients client.Clients) {
	task.clients = clients

	task.processor = lib.Processor{
		ApiClient:       task.clients.GetApiClient(),
		TaskProcessFunc: task.process,
		State:           base.PageRefState_PARSE,
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

func (task *ParseTask) Run() {
	task.processor.Start()

	log.Print(task.Name(), " task started processing")

	task.processor.Wait()

	log.Print(task.Name(), " task stopped processing")
}

func (task *ParseTask) process(pageRef *base.PageRef) *base.PageRef {
	log.Tracef("page-ref received for download %s", pageRef.Url)

	parseTaskMetricsRegistry.Inc("parse-request", 1, common.PageRefRecordToTags2(*pageRef))

	result := task.clients.GetBackendStorageClient().Get(pageRef)

	if pageRef.Status != base.PageRefStatus_FINISHED {
		pageRef.Status = base.PageRefStatus_FAILED
	}

	if result.Ok {
		pageRef.Status = task.parseItem(result, pageRef)
	} else {
		common.PageRefLogger(pageRef, "fail-to-get-html").
			Warnf("could not get page-ref html")
		pageRef.Status = base.PageRefStatus_FAILED
	}

	parseTaskMetricsRegistry.Inc("parse-response", 1, common.PageRefRecordToTags2(*pageRef))

	return pageRef
}

func (task *ParseTask) parseItem(result *pb.StoreResult, pageRef *base.PageRef) base.PageRefStatus {
	common.PageRefLogger(pageRef, "start-parse").
		Trace("starting parse process")

	defer func() {
		if r := recover(); r != nil {
			common.PageRefLogger(pageRef, "parse-panic").
				Errorf("panicing parse: %s", r)
		}
	}()
	parseResult := result.Content

	res := task.clients.GetModelProcessorClient().Parse(parseResult, pageRef)

	for _, tag := range pageRef.Tags {
		res.Tags = append(res.Tags, tag)
	}

	if res == nil {
		common.PageRefLogger(pageRef, "parse-html-could-not-get").
			Trace("could not get html file")
		return base.PageRefStatus_FAILED
	}

	pageData := new(model.PageData)
	pageData.Id = pageRef.Id
	pageData.Record = res

	_, err := task.pageDataCol.DeleteOne(context.TODO(), bson.M{"_id": pageRef.Id})

	if err != nil {
		lib.CheckWithPageRefLogOnly(err, pageRef)
		return base.PageRefStatus_FAILED
	}

	_, err = task.pageDataCol.InsertOne(context.TODO(), pageData)

	if err != nil {
		lib.CheckWithPageRefLogOnly(err, pageRef)
		return base.PageRefStatus_FAILED
	}

	common.PageRefLogger(pageRef, "finish-parse").
		Trace("finishing parse process")

	return base.PageRefStatus_FINISHED
}
