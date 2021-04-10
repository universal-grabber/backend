package main

import (
	appPackage "backend/processor/app"
	"backend/processor/model"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)

	app := new(appPackage.App)

	configure(app)

	app.Init()

	app.Run()
}

func configure(app *appPackage.App) {
	config := model.Config{}

	//debugStr := getStringConfig("LOG_LEVEL", strconv.Itoa(int(log.DebugLevel)))
	enabledWebsites := getStringConfig("ENABLED_WEBSITES", "")
	enabledTasks := getStringConfig("ENABLED_TASKS", "")
	storageApi := getStringConfig("STORAGE_API", "")
	modelProcessorApi := getStringConfig("MODEL_PROCESSOR_API", "")
	backendApi := getStringConfig("BACKEND_API", "")
	backendGrpcApi := getStringConfig("BACKEND_GRPC_API", "")
	parseMongoUri := getStringConfig("PARSE_MONGO_URI", "")

	//level, err := strconv.Atoi(debugStr)

	//lib.Check(err)

	config.LogLevel = log.TraceLevel

	config.UgbStorageUri = storageApi
	config.UgbModelProcessorUri = modelProcessorApi
	config.UgbApiUri = backendApi
	config.UgbApiGrpcUri = backendGrpcApi
	config.ParseMongoUri = parseMongoUri

	if enabledWebsites != "" {
		config.EnabledWebsites = strings.Split(enabledWebsites, ",")
	}

	if enabledTasks != "" {
		config.EnabledTasks = strings.Split(enabledTasks, ",")
	}

	app.Configure(config)
}

func getStringConfig(name string, defaultValue string) string {
	envVal := os.Getenv(name)

	if envVal != "" {
		return envVal
	}

	return defaultValue
}
