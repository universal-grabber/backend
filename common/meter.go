package common

import (
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
	"strings"
	"time"
)

type Meter interface {
	Inc(operation string, count int64, tags ...string)
}

type meter struct {
	counters map[string]metrics.Counter
	registry metrics.Registry
}

func (m *meter) Init(name string) {
	m.counters = make(map[string]metrics.Counter)
	m.registry = metrics.NewRegistry()

	go influxdb.InfluxDB(m.registry,
		time.Second,
		"http://ug.tisserv.net:8086",
		"ug",
		name,
		"",
		"",
		true,
	)
}

func (m *meter) Inc(operation string, count int64, tags ...string) {
	key := operation + ":" + strings.Join(tags, ":")

	if m.counters[key] == nil {
		c := metrics.NewCounter()

		err := m.registry.Register(key, c)

		if err != nil {
			log.Error(err)
		}

		m.counters[key] = c
	}

	m.counters[key].Inc(count)
}

func NewMeter(name string) Meter {
	m := new(meter)

	m.Init(name)

	return m
}

func RegisterMetric(r metrics.Registry, name string) {
	go influxdb.InfluxDB(r,
		time.Second,
		"http://ug.tisserv.net:8086",
		"ug",
		name,
		"",
		"",
		true,
	)
}

func CreateAndRegisterMetric(name string) metrics.Registry {
	r := metrics.NewRegistry()

	RegisterMetric(r, name)

	return r
}

func CreateCounterMetric(registry metrics.Registry, name string) metrics.Counter {
	c := metrics.NewCounter()

	err := registry.Register(name, c)

	if err != nil {
		log.Error(err)
	}

	return c
}
