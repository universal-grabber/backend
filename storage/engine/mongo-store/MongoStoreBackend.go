package simple

import (
	"context"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoStoreBackend struct {
	inited   bool
	ugbMongo *UgbMongo
	col      *mongo.Collection
}

func (storage *MongoStoreBackend) Get(id uuid.UUID) ([]byte, error) {
	filter := bson.M{"_id": id.String()}
	cur, err := storage.col.Find(context.TODO(), filter)

	if err != nil {
		return nil, err
	}

	if cur.Next(context.TODO()) {
		pageHtml := new(MongoStorePageHtml)

		if err = cur.Decode(pageHtml); err != nil {
			log.Error(err)
		}

		return []byte(pageHtml.Content), nil
	}

	return nil, nil
}

func (storage *MongoStoreBackend) Delete(id uuid.UUID) (bool, error) {
	filter := bson.M{"_id": id.String()}

	resp, err := storage.col.DeleteOne(context.TODO(), filter)

	return resp != nil && resp.DeletedCount > 0, err
}

func (storage *MongoStoreBackend) Exists(id uuid.UUID) (bool, error) {
	filter := bson.M{"_id": id.String()}

	cur, err := storage.col.Find(context.TODO(), filter)

	if err != nil {
		return false, err
	}

	if cur.Next(context.TODO()) {
		return true, nil
	}

	return false, nil
}

func (storage *MongoStoreBackend) Add(id uuid.UUID, data []byte) error {
	err := storage.TryAdd(id, data)

	// retry if cannot insert
	if err != nil {
		storage.Delete(id)
	} else {
		return err
	}

	err = storage.TryAdd(id, data)

	return err
}

func (storage *MongoStoreBackend) TryAdd(id uuid.UUID, data []byte) error {
	pageHtml := new(MongoStorePageHtml)
	pageHtml.Id = id.String()
	pageHtml.Content = string(data)
	pageHtml.ContentSize = len(data)

	_, err := storage.col.InsertOne(context.TODO(), pageHtml)

	return err
}

func (storage *MongoStoreBackend) init(dbName string, colName string) {
	if !storage.inited {
		storage.ugbMongo = new(UgbMongo)
		storage.ugbMongo.Init()
	}

	storage.col = storage.ugbMongo.GetCollection(dbName, colName)

	storage.inited = true
}

func NewInstance(dbName string, colName string) *MongoStoreBackend {
	instance := new(MongoStoreBackend)

	instance.init(dbName, colName)

	return instance
}
