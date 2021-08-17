package helper

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type TimeCalc struct {
	name    string
	start   time.Time
	current time.Time

	lastCount uint64
	counter   uint64
	_log      *log.Entry
}

func (obj *TimeCalc) Init(name string) {
	obj.name = name

	obj.start = time.Now()
	obj.current = time.Now()

	obj.lastCount = 0
	obj.counter = 0

	obj._log = log.WithField("component", "time-calc")
}

func (obj *TimeCalc) Logger(logEntry *log.Entry) {
	obj._log = logEntry
}

func (obj *TimeCalc) Step() {
	obj.StepWithExceedMillis(3000)
}

func (obj *TimeCalc) StepWithExceedMillis(exceedMillis int64) {
	obj.counter++

	diff := obj.counter - obj.lastCount
	diffTime := (time.Now().UnixNano() - obj.current.UnixNano()) / 1000000
	diffTimeFromStart := (time.Now().UnixNano() - obj.start.UnixNano()) / 1000000

	if diffTime > exceedMillis {
		obj.current = time.Now()
		obj.lastCount = obj.counter

		lastSpeed := float32(diff) * 1000 / float32(diffTime)
		speed := float32(obj.counter) * 1000 / float32(diffTimeFromStart)

		obj._log.Printf("%s: %.2f ops, %.2f aops %d \n", obj.name, lastSpeed, speed, obj.counter)
	}
}
