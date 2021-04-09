package mongo_old

import (
	"context"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SimpleMongoBackend struct {
	inited   bool
	ugbMongo *UgbMongo
	col      *mongo.Collection
}

func (storage *SimpleMongoBackend) Get(id uuid.UUID) ([]byte, error) {
	filter := bson.M{"_id": id}
	cur, err := storage.col.Find(context.TODO(), filter)

	if err != nil {
		return nil, err
	}

	if cur.Next(context.TODO()) {
		pageHtml := new(SimplePageHtml)

		if err = cur.Decode(pageHtml); err != nil {
			log.Fatal(err)
		}

		return []byte(pageHtml.Content), nil
	}

	return nil, nil
}

func (storage *SimpleMongoBackend) Delete(id uuid.UUID) (bool, error) {
	filter := bson.M{"_id": id}

	resp, err := storage.col.DeleteOne(context.TODO(), filter)

	return resp != nil && resp.DeletedCount > 0, err
}

func (storage *SimpleMongoBackend) Exists(id uuid.UUID) (bool, error) {
	filter := bson.M{"_id": id}

	cur, err := storage.col.Find(context.TODO(), filter)

	if err != nil {
		return false, err
	}

	if cur.Next(context.TODO()) {
		return true, nil
	}

	return false, nil
}

func (storage *SimpleMongoBackend) Add(id uuid.UUID, data []byte) error {
	pageHtml := new(SimplePageHtml)
	pageHtml.Id = id
	pageHtml.Content = string(data)
	pageHtml.ContentSize = len(data)

	_, err := storage.col.InsertOne(context.TODO(), pageHtml)

	return err
}

func (storage *SimpleMongoBackend) init(dbName string, colName string) {
	if !storage.inited {
		storage.ugbMongo = new(UgbMongo)
		storage.ugbMongo.Init()
	}

	storage.col = storage.ugbMongo.GetCollection(dbName, colName)

	storage.inited = true
}

func NewInstance(dbName string, colName string) *SimpleMongoBackend {
	instance := new(SimpleMongoBackend)

	instance.init(dbName, colName)

	return instance
}
