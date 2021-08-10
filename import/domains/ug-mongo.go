package main

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
	clientOptions := options.Client().ApplyURI("mongodb://ug.tisserv.net:27017")
	clientOptions.Auth = new(options.Credential)
	clientOptions.Auth.Username = "admin"
	clientOptions.Auth.Password = "256b93c7870005dca0db75c3ff0b026f"
	clientOptions.Auth.AuthSource = "admin"
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
		Keys: bson.M{"data.url": 1},
		Options: &options.IndexOptions{
			Unique: truePointer,
		},
	})

	obj.GetCollection(_const.UgbMongoDb, "pageRef").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"data.websiteName": 1, "data.state": 1, "data.status": 1},
	})

	obj.GetCollection(_const.UgbMongoDb, "pageRef").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"data.state": 1, "data.status": 1},
	})

	obj.GetCollection(_const.UgbMongoDb, "pageRef").Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.M{"data.websiteName": 1, "data.tags": 1},
	})
}
