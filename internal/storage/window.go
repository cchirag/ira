package storage

// Package storage implements persistent window storage for Ira using BoltDB.
//
// Windows are scoped to a session and stored by UUID. Each session has its
// own sub-bucket inside the main WINDOW bucket.
//
// BoltDB layout:
//
//   WINDOW (bucket)
//     └── <session-id-uuid> (bucket)
//           ├── <window-id-uuid> → JSON(WindowEntry)
//
// Notes:
//   - Windows have a unique ID (UUID) and a generated name for display.
//   - Index is stored for future reordering but has no inherent ordering role.
//   - All operations require a valid BoltDB transaction.
//   - Windows are tied to sessions; deleting a session should remove its windows.

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	nanoid "github.com/matoous/go-nanoid/v2"
	"go.etcd.io/bbolt"
)

var (
	ErrWindowNotFound              = errors.New("window not found")
	ErrWindowBucketNotFound        = errors.New("window bucket now found")
	ErrWindowSessionBucketNotFound = errors.New("window session bucket not found")
)

var windowBucketName = []byte("WINDOW")

type WindowEntry struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Index     int       `json:"index"`
	SessionID uuid.UUID `json:"sessionId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewWindow(tx *bbolt.Tx, sessionId uuid.UUID) (WindowEntry, error) {
	if tx == nil {
		return WindowEntry{}, ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionId)
	if err != nil {
		return WindowEntry{}, err
	}

	bucket, err := tx.CreateBucketIfNotExists(windowBucketName)
	if err != nil {
		return WindowEntry{}, err
	}

	sessionBucket, err := bucket.CreateBucketIfNotExists([]byte(session.ID.String()))
	if err != nil {
		return WindowEntry{}, err
	}

	stats := sessionBucket.Stats()
	index := stats.KeyN

	id, err := nanoid.Generate("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_-", 8)
	if err != nil {
		return WindowEntry{}, err
	}
	name := fmt.Sprintf("Window-%s", id)

	window := WindowEntry{
		ID:        uuid.New(),
		Name:      name,
		Index:     index,
		SessionID: session.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	bytes, err := json.Marshal(window)
	if err != nil {
		return WindowEntry{}, err
	}

	if err := sessionBucket.Put([]byte(window.ID.String()), bytes); err != nil {
		return WindowEntry{}, err
	}

	return window, nil
}

func GetWindow(tx *bbolt.Tx, sessionId, windowId uuid.UUID) (WindowEntry, error) {
	if tx == nil {
		return WindowEntry{}, ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionId)
	if err != nil {
		return WindowEntry{}, err
	}

	bucket := tx.Bucket(windowBucketName)
	if bucket == nil {
		return WindowEntry{}, ErrWindowBucketNotFound
	}

	sessionBucket := bucket.Bucket([]byte(session.ID.String()))
	if sessionBucket == nil {
		return WindowEntry{}, ErrWindowSessionBucketNotFound
	}

	entry := sessionBucket.Get([]byte(windowId.String()))
	if entry == nil {
		return WindowEntry{}, ErrWindowNotFound
	}

	var window WindowEntry
	if err := json.Unmarshal(entry, &window); err != nil {
		return WindowEntry{}, err
	}

	return window, nil
}

func GetWindows(tx *bbolt.Tx, sessionId uuid.UUID) ([]WindowEntry, error) {
	if tx == nil {
		return nil, ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionId)
	if err != nil {
		return nil, err
	}

	bucket := tx.Bucket(windowBucketName)
	if bucket == nil {
		return nil, ErrWindowBucketNotFound
	}

	sessionBucket := bucket.Bucket([]byte(session.ID.String()))
	if sessionBucket == nil {
		return nil, ErrWindowSessionBucketNotFound
	}

	stats := sessionBucket.Stats()

	windows := make([]WindowEntry, 0, stats.KeyN)

	if err = sessionBucket.ForEach(func(k, v []byte) error {
		var window WindowEntry
		if err = json.Unmarshal(v, &window); err != nil {
			return err
		}

		windows = append(windows, window)
		return nil
	}); err != nil {
		return nil, err
	}

	return windows, nil
}

func DeleteWindow(tx *bbolt.Tx, sessionId, windowId uuid.UUID) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionId)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(windowBucketName)
	if err != nil {
		return err
	}

	sessionBucket, err := bucket.CreateBucketIfNotExists([]byte(session.ID.String()))
	if err != nil {
		return err
	}

	if err := sessionBucket.Delete([]byte(windowId.String())); err != nil {
		return err
	}

	return nil
}

func DeleteWindows(tx *bbolt.Tx, sessionId uuid.UUID) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionId)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(windowBucketName)
	if err != nil {
		return err
	}

	sessionBucket, err := bucket.CreateBucketIfNotExists([]byte(session.ID.String()))
	if err != nil {
		return err
	}

	if err := sessionBucket.ForEach(func(k, v []byte) error {
		var window WindowEntry

		if err := json.Unmarshal(v, &window); err != nil {
			return err
		}

		if err := DeletePanes(tx, sessionId, window.ID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if err := bucket.DeleteBucket([]byte(session.ID.String())); err != nil {
		return err
	}

	return nil
}
