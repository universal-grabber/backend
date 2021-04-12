package client

import (
	"backend/gen/proto/base"
	pb "backend/gen/proto/service"
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

type ApiClientNew struct {
	insertLock  sync.Mutex
	updateLock  sync.Mutex
	insertQueue []*base.PageRef
	updateQueue []*base.PageRef
	config      model.Config

	insertSemaphore *semaphore.Weighted

	pageServiceClient pb.PageRefServiceClient
	connection        *grpc.ClientConn
}

func (client *ApiClientNew) Init(config model.Config) {
	client.config = config
	go client.scheduleInsertChannel()
	go client.scheduleUpdateChannel()

	client.insertSemaphore = semaphore.NewWeighted(10000)

	// initialize grpc
	// Set up a connection to the server.
	conn, err := grpc.Dial(config.UgbApiGrpcUri, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client.connection = conn

	client.pageServiceClient = pb.NewPageRefServiceClient(conn)
}

func (client *ApiClientNew) AcceptPages(state base.PageRefState) chan *base.PageRef {
	timeCalc := new(lib.TimeCalc)
	timeCalc.Init("timer1")

	log.Debugf("starting to accept pages for state %s", state)

	req := new(pb.PageRefServiceUpdateRequest)
	req.FairSearch = true
	req.State = state
	req.Status = base.PageRefStatus_PENDING
	req.ToState = state
	req.ToStatus = base.PageRefStatus_EXECUTING
	req.EnabledWebsites = client.config.EnabledWebsites

	pageRefReadChannel := make(chan *base.PageRef)

	go func() {
		for {
			resp, err := client.pageServiceClient.UpdateAndAccept(context.TODO(), req)

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

			time.Sleep(1 * time.Second)
		}
	}()

	return pageRefReadChannel
}

func (client *ApiClientNew) InsertPageRef(ref *base.PageRef) {
	err := client.insertSemaphore.Acquire(context.TODO(), 1)

	if err != nil {
		log.Print(err)
	}

	client.insertLock.Lock()

	client.insertQueue = append(client.insertQueue, ref)

	client.insertLock.Unlock()
}

func (client *ApiClientNew) UpdatePageRef(ref *base.PageRef) {
	client.updateLock.Lock()

	client.updateQueue = append(client.updateQueue, ref)

	client.updateLock.Unlock()
}

func (client *ApiClientNew) scheduleInsertChannel() {
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

func (client *ApiClientNew) scheduleUpdateChannel() {
	for {
		if len(client.updateQueue) > 0 {
			queueCopy := client.updateQueue
			client.updateQueue = append([]*base.PageRef{})

			client.bulkUpdate(queueCopy)
		}
		time.Sleep(1 * time.Second)
	}
}

func (client *ApiClientNew) bulkUpdate(buffer []*base.PageRef) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panicing update: %s", r)
		}
	}()

	log.Info("updating pageRefs (" + strconv.Itoa(len(buffer)) + ")")

	req := new(pb.PageRefList)

	req.List = buffer

	_, err := client.pageServiceClient.Update(context.TODO(), req)

	log.Info("updating pageRefs done (" + strconv.Itoa(len(buffer)) + ")")

	lib.Check(err)
}

func (client *ApiClientNew) bulkInsert(buffer []*base.PageRef) {
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
