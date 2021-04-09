package lib

import (
	"backend/processor/model"
	log "github.com/sirupsen/logrus"
	"sync"
)

type ProcessorFunc func(pageRef *model.PageRef) *model.PageRef

type Processor struct {
	ApiClient       ApiClientInterface
	TaskProcessFunc ProcessorFunc
	TaskName        string
	Parallelism     int

	// private data
	started            bool
	wg                 *sync.WaitGroup
	pageRefReadChannel chan *model.PageRef
	timeCalc           *TimeCalc

	//progressingPagesMutex *sync.Mutex
	//progressingPages      map[string]bool
}

type ApiClientInterface interface {
	AcceptPages(taskName string) chan *model.PageRef
	UpdatePageRef(ref *model.PageRef)
}

func (processor *Processor) Start() {
	if processor.started {
		log.Panic("processor already started")
	}

	//processor.progressingPagesMutex = new(sync.Mutex)
	//processor.progressingPages = make(map[string]bool)

	processor.timeCalc = new(TimeCalc)
	processor.timeCalc.Init(processor.TaskName)

	// init
	processor.started = true
	processor.wg = new(sync.WaitGroup)

	processor.pageRefReadChannel = processor.ApiClient.AcceptPages(processor.TaskName)

	for i := 0; i < processor.Parallelism; i++ {
		processor.wg.Add(1)
		go processor.runProcess(i)
	}
}

func (processor *Processor) Wait() {
	processor.wg.Wait()
}

func (processor *Processor) runProcess(i int) {
	for pageRef := range processor.pageRefReadChannel {
		processor.runProcess2(i, pageRef)
	}
}

func (processor *Processor) runProcess2(i int, pageRef *model.PageRef) {
	//processor.progressingPagesMutex.Lock()
	//if processor.progressingPages[pageRef.Id.String()] {
	//	processor.progressingPagesMutex.Unlock()
	//	PageRefLogger(pageRef, "already-processing").
	//		Debug("item is already processing, sending back")
	//	return
	//}
	//processor.progressingPagesMutex.Unlock()
	//
	//processor.addItemToUniqueSet(pageRef.Id)
	//defer func() {
	//	processor.removeItemFromUniqueSet(pageRef.Id)
	//}()
	processor.processItem(i, pageRef)
}

//func (processor *Processor) addItemToUniqueSet(id uuid.UUID) {
//	defer processor.progressingPagesMutex.Unlock()
//	processor.progressingPagesMutex.Lock()
//
//	processor.progressingPages[id.String()] = true
//}
//
//func (processor *Processor) removeItemFromUniqueSet(id uuid.UUID) {
//	defer processor.progressingPagesMutex.Unlock()
//	processor.progressingPagesMutex.Lock()
//
//	delete(processor.progressingPages, id.String())
//}

func (processor *Processor) processItem(i int, pageRef *model.PageRef) {
	origPageRef := pageRef

	PageRefLogger(pageRef, "before-run-process-item").
		Debug("sending item to process")

	pageRef = processor.runProcessItem(pageRef, i)

	processor.timeCalc.Step()

	if pageRef == nil {
		pageRef = origPageRef
		pageRef.Status = "FAILED"
	}

	// fix status
	if pageRef.Status != "FINISHED" {
		pageRef.Status = "FAILED"
	}

	PageRefLogger(pageRef, "after-run-process-item-status-fix").
		Debug("process result for single item")

	processor.ApiClient.UpdatePageRef(pageRef)
}

func (processor *Processor) runProcessItem(pageRef *model.PageRef, processorIndex int) *model.PageRef {
	defer func() {
		if r := recover(); r != nil {
			PageRefLogger(pageRef, "run-process-panic").
				Errorf("panicing process[%d]: %s / %s / %s / %s",
					processorIndex,
					processor.TaskName,
					pageRef.Id,
					pageRef.Url,
					r)
		}
	}()
	return processor.TaskProcessFunc(pageRef)
}
