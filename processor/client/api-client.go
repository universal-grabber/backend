package client

import (
	"backend/processor/lib"
	"backend/processor/model"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"io"
	"net/http"
	"sync"
	"time"
)

type ApiClient struct {
	insertLock  sync.Mutex
	updateLock  sync.Mutex
	insertQueue []*model.PageRef
	updateQueue []*model.PageRef
	config      model.Config

	insertSemaphore *semaphore.Weighted
}

func (client *ApiClient) Init(config model.Config) {
	client.config = config
	go client.scheduleInsertChannel()
	go client.scheduleUpdateChannel()

	client.insertSemaphore = semaphore.NewWeighted(10000)
}

func (client *ApiClient) AcceptPages(taskName string) chan *model.PageRef {
	log.Debugf("starting to accept pages for task %s", taskName)
	pageRefReadChannel := make(chan *model.PageRef)

	go func() {
		var round uint64
		for {
			round++
			log.Tracef("starting to accept pages for task %s for round %d", taskName, round)

			client.RequestPageRefs(taskName, pageRefReadChannel)

			log.Tracef("finished to accept pages for task %s for round %d", taskName, round)
		}
	}()

	return pageRefReadChannel
}

func (client *ApiClient) RequestPageRefs(taskName string, pageRefChan chan *model.PageRef) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("panicing while request page refs for task: %s / %s", taskName, r)
		}
	}()

	url := client.config.UgbApiUri + "/api/1.0/page-refs/update-state?pageSize=100&fairSearch=true&state=" + taskName + "&status=PENDING&toState=" + taskName + "&toStatus=EXECUTING"

	if client.config.EnabledWebsites != nil {
		url = client.config.UgbApiUri + "/api/1.0/page-refs/update-state?pageSize=100&websiteName=" + client.config.EnabledWebsites[0] + "&state=" + taskName + "&status=PENDING&toState=" + taskName + "&toStatus=EXECUTING"
	}

	resp, err := http.Post(url, "application/text", nil)

	lib.Check(err)

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	client.read(reader, pageRefChan)
}

func (client *ApiClient) read(reader *bufio.Reader, pageRefChan chan *model.PageRef) {
	counter := 0
	for {
		line, _, err := reader.ReadLine()
		str := string(line)

		if err == io.EOF {
			break
		}

		if len(str) <= 1 {
			continue
		}

		if str[len(str)-1] == ',' {
			str = str[0 : len(str)-1]
			line = line[0 : len(line)-1]
		}

		pageRef := new(model.PageRef)

		err = json.Unmarshal(line, &pageRef)

		lib.Check(err)

		counter++
		pageRefChan <- pageRef
	}

	// wait for 1 seconds if no task received
	if counter == 0 {
		time.Sleep(1 * time.Second)
	}
}

func (client *ApiClient) InsertPageRef(ref *model.PageRef) {
	err := client.insertSemaphore.Acquire(context.TODO(), 1)

	if err != nil {
		log.Print(err)
	}

	client.insertLock.Lock()

	client.insertQueue = append(client.insertQueue, ref)

	client.insertLock.Unlock()
}

func (client *ApiClient) UpdatePageRef(ref *model.PageRef) {
	client.updateLock.Lock()

	client.updateQueue = append(client.updateQueue, ref)

	client.updateLock.Unlock()
}

func (client *ApiClient) scheduleInsertChannel() {
	for {
		if len(client.insertQueue) > 0 {
			queueCopy := client.insertQueue
			client.insertQueue = append([]*model.PageRef{})

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
			client.updateQueue = append([]*model.PageRef{})

			client.bulkUpdate(queueCopy)
		}
		time.Sleep(1 * time.Second)
	}
}

func (client *ApiClient) bulkUpdate(buffer []*model.PageRef) {
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

func (client *ApiClient) bulkInsert(buffer []*model.PageRef) {
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
