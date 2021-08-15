package lib

import (
	"backend/common"
	"backend/gen/proto/base"
	log "github.com/sirupsen/logrus"
	"runtime/debug"
	"sync"
)

type ProcessorFunc func(pageRef *base.PageRef) *base.PageRef

type Processor struct {
	ApiClient       ApiClientInterface
	TaskProcessFunc ProcessorFunc
	State           base.PageRefState
	Parallelism     int

	// private data
	started            bool
	wg                 *sync.WaitGroup
	pageRefReadChannel chan *base.PageRef
	timeCalc           *common.TimeCalc

	//progressingPagesMutex *sync.Mutex
	//progressingPages      map[string]bool
}

type ApiClientInterface interface {
	AcceptPages(state base.PageRefState) chan *base.PageRef
	UpdatePageRef(ref *base.PageRef)
}

func (processor *Processor) Start() {
	if processor.started {
		log.Panic("processor already started")
	}

	//processor.progressingPagesMutex = new(sync.Mutex)
	//processor.progressingPages = make(map[string]bool)

	processor.timeCalc = new(common.TimeCalc)
	processor.timeCalc.Init(processor.State.String())

	// init
	processor.started = true
	processor.wg = new(sync.WaitGroup)

	processor.pageRefReadChannel = processor.ApiClient.AcceptPages(processor.State)

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

func (processor *Processor) runProcess2(i int, pageRef *base.PageRef) {
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

func (processor *Processor) processItem(i int, pageRef *base.PageRef) {
	origPageRef := pageRef

	common.PageRefLogger(pageRef, "before-run-process-item").
		WithField("processIndex", i).
		Debug("sending item to process")

	pageRef = processor.runProcessItem(pageRef, i)

	processor.timeCalc.Step()

	if pageRef == nil {
		pageRef = origPageRef
		pageRef.Status = base.PageRefStatus_FAILED
	}

	// fix status
	if pageRef.Status != base.PageRefStatus_FINISHED {
		pageRef.Status = base.PageRefStatus_FAILED
	}

	common.PageRefLogger(pageRef, "after-run-process-item-status-fix").
		WithField("processIndex", i).
		Debug("process result for single item")

	processor.ApiClient.UpdatePageRef(pageRef)
}

func (processor *Processor) runProcessItem(pageRef *base.PageRef, processorIndex int) *base.PageRef {
	defer func() {
		if r := recover(); r != nil {
			common.PageRefLogger(pageRef, "run-process-panic").
				Errorf("panicing process[%d]: %s / %s / %s / %s / %s",
					processorIndex,
					processor.State,
					pageRef.Id,
					pageRef.Url,
					r,
					string(debug.Stack()))
		}
	}()
	return processor.TaskProcessFunc(pageRef)
}
