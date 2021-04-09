package helper

import (
	"backend/api/const"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

type UgbMongo struct {
	client *mongo.Client
}

var (
	UgbMongoInstance = new(UgbMongo)
)

func (obj *UgbMongo) Init() {
	if obj.client != nil {
		return
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(_const.UgbMongoUri)
	clientOptions.Registry = MongoRegistry

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	obj.client = client

	if err != nil {
		log.Panic(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Panic(err)
	}

	// ensure indexes
	obj.initIndexes()
}

func (obj *UgbMongo) GetCollection(database string, name string) *mongo.Collection {
	return obj.client.Database(database).Collection(name)
}

func (obj *UgbMongo) GetDatabase(database string) *mongo.Database {
	return obj.client.Database(database)
}

func (obj *UgbMongo) initIndexes() {
	truePointer := new(bool)
	*truePointer = true

	obj.GetCollection(_const.UgbMongoDb, "pageRef").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"url": 1},
		Options: &options.IndexOptions{
			Unique: truePointer,
		},
	})

	obj.GetCollection(_const.UgbMongoDb, "pageRef").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"websiteName": 1, "state": 1, "status": 1},
	})

	obj.GetCollection(_const.UgbMongoDb, "pageRef").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"state": 1, "status": 1},
	})

	obj.GetCollection(_const.UgbMongoDb, "pageRef").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"websiteName": 1, "tags": 1},
	})
}
