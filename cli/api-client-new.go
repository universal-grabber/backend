package main

import (
	pb "backend/gen/proto/service/api"
	"backend/processor/model"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type ApiClient struct {
	pageServiceClient pb.PageRefServiceClient
	connection        *grpc.ClientConn
}

func (client *ApiClient) Init(config model.Config) {
	// initialize grpc
	// Set up a connection to the server.
	conn, err := grpc.Dial("10.0.1.77:30004", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client.connection = conn

	client.pageServiceClient = pb.NewPageRefServiceClient(conn)
}
