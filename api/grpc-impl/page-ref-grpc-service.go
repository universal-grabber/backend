package grpc_impl

import (
	"backend/api/helper"
	"backend/api/model"
	"backend/api/service"
	commonModel "backend/common/model"
	base "backend/gen/proto/base"
	pb "backend/gen/proto/service"
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
	log.Print("starting to send items")
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

func convertPageRef(ref *model.PageRef) *base.PageRef {
	return &base.PageRef{
		Id:          ref.Id.String(),
		WebsiteName: ref.WebsiteName,
		Url:         ref.Url,
		Tags:        *ref.Tags,
	}
}
