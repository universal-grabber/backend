package client

import (
	"backend/gen/proto/base"
	pb "backend/gen/proto/service"
	"backend/processor/lib"
	"backend/processor/model"
	"bytes"
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"net/http"
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

	for {
		resp, err := client.pageServiceClient.UpdateAndAccept(context.TODO(), req)

		if err != nil {
			log.Warn(err)
			time.Sleep(1 * time.Second)
			continue
		}

		for {
			item, err := resp.Recv()

			if err.Error() == "EOF" {
				break
			}

			if err != nil {
				log.Warn(err)
				break
			}

			timeCalc.Step()
			pageRefReadChannel <- item
		}

		time.Sleep(1 * time.Second)
	}

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
	//log.Printf("starting update flush %d items", len(buffer))
	body, err := json.Marshal(buffer)

	if err != nil {
		log.Panic(err)
	}

	req, err := http.NewRequest("PATCH", client.config.UgbApiUri+"/api/1.0/page-refs/bulk", bytes.NewReader(body))

	if err != nil {
		log.Panic(err)
	}

	_, err = http.DefaultClient.Do(req)

	if err != nil {
		log.Panic(err)
	}
	//log.Printf("finished update flush %d items", len(buffer))
}

func (client *ApiClientNew) bulkInsert(buffer []*base.PageRef) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panicing insert: %s", r)
		}
	}()

	//log.Printf("starting insert flush %d items", len(buffer))
	body, err := json.Marshal(buffer)

	if err != nil {
		log.Panic(err)
	}

	req, err := http.NewRequest("POST", client.config.UgbApiUri+"/api/1.0/page-refs/bulk", bytes.NewReader(body))

	if err != nil {
		log.Panic(err)
	}

	_, err = http.DefaultClient.Do(req)

	if err != nil {
		log.Panic(err)
	}
	//log.Printf("finished insert flush %d items", len(buffer))
}
