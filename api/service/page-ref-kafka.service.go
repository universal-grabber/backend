package service

import (
	"backend/api/helper"
	"backend/api/model"
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/api"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type PageRefKafkaService struct {
	pageRefService *PageRefService
}

func (service *PageRefKafkaService) Init() {
	service.pageRefService = new(PageRefService)
}

func (service *PageRefKafkaService) BulkWrite2(list []model.PageRef) {

}

func (service *PageRefKafkaService) BulkInsert(list []*model.PageRef) {

}

func (service *PageRefKafkaService) Fetch(state base.PageRefState, websites []string, req *pb.PageRefFetchRequest, interruptChan chan bool) chan *model.PageRef {
	log.Debug("Fetch requested")
	pageChan := make(chan *model.PageRef)
	kafka := helper.UgbKafkaInstance

	var interruptions []chan bool

	go func() {
		counter := 0

		log.Debug("Request topics from kafka")
		topics := kafka.ListTopics()
		log.Debug("Topic list from kafka: {}", topics)

		wg := new(sync.WaitGroup)

		for _, topic := range topics {
			if !strings.Contains(topic, state.String()) {
				log.Print("topic {} ignored for state {}", topic, state)
				continue
			}

			wg.Add(1)

			localInterruptChan := make(chan bool)
			interruptions = append(interruptions, localInterruptChan)

			log.Debug("requesting kafka to fetch with topic: {}", topic)
			localPageChan := kafka.RecvPageRef(topic, "FetchGroup", localInterruptChan)
			log.Debug("request accepted by kafka to fetch with topic: {}", topic)

			topic := topic
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Panic: %+v\n", r)
					}
				}()

			MainLoop:
				for {
					select {
					case pageRef, ok := <-localPageChan:
						if !ok {
							interruptChan <- false
							log.Print("localPageChan not ok: {}", topic)
							break MainLoop
						}
						pageChan <- pageRef
						log.Print("accepted item: {}", pageRef)
					case <-time.After(3 * time.Second):
						interruptChan <- false
						log.Print("timeout on topic: {}", topic)
						break MainLoop
					}
					counter++

					if counter == 10000 {
						interruptChan <- false
						log.Debug("interrupt signal sent after max counter reached")
					}
				}
				log.Debug("request finished to fetch with topic: {}", topic)
				wg.Done()
			}()
		}

		wg.Wait()
		interruptChan <- false
		log.Debug("interrupt signal sent after wg done")
	}()

	go func() {
		log.Debug("interrupt signal received")
		<-interruptChan

		for _, interrupt := range interruptions {
			interrupt <- true
		}

		close(pageChan)
	}()

	return pageChan
}

func (service *PageRefKafkaService) Complete(list []*model.PageRef) error {
	return nil
}
