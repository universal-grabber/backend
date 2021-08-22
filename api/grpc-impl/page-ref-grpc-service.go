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
	ctx, cancel := context.WithCancel(res.Context())

	ctx = common.WithLogger(ctx)
	ctx = common.WithMeter(ctx, grpcMetricsRegistry)

	common.UseMeter(ctx).Inc("grpc-fetch-request", 1, map[string]string{
		"state": req.State.String(),
	})

	pageChan := receiver.service.Fetch(ctx, req.State, req.Websites)

	for record := range pageChan {
		err := res.Send(convertPageRef(record))
		common.UseMeter(ctx).Inc("grpc-fetch-send", 1, common.PageRefRecordToTags(*record))

		if err != nil {
			log.Error(err)

			cancel()
			return err
		}
	}

	// cancel operation after we finished
	cancel()

	return nil
}

func (receiver PageRefGrpcService) Complete(_ context.Context, req *pb.PageRefList) (*base.Empty, error) {
	var items []model.PageRef

	grpcMetricsRegistry.Inc("grpc-complete-request", 1, nil)

	for _, record := range req.List {
		items = append(items, *convertBasePageRef(record))
		grpcMetricsRegistry.Inc("grpc-complete-receive", 1, common.PageRefRecordToTags2(*record))
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
