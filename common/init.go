package common

func Init(service string) {
	EnableGrayLog(service)
	EnablePrometheusMetrics(service)
}
