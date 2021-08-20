package tasks

import (
	"backend/gen/proto/base"
	"backend/processor/client"
	"backend/processor/lib"
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
		State:           base.PageRefState_DOWNLOAD,
		Parallelism:     5000,
	}
}

func (task *DownloadTask) Run() {
	task.processor.Start()

	log.Print(task.Name(), " task started processing")

	task.processor.Wait()

	log.Print(task.Name(), " task stopped processing")
}

func (task *DownloadTask) process(item *base.PageRef) *base.PageRef {
	log.Tracef("page-ref received for download %s", item.Url)

	result := task.clients.GetBackendStorageClient().Store(item)

	if !result.Ok {
		item.Status = base.PageRefStatus_FAILED
	} else {
		item.Status = base.PageRefStatus_FINISHED
	}

	return item
}
