package app

import (
	"backend/processor/client"
	"backend/processor/lib"
	"backend/processor/model"
	"backend/processor/tasks"
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

type App struct {
	apiClient            *client.ApiClientNew
	backendStorageClient *client.BackendStorageClient
	modelProcessorClient *client.ModelProcessorClient
	publisherClient      *client.PublisherClient

	// private data
	tasks  []tasks.BaseTask
	config model.Config
}

func (app *App) Init() {

	app.initClients()

	app.initTasks()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (app *App) Run() {
	wg := new(sync.WaitGroup)
	wg.Add(len(app.tasks))

	for _, task := range app.tasks {
		if app.config.EnabledTasks != nil && !lib.Contains(app.config.EnabledTasks, task.Name()) {
			continue
		}

		theTask := task

		go func() {
			log.Printf("starting task: %s", theTask.Name())

			defer func() {
				if r := recover(); r != nil {
					log.Printf("panicing task: %s / %s", theTask.Name(), r)
				}
			}()

			theTask.Run()

			log.Printf("task finished: %s", theTask.Name())
			wg.Done()
		}()
	}

	wg.Wait()
}

func (app *App) initClients() {
	app.apiClient = new(client.ApiClientNew)
	app.backendStorageClient = new(client.BackendStorageClient)
	app.modelProcessorClient = new(client.ModelProcessorClient)
	app.publisherClient = new(client.PublisherClient)

	app.apiClient.Init(app.GetConfig())
	app.backendStorageClient.Init(app.GetConfig())
	app.modelProcessorClient.Init(app.GetConfig())
	app.publisherClient.Init(app.GetConfig())
}

func (app *App) initTasks() {
	// load all tasks
	app.tasks = append([]tasks.BaseTask{},
		new(tasks.DownloadTask),
		new(tasks.DeepScanTask),
		new(tasks.ParseTask),
		new(tasks.PublishTask),
	)

	// init all tasks
	for _, task := range app.tasks {
		if app.config.EnabledTasks != nil && !lib.Contains(app.config.EnabledTasks, task.Name()) {
			continue
		}

		task.Init(app)
	}
}

func (app *App) GetApiClient() *client.ApiClientNew {
	return app.apiClient
}
func (app *App) GetBackendStorageClient() *client.BackendStorageClient {
	return app.backendStorageClient
}
func (app *App) GetModelProcessorClient() *client.ModelProcessorClient {
	return app.modelProcessorClient
}
func (app *App) GetPublisherClient() *client.PublisherClient {
	return app.publisherClient
}

func (app *App) Configure(config model.Config) {
	app.config = config
}

func (app *App) GetConfig() model.Config {
	return app.config
}
