package root

import (
	"context"

	"github.com/cchirag/ira/internal/core"
	rootProtov1 "github.com/cchirag/ira/proto/gen/services/v1/root"
	"go.etcd.io/bbolt"
)

type Service struct {
	rootProtov1.UnimplementedRootServiceServer
	Db      *bbolt.DB
	CloseCh chan struct{}
}

func (s *Service) Ping(ctx context.Context, request *rootProtov1.PingRequest) (*rootProtov1.PingResponse, error) {
	var db bool
	if s.Db != nil {
		db = true
	}

	return &rootProtov1.PingResponse{
		Db: db,
	}, nil
}

func (s *Service) NewSession(ctx context.Context, request *rootProtov1.NewSessionRequest) (*rootProtov1.NewSessionResponse, error) {
	session := new(rootProtov1.NewSessionResponse)
	if err := s.Db.Update(func(tx *bbolt.Tx) error {
		s, err := core.NewSession(tx, request.Name)
		if err != nil {
			return err
		}

		session.Id = s.ID.String()
		session.Name = s.Name

		return nil
	}); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) Terminate(ctx context.Context, request *rootProtov1.TerminateRequest) (*rootProtov1.TerminateResponse, error) {
	close(s.CloseCh)
	return &rootProtov1.TerminateResponse{}, nil
}
