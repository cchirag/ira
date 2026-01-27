package root

import (
	"context"

	protov1 "github.com/cchirag/ira/proto/gen/services/v1"
	"go.etcd.io/bbolt"
)

type Service struct {
	protov1.UnimplementedRootServiceServer
	Db *bbolt.DB
}

func (s *Service) Ping(ctx context.Context, request *protov1.PingRequest) (*protov1.PingResponse, error) {
	var db bool
	if s.Db != nil {
		db = true
	}

	return &protov1.PingResponse{
		Db: db,
	}, nil
}
