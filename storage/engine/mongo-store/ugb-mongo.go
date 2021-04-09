package simple

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	UgbMongoHtmlUri = "mongodb://10.0.1.77:27018"
)

type UgbMongo struct {
	client *mongo.Client
}

func (obj *UgbMongo) Init() {
	if obj.client != nil {
		return
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(UgbMongoHtmlUri)
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
}

func (obj *UgbMongo) GetCollection(database string, name string) *mongo.Collection {
	return obj.client.Database(database).Collection(name)
}
