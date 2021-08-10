package main

import (
	"backend/api/helper"
	"context"
	"github.com/segmentio/kafka-go"
)

func main() {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("ScheduleKafka")

	// to produce messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"ug.tisserv.net:9092"},
		GroupID:  "consumer-group-id2",
		Topic:    "ug_all-domains_DOWNLOAD_PENDING",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB,
	})

	for {
		_, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		//log.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		timeCalc.Step()
	}

}
