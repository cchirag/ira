package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cchirag/ira/internal/core/window"
	"go.etcd.io/bbolt"
)

type SessionStatus int

const (
	Attached SessionStatus = iota
	Detached
	Stopped
)

const SESSION_BUCKET = "SESSION"

type Session struct {
	Name      string        `json:"name" yaml:"name"`
	Template  string        `json:"template" yaml:"template"`
	CreatedAt time.Time     `json:"created_at" yaml:"created_at"`
	Status    SessionStatus `json:"status" yaml:"status"`
	Windows   []*window.Window
}

func New(ctx context.Context, db *bbolt.DB, name, template string) (Session, error) {
	// Check if the name is a valid name
	// Check if the template is a valid template
	// Add the data to the database with the name
	session := Session{
		Name:      name,
		Template:  template,
		CreatedAt: time.Now(),
		Status:    Stopped,
		Windows:   nil,
	}

	db.Update(func(tx *bbolt.Tx) error {
		if bucket, err := tx.CreateBucketIfNotExists([]byte(SESSION_BUCKET)); err != nil {
			return fmt.Errorf("error creating session bucket: %s", err.Error())
		} else {
			bytes, err := json.Marshal(session)
			if err != nil {
				return fmt.Errorf("error marshaling session data: %s", err.Error())
			}
			if err := bucket.Put([]byte(name), bytes); err != nil {
				return fmt.Errorf("error storing session data: %s", err.Error())
			}
		}
		return nil
	})

	return session, nil
}

func GetSession(ctx context.Context, db *bbolt.DB, name string) (Session, error) {
	return Session{}, nil
}

func GetAllSessions(ctx context.Context, db *bbolt.DB) ([]Session, error) {
	sessions := make([]Session, 0)

	return sessions, nil
}
