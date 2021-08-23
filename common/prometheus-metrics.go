package common

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func EnablePrometheusMetrics(service string) {
	go func() {
		http.Handle("/"+service+"/metrics", promhttp.Handler())
		log.Error(http.ListenAndServe(":1111", nil))
	}()
}
