package service

import (
	"backend/api/helper"
	"backend/api/model"
	"backend/common"
	"backend/gen/proto/base"
	"context"
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

func (service *PageRefKafkaService) Fetch(ctx context.Context, state base.PageRefState, websites []string) chan *model.PageRef {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	common.UseLogger(ctx).Debug("Fetch requested")
	pageChan := make(chan *model.PageRef)
	kafka := helper.UgbKafkaInstance

	go func() {
		counter := 0

		defer func() {
			common.UseLogger(ctx).Debug("closing pageChan")
			close(pageChan)
		}()

		common.UseLogger(ctx).Debug("Request topics from kafka")
		topics := kafka.ListTopics()
		common.UseLogger(ctx).Debug("Topic list from kafka: {}", topics)

		wg := new(sync.WaitGroup)

		for _, topic := range topics {
			if !strings.Contains(topic, state.String()) {
				common.UseLogger(ctx).Debug("topic {} ignored for state {}", topic, state)
				continue
			}

			common.UseLogger(ctx).Debug("increase wg")
			wg.Add(1)

			common.UseLogger(ctx).Debug("requesting kafka to fetch with topic: {}", topic)
			localPageChan := kafka.RecvPageRef(ctx, topic, "FetchGroup")
			common.UseLogger(ctx).Debug("request accepted by kafka to fetch with topic: {}", topic)

			localTopic := topic
			go func() {
				defer func() {
					if r := recover(); r != nil {
						common.UseLogger(ctx).Fatal("Panic: %+v\n", r)
					}
				}()

			MainLoop:
				for {
					select {
					case pageRef, ok := <-localPageChan:
						if !ok {
							cancel()
							common.UseLogger(ctx).Print("localPageChan not ok: {}", localTopic)
							break MainLoop
						}
						counter++
						pageChan <- pageRef
						common.UseLogger(ctx).Tracef("accepted item: %s", pageRef.Id)
					case <-time.After(60 * time.Second):
						cancel()
						common.UseLogger(ctx).Infof("timeout on topic: %s", localTopic)
						break MainLoop
					}

					if counter == 10000 {
						cancel()
						common.UseLogger(ctx).Debug("cancel signal sent after max counter reached")
					}
				}
				common.UseLogger(ctx).Debug("request finished to fetch with topic: {}", localTopic)
				wg.Done()
			}()
		}

		wg.Wait()
		common.UseLogger(ctx).Debug("accepted items from kafka and send to processor: ", counter)
	}()

	return pageChan
}

func (service *PageRefKafkaService) Complete(list []model.PageRef) error {
	kafka := helper.UgbKafkaInstance

	updated := service.pageRefService.BulkWrite(list)

	err := kafka.SendPageRef(updated)

	if err != nil {
		log.Error(err)
	}

	return err
}

func (service *PageRefKafkaService) BulkInsert(list []model.PageRef) {
	service.pageRefService.BulkInsert(list)
}
