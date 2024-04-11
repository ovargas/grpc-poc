package dummy

import (
	"context"
	"fmt"
	dummyv1 "grpc-poc/api/dummy/v1"
	"grpc-poc/cmd/system"
)

type (
	Service struct {
		dummyv1.UnimplementedDummyServiceServer
	}

	dummy struct{}
)

func New() *Service {
	return &Service{}
}

func (s *Service) GetDummy(ctx context.Context, r *dummyv1.GetDummyRequest) (*dummyv1.GetDummyResponse, error) {
	o := system.Organization(ctx)
	return &dummyv1.GetDummyResponse{
		Value: fmt.Sprintf("%s, %s", o, r.Value),
	}, nil
}
