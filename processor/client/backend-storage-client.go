package client

import (
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/storage"
	"backend/processor/lib"
	"backend/processor/model"
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type BackendStorageClient struct {
	config               model.Config
	connection           *grpc.ClientConn
	storageServiceClient pb.StorageServiceClient
}

func (client *BackendStorageClient) Init(config model.Config) {
	client.config = config

	// initialize grpc
	// Set up a connection to the server.
	conn, err := grpc.Dial(config.UgbStorageUri, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Errorf("did not connect: %v", err)
	}

	client.connection = conn

	client.storageServiceClient = pb.NewStorageServiceClient(conn)
}

func (client *BackendStorageClient) Store(item *base.PageRef) *pb.StoreResult {
	res, err := client.storageServiceClient.Store(context.TODO(), item)

	lib.Check(err)

	return res
}

func (client *BackendStorageClient) Get(item *base.PageRef) *pb.StoreResult {
	res, err := client.storageServiceClient.Get(context.TODO(), item)

	lib.Check(err)

	return res
}
