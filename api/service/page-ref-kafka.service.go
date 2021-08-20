package service

import (
	"backend/api/helper"
	"backend/api/model"
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/api"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
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
	requestId := rand.Intn(1000000)

	log.WithField("requestId", requestId).Debug("Fetch requested")
	pageChan := make(chan *model.PageRef)
	kafka := helper.UgbKafkaInstance

	var interruptions []chan bool

	go func() {
		counter := 0

		log.WithField("requestId", requestId).Debug("Request topics from kafka")
		topics := kafka.ListTopics()
		log.WithField("requestId", requestId).Debug("Topic list from kafka: {}", topics)

		wg := new(sync.WaitGroup)

		for _, topic := range topics {
			if !strings.Contains(topic, state.String()) {
				log.WithField("requestId", requestId).Debug("topic {} ignored for state {}", topic, state)
				continue
			}

			log.WithField("requestId", requestId).Debug("increase wg")
			wg.Add(1)

			localInterruptChan := make(chan bool)
			interruptions = append(interruptions, localInterruptChan)

			log.WithField("requestId", requestId).Debug("requesting kafka to fetch with topic: {}", topic)
			localPageChan := kafka.RecvPageRef(topic, "FetchGroup", localInterruptChan)
			log.WithField("requestId", requestId).Debug("request accepted by kafka to fetch with topic: {}", topic)

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
							log.WithField("requestId", requestId).Print("localPageChan not ok: {}", topic)
							break MainLoop
						}
						counter++
						pageChan <- pageRef
						log.WithField("requestId", requestId).Tracef("accepted item: %s", pageRef.Id)
					case <-time.After(60 * time.Second):
						interruptChan <- false
						log.WithField("requestId", requestId).Infof("timeout on topic: %s", topic)
						break MainLoop
					}

					if counter == 100000 {
						interruptChan <- false
						log.WithField("requestId", requestId).Debug("interrupt signal sent after max counter reached")
					}
				}
				log.WithField("requestId", requestId).Debug("request finished to fetch with topic: {}", topic)
				wg.Done()
			}()
		}

		wg.Wait()
		interruptChan <- false
		log.WithField("requestId", requestId).Debug("interrupt signal sent after wg done")
		log.WithField("requestId", requestId).Debug("accepted items from kafka and send to processor: ", counter)
	}()

	go func() {
		<-interruptChan
		log.WithField("requestId", requestId).Debug("interrupt signal received")

		for _, interrupt := range interruptions {
			interrupt <- true
		}

		close(pageChan)
	}()

	return pageChan
}

func (service *PageRefKafkaService) Complete(list []model.PageRef) error {
	service.pageRefService.BulkWrite2(list)

	return nil
}
