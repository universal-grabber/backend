package helper

import (
	"backend/api/model"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"time"
)

const kafkaHost = "localhost:9092"

var (
	UgbKafkaInstance = new(UgbKafka)
)

type UgbKafka struct {
	writerMap map[string]*kafka.Writer
}

func (s *UgbKafka) getReader(topic string, group string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaHost},
		Topic:    topic,
		GroupID:  group,
		MinBytes: 10,   // 10KB
		MaxBytes: 10e6, // 10MB,
	})
}

func (s *UgbKafka) GetWriter(topic string) *kafka.Writer {
	if s.writerMap == nil {
		s.writerMap = make(map[string]*kafka.Writer)
	}
	if s.writerMap[topic] == nil {
		s.writerMap[topic] = &kafka.Writer{
			Addr:         kafka.TCP(kafkaHost),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 1 * time.Nanosecond,
		}
	}

	return s.writerMap[topic]
}

func (s *UgbKafka) SendPageRef(pageRef *model.PageRef) error {
	topic := locatePageRefTopic(pageRef)

	body, err := json.Marshal(pageRef)

	if err != nil {
		return err
	}

	err = s.GetWriter(topic).WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(pageRef.Id.String()),
		Value: body,
	})

	return err
}

func (s *UgbKafka) RecvPageRef(topic string, group string, interruptChan <-chan bool) <-chan *model.PageRef {
	pageChan := make(chan *model.PageRef)

	r := s.getReader(topic, group)

	go func() {
		defer func() {
			close(pageChan)
			r.Close()
		}()
		for {
			select {
			case <-interruptChan:
				log.Print("Stopping receiving items as client disconnected\n")
				return
			default:
			}
			msg, err := r.ReadMessage(context.Background())

			if err != nil {
				log.Error(err)
				return
			}

			pageRef := new(model.PageRef)

			err = json.Unmarshal(msg.Value, pageRef)

			if err != nil {
				log.Error(err)
				return
			}

			pageChan <- pageRef
		}
	}()

	return pageChan
}

func locatePageRefTopic(ref *model.PageRef) string {
	return "ug_" + ref.Data.Source + "_" + ref.Data.State + "_" + ref.Data.Status
}
