package main
//
//import (
//	"github.com/rcrowley/go-metrics"
//	log "github.com/sirupsen/logrus"
//	"time"
//)
//
//func main() {
//	//timer := metrics.NewTimer()
//
//}
//
//func runMethodOnSpeed(c metrics.Counter, executionPerSecond int) {
//	sleepTime := time.Duration(time.Second.Nanoseconds()/int64(executionPerSecond)) * 10
//
//	log.Print("sleepTime", sleepTime)
//
//	for i := 0; i < 10; i++ {
//		go func() {
//			for {
//				time.Sleep(sleepTime)
//
//				c.Inc(1)
//			}
//		}()
//	}
//
//}
