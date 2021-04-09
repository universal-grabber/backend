package tasks

import (
	"backend/processor/client"
	"backend/processor/lib"
	"backend/processor/model"
	log "github.com/sirupsen/logrus"
)

type DownloadTask struct {
	clients   client.Clients
	processor lib.Processor
}

func (task *DownloadTask) Name() string {
	return "DOWNLOAD"
}

func (task *DownloadTask) Init(clients client.Clients) {
	task.clients = clients

	task.processor = lib.Processor{
		ApiClient:       task.clients.GetApiClient(),
		TaskProcessFunc: task.process,
		TaskName:        task.Name(),
		Parallelism:     150,
	}
}

func (task *DownloadTask) Run() {
	task.processor.Start()

	log.Print(task.Name(), " task started processing")

	task.processor.Wait()

	log.Print(task.Name(), " task stopped processing")
}

func (task *DownloadTask) process(item *model.PageRef) *model.PageRef {
	log.Tracef("page-ref received for download %s", item.Url)

	result := task.clients.GetBackendStorageClient().Store(item)

	if !result.Ok {
		item.Status = "FAILED"
	} else {
		item.Status = "FINISHED"
	}

	return item
}
