package session

import (
	"context"
	"time"

	sessionProtoV1 "github.com/cchirag/ira/proto/gen/services/v1/session"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	sessionProtoV1.UnimplementedSessionServiceServer
	Db *bbolt.DB
}

func (s *Service) NewSession(ctx context.Context, request *sessionProtoV1.NewSessionRequest) (*sessionProtoV1.NewSessionResponse, error) {
	return &sessionProtoV1.NewSessionResponse{
		Name:      request.Name,
		Template:  request.Template,
		CreatedAt: timestamppb.New(time.Now()),
		Status:    sessionProtoV1.SessionStatus_DETACHED,
	}, nil
}
