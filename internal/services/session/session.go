package session

import (
	sessionProtoV1 "github.com/cchirag/ira/proto/gen/services/v1/session"
	"go.etcd.io/bbolt"
)

type Service struct {
	sessionProtoV1.UnimplementedSessionServiceServer
	Db *bbolt.DB
}
