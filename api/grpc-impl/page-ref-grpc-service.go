package grpc_impl

import (
	"backend/api/model"
	"backend/api/service"
	"backend/common"
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/api"
	"context"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

var (
	grpcMetricsRegistry = common.NewMeter("uga-grpc")
)

type PageRefGrpcService struct {
	pb.UnimplementedPageRefServiceServer

	service *service.PageRefKafkaService
}

func (receiver *PageRefGrpcService) Init() {
	receiver.service = new(service.PageRefKafkaService)
}

func (receiver PageRefGrpcService) Fetch(req *pb.PageRefFetchRequest, res pb.PageRefService_FetchServer) error {
	grpcMetricsRegistry.Inc("grpc-fetch-request", 1, req.State.String())

	interruptChan := make(chan bool)

	pageChan := receiver.service.Fetch(req.State, req.Websites, req, interruptChan)

	for record := range pageChan {
		err := res.Send(convertPageRef(record))
		grpcMetricsRegistry.Inc("grpc-fetch-send", 1, record.Data.State, record.Data.Status)

		if err != nil {
			log.Error(err)

			interruptChan <- false

			return err
		}
	}

	return nil
}

func (receiver PageRefGrpcService) Complete(_ context.Context, req *pb.PageRefList) (*base.Empty, error) {
	var items []model.PageRef

	grpcMetricsRegistry.Inc("grpc-complete-request", 1)

	for _, record := range req.List {
		items = append(items, *convertBasePageRef(record))
		grpcMetricsRegistry.Inc("grpc-complete-receive", 1, record.State.String(), record.Status.String())
	}

	err := receiver.service.Complete(items)

	if err != nil {
		log.Error(err)
	}

	return new(base.Empty), err
}

func (receiver PageRefGrpcService) Create(_ context.Context, req *pb.PageRefList) (*base.Empty, error) {
	var items []model.PageRef

	for _, record := range req.List {
		items = append(items, *convertBasePageRef(record))
	}

	receiver.service.BulkInsert(items)

	return new(base.Empty), nil
}

func convertPageRef(ref *model.PageRef) *base.PageRef {
	var tags []string

	if ref.Data.Tags != nil {
		tags = *ref.Data.Tags
	}

	return &base.PageRef{
		Id:          ref.Id.String(),
		WebsiteName: ref.Data.Source,
		Url:         ref.Data.Url,
		State:       base.PageRefState(base.PageRefState_value[ref.Data.State]),
		Status:      base.PageRefStatus(base.PageRefStatus_value[ref.Data.State]),
		Tags:        tags,
	}
}

func convertBasePageRef(record *base.PageRef) *model.PageRef {
	id, err := uuid.FromString(record.Id)

	common.Check(err)

	return &model.PageRef{
		Id: id,
		Data: model.PageRefData{
			Source: record.WebsiteName,
			Url:    record.Url,
			State:  record.State.String(),
			Status: record.Status.String(),
			Tags:   &record.Tags,
		},
	}
}
