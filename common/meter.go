package common

import (
	"bytes"
	"fmt"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
	"time"
)

type Meter interface {
	Inc(operation string, count int64, tags map[string]string)
}

type meter struct {
	counters map[string]metrics.Counter
	name     string
}

func (m *meter) Init(name string) {
	m.counters = make(map[string]metrics.Counter)
	m.name = name
}

func (m *meter) Inc(operation string, count int64, tags map[string]string) {
	key := operation + ":" + createKeyValuePairs(tags)

	if m.counters[key] == nil {
		c := metrics.NewCounter()

		registry := metrics.NewRegistry()

		if tags == nil {
			tags = make(map[string]string)
		}

		tags["operation"] = operation

		go influxdb.InfluxDBWithTags(registry,
			time.Second,
			"http://ug.tisserv.net:8086",
			"ug",
			m.name,
			"",
			"",
			tags,
			true,
		)

		err := registry.Register("value", c)

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

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}
