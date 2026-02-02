package core

import (
	"github.com/cchirag/ira/internal/enums"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

type Pane struct {
	ID                  uuid.UUID
	Type                enums.PaneType
	Width, Height, X, Y int32
	Left                *Pane
	Right               *Pane
}

func NewRootPane(tx *bbolt.Tx, width, height int32) (*Pane, error) {
	// Defer DB update for now and test things out

	return &Pane{
		ID:     uuid.New(),
		Type:   enums.HSplit,
		Width:  width,
		Height: height,
		Left: &Pane{
			ID:     uuid.New(),
			Type:   enums.Leaf,
			Width:  width,
			Height: height,
		},
	}, nil
}
