package main

import (
	"backend/api/helper"
	"context"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/atexit"
)

func main() {
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("ScheduleKafka")

	// to produce messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		GroupID:  "consumer-group-id2",
		Topic:    "ug_all-domains_DOWNLOAD_PENDING",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB,
	})

	atexit.Register(func() {
		if err := r.Close(); err != nil {
			log.Fatal("failed to close reader:", err)
		} else {
			log.Info("closed", err)
		}
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
