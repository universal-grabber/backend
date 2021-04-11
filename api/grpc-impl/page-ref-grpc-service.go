package grpc_impl

import (
	"backend/api/helper"
	"backend/api/model"
	"backend/api/service"
	commonModel "backend/common/model"
	"backend/gen/proto/base"
	pb "backend/gen/proto/service"
	"backend/storage/lib"
	"context"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type PageRefGrpcService struct {
	pb.UnimplementedPageRefServiceServer

	service *service.PageRefService
}

func (receiver *PageRefGrpcService) Init() {
	receiver.service = new(service.PageRefService)
}

func (receiver PageRefGrpcService) UpdateAndAccept(req *pb.PageRefServiceUpdateRequest, res pb.PageRefService_UpdateAndAcceptServer) error {
	log.Print("requested UpdateAndAccept for " + req.String())
	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	if len(req.GetToState().String()) == 0 {
		return commonModel.NewException("toState is missing")
	}

	if len(req.GetToStatus().String()) == 0 {
		return commonModel.NewException("toStatus is missing")
	}

	searchPageRef := new(model.SearchPageRef)

	searchPageRef.FairSearch = req.FairSearch
	searchPageRef.PageSize = 1000
	searchPageRef.State = req.State.String()
	searchPageRef.Status = req.Status.String()
	if req.EnabledWebsites != nil && len(req.EnabledWebsites) > 0 {
		searchPageRef.WebsiteName = req.EnabledWebsites[0]
	}
	// implement logic for websites and tags

	var closeChan = make(chan bool)

	go func() {
		<-res.Context().Done()
		closeChan <- true
	}()

	pageChan, updateChan := receiver.service.UpdateStatesBulk2(searchPageRef, req.GetToState().String(), req.GetToStatus().String(), closeChan)

	defer close(updateChan)

	for pageRef := range pageChan {
		helper.PageRefLogger(pageRef, "request-update-state").Debug("pageRef state updated")

		err := res.Send(convertPageRef(pageRef))

		if err != nil {
			log.Warn(err)
			break
		}

		updateChan <- pageRef
	}

	return nil
}

func (receiver PageRefGrpcService) Update(_ context.Context, req *pb.PageRefList) (*base.Empty, error) {
	var items []model.PageRef

	for _, record := range req.List {
		items = append(items, convertBasePageRef(record))
	}

	receiver.service.BulkWrite2(items)

	return nil, nil
}

func (receiver PageRefGrpcService) Create(_ context.Context, req *pb.PageRefList) (*base.Empty, error) {
	var items []model.PageRef

	for _, record := range req.List {
		items = append(items, convertBasePageRef(record))
	}

	receiver.service.BulkInsert(items)

	return nil, nil
}

func convertPageRef(ref *model.PageRef) *base.PageRef {
	var tags []string

	if ref.Tags != nil {
		tags = *ref.Tags
	}

	return &base.PageRef{
		Id:          ref.Id.String(),
		WebsiteName: ref.WebsiteName,
		Url:         ref.Url,
		Tags:        tags,
	}
}

func convertBasePageRef(record *base.PageRef) model.PageRef {
	id, err := uuid.FromString(record.Id)

	lib.Check(err)

	return model.PageRef{
		Id:          id,
		WebsiteName: record.WebsiteName,
		Url:         record.Url,
		State:       record.State.String(),
		Status:      record.Status.String(),
		Tags:        &record.Tags,
	}
}
