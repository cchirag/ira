package root

import (
	"context"
	"os"

	"github.com/cchirag/ira/internal/config"
	rootProtov1 "github.com/cchirag/ira/proto/gen/services/v1/root"
	"go.etcd.io/bbolt"
)

type Service struct {
	rootProtov1.UnimplementedRootServiceServer
	Db      *bbolt.DB
	CloseCh chan struct{}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

func (s *Service) Health(ctx context.Context, request *rootProtov1.HealthRequest) (*rootProtov1.HealthResponse, error) {
	return &rootProtov1.HealthResponse{
		Db:     fileExists(config.Current.DaemonDBPath),
		Mode:   config.Current.Mode,
		Log:    fileExists(config.Current.DaemonLogPath),
		Socket: fileExists(config.Current.DaemonSocketPath),
	}, nil
}

func (s *Service) Shutdown(ctx context.Context, request *rootProtov1.ShutdownRequest) (*rootProtov1.ShutdownResponse, error) {
	close(s.CloseCh)
	return &rootProtov1.ShutdownResponse{
		Message: "ira daemon shutting down",
	}, nil
}
