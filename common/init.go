package common

func Init(service string) {
	EnablePrometheusMetrics(service)
	EnableGrayLog(service)
}
