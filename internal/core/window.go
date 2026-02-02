package core

import (
	"github.com/cchirag/ira/internal/storage"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

type Window struct {
	ID   uuid.UUID
	Name string
	Pane *Pane
}

func NewWindow(tx *bbolt.Tx, sessionId uuid.UUID, width, height int32) (*Window, error) {
	w, err := storage.NewWindow(tx, sessionId)
	if err != nil {
		return nil, err
	}

	pane, err := NewRootPane(tx, width, height)
	if err != nil {
		return nil, err
	}

	window := &Window{
		ID:   w.ID,
		Name: w.Name,
		Pane: pane,
	}

	return window, nil
}
