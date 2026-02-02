package core

import (
	"github.com/cchirag/ira/internal/enums"
	"github.com/cchirag/ira/internal/storage"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

type Session struct {
	ID      uuid.UUID
	Name    string
	Status  enums.SessionStatus
	Windows []*Window
}

func NewSession(tx *bbolt.Tx, name string) (*Session, error) {
	s, err := storage.NewSession(tx, name)
	if err != nil {
		return nil, err
	}

	window, err := NewWindow(tx, s.ID, 0, 0)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:      s.ID,
		Name:    s.Name,
		Status:  s.Status,
		Windows: []*Window{window},
	}

	return session, nil
}
