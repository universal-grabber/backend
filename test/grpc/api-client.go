package main

import (
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/api"
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	// initialize grpc
	// Set up a connection to the server.
	conn, err := grpc.Dial("kube.tisserv.net:30004", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Errorf("did not connect: %v", err)
	}

	pageServiceClient := pb.NewPageRefServiceClient(conn)

	req := new(pb.PageRefFetchRequest)
	req.State = base.PageRefState_DOWNLOAD
	resp, err := pageServiceClient.Fetch(context.TODO(), req)

	if err != nil {
		log.Panic(err)
	}

	for {
		item, err := resp.Recv()

		if err != nil {
			break
		}

		log.Print(item)
	}
}
