package helper

import (
	"backend/api/model"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"time"
)

const kafkaHost = "ug.tisserv.net:9092"

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
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		//Logger:   log.WithField("topic", topic).WithField("group", group),
	})
}

func (s *UgbKafka) ListTopics() []string {
	conn, err := kafka.Dial("tcp", kafkaHost)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		panic(err.Error())
	}

	m := map[string]struct{}{}

	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}

	var list []string

	for k := range m {
		list = append(list, k)
	}

	return list
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
			BatchTimeout: 1 * time.Second,
			//Logger:       log.WithField("topic", topic),
			Async: false,
		}
	}

	return s.writerMap[topic]
}

func (s *UgbKafka) SendPageRef(list []model.PageRef) error {
	var messagesMap = make(map[string][]kafka.Message)

	for _, pageRef := range list {
		if pageRef.Data.Status != "PENDING" {
			log.Tracef("not sending item to kafka as it is not in pending state %s / %s", pageRef.Id.String(), pageRef.Data.Status)
			continue
		}

		topic := s.LocatePageRefTopic(pageRef)
		body, err := json.Marshal(pageRef)

		if err != nil {
			return err
		}

		message := kafka.Message{
			Key:   []byte(pageRef.Id.String()),
			Value: body,
		}

		messagesMap[topic] = append(messagesMap[topic], message)
	}

	for topic, list := range messagesMap {
		log.Infof("begin sending %d items to kafka on topic %s", len(list), topic)

		err := s.GetWriter(topic).WriteMessages(context.Background(), list...)

		log.Infof("finish sending %d items to kafka on topic %s", len(list), topic)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UgbKafka) RecvPageRef(context context.Context, topic string, group string) <-chan *model.PageRef {
	pageChan := make(chan *model.PageRef, 1000)

	r := s.getReader(topic, group)

	go func() {
		defer func() {
			close(pageChan)
			r.Close()
		}()
		for {
			msg, err := r.ReadMessage(context)

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

func (s *UgbKafka) GetConsumerGroupStats(groupName string, topics []string) map[string][]kafka.PartitionAssignment {
	conn, err := kafka.Dial("tcp", kafkaHost)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	group, err := kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		ID:      groupName,
		Brokers: []string{kafkaHost},
		Topics:  topics,
	})

	gen, err := group.Next(context.TODO())

	return gen.Assignments
}

func (s *UgbKafka) LocatePageRefTopic(ref model.PageRef) string {
	return "ug_" + ref.Data.Source + "_" + ref.Data.State
}

func (s *UgbKafka) ProvisionTopic(topic string) error {
	conn, err := kafka.Dial("tcp", kafkaHost)
	if err != nil {
		return err
	}

	defer conn.Close()

	log.Infof("provisioning / deleing topic: %s", topic)
	err = conn.DeleteTopics(topic)

	log.Infof("provisioning / creating topic: %s", topic)
	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     8,
		ReplicationFactor: 1,
	})
}
