package client

import (
	"backend/common"
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/api"
	"backend/processor/lib"
	"backend/processor/model"
	"context"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"strconv"
	"sync"
	"time"
)

type ApiClient struct {
	insertLock  sync.Mutex
	updateLock  sync.Mutex
	insertQueue []*base.PageRef
	updateQueue []*base.PageRef
	config      model.Config

	insertSemaphore *semaphore.Weighted

	pageServiceClient pb.PageRefServiceClient
	connection        *grpc.ClientConn
}

func (client *ApiClient) Init(config model.Config) {
	client.config = config
	go client.scheduleInsertChannel()
	go client.scheduleUpdateChannel()

	client.insertSemaphore = semaphore.NewWeighted(10000)

	// initialize grpc
	// Set up a connection to the server.
	conn, err := grpc.Dial(config.UgbApiGrpcUri, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Errorf("did not connect: %v", err)
	}

	client.connection = conn

	client.pageServiceClient = pb.NewPageRefServiceClient(conn)
}

func (client *ApiClient) AcceptPages(state base.PageRefState) chan *base.PageRef {
	timeCalc := new(common.TimeCalc)
	timeCalc.Init("timer1")

	log.Debugf("starting to accept pages for state %s", state)

	req := new(pb.PageRefFetchRequest)
	req.State = state
	req.Websites = client.config.EnabledWebsites

	pageRefReadChannel := make(chan *base.PageRef)

	go func() {
		for {
			resp, err := client.pageServiceClient.Fetch(context.TODO(), req)

			if err != nil {
				log.Warn(err)
				time.Sleep(1 * time.Second)
				continue
			}

			for {
				item, err := resp.Recv()

				if err != nil {
					break
				}

				timeCalc.Step()
				pageRefReadChannel <- item
			}

			// wait for one minute, as queue was empty
			time.Sleep(60 * time.Second)
		}
	}()

	return pageRefReadChannel
}

func (client *ApiClient) InsertPageRef(ref *base.PageRef) {
	err := client.insertSemaphore.Acquire(context.TODO(), 1)

	if err != nil {
		log.Print(err)
	}

	client.insertLock.Lock()

	client.insertQueue = append(client.insertQueue, ref)

	client.insertLock.Unlock()
}

func (client *ApiClient) UpdatePageRef(ref *base.PageRef) {
	client.updateLock.Lock()

	client.updateQueue = append(client.updateQueue, ref)

	client.updateLock.Unlock()
}

func (client *ApiClient) scheduleInsertChannel() {
	for {
		if len(client.insertQueue) > 0 {
			queueCopy := client.insertQueue
			client.insertQueue = append([]*base.PageRef{})

			client.bulkInsert(queueCopy)
			client.insertSemaphore.Release(int64(len(queueCopy)))
		}
		time.Sleep(1 * time.Second)
	}
}

func (client *ApiClient) scheduleUpdateChannel() {
	for {
		if len(client.updateQueue) > 0 {
			queueCopy := client.updateQueue
			client.updateQueue = append([]*base.PageRef{})

			client.bulkUpdate(queueCopy)
		}
		time.Sleep(1 * time.Second)
	}
}

func (client *ApiClient) bulkUpdate(buffer []*base.PageRef) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panicing update: %s", r)
		}
	}()

	log.Info("updating pageRefs (" + strconv.Itoa(len(buffer)) + ")")

	req := new(pb.PageRefList)

	req.List = buffer

	_, err := client.pageServiceClient.Complete(context.TODO(), req)

	log.Info("updating pageRefs done (" + strconv.Itoa(len(buffer)) + ")")

	lib.Check(err)
}

func (client *ApiClient) bulkInsert(buffer []*base.PageRef) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panicing insert: %s", r)
		}
	}()

	log.Info("inserting pageRefs (" + strconv.Itoa(len(buffer)) + ")")

	req := new(pb.PageRefList)

	req.List = buffer

	_, err := client.pageServiceClient.Create(context.TODO(), req)

	log.Info("inserting pageRefs end (" + strconv.Itoa(len(buffer)) + ")")

	lib.Check(err)
}
