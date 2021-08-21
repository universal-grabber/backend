package main

import (
	"backend/common"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	grpcMetricsRegistry := common.NewMeter("ugaTest")

	grpcMetricsRegistry.Inc("op-1", 1)
	grpcMetricsRegistry.Inc("op-1", 1)
	grpcMetricsRegistry.Inc("op-1", 1)

	grpcMetricsRegistry.Inc("op-1", 1, "tag-1")
	grpcMetricsRegistry.Inc("op-1", 1, "tag-2")
	grpcMetricsRegistry.Inc("op-1", 1, "tag-1")

	time.Sleep(2 * time.Second)
}

func runMethodOnSpeed(c metrics.Counter, executionPerSecond int) {
	sleepTime := time.Duration(time.Second.Nanoseconds()/int64(executionPerSecond)) * 10

	log.Print("sleepTime", sleepTime)

	for i := 0; i < 10; i++ {
		go func() {
			for {
				time.Sleep(sleepTime)

				c.Inc(1)
			}
		}()
	}

}
