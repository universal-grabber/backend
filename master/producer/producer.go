package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func main() {
	// to produce messages

	w := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9092"),
		Topic:        "test",
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 1 * time.Millisecond,
	}

	i := 0
	for {
		i++
		log.Info("begin: " + strconv.Itoa(i))
		err := w.WriteMessages(context.TODO(),
			kafka.Message{
				Key:   []byte("Key-A" + strconv.Itoa(i)),
				Value: []byte("Hello World!" + strconv.Itoa(i)),
			},
		)

		time.Sleep(10 * time.Millisecond)
		log.Info("end: " + strconv.Itoa(i))

		if err != nil {
			log.Warn("failed to write messages:", err)
		}
	}

	if err := w.Close(); err != nil {
		log.Warn("failed to close writer:", err)
	}

}
