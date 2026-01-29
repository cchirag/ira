package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	"go.etcd.io/bbolt"
)

var ErrWindowNotFound = errors.New("window not found")

var windowBucketName = []byte("WINDOW")

type WindowEntry struct {
	Name        string    `json:"name"`
	Index       int       `json:"index"`
	SessionName string    `json:"sessionName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewWindow(tx *bbolt.Tx, sessionName string) (WindowEntry, error) {
	if tx == nil {
		return WindowEntry{}, ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionName)
	if err != nil {
		return WindowEntry{}, err
	}

	bucket, err := tx.CreateBucketIfNotExists(windowBucketName)
	if err != nil {
		return WindowEntry{}, err
	}

	sessionBucket, err := bucket.CreateBucketIfNotExists([]byte(session.Name))
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
		Name:        name,
		Index:       index,
		SessionName: session.Name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	bytes, err := json.Marshal(window)
	if err != nil {
		return WindowEntry{}, err
	}

	if err := sessionBucket.Put([]byte(window.Name), bytes); err != nil {
		return WindowEntry{}, err
	}

	return window, nil
}

func GetWindow(tx *bbolt.Tx, sessionName, name string) (WindowEntry, error) {
	if tx == nil {
		return WindowEntry{}, ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionName)
	if err != nil {
		return WindowEntry{}, err
	}

	bucket, err := tx.CreateBucketIfNotExists(windowBucketName)
	if err != nil {
		return WindowEntry{}, err
	}

	sessionBucket, err := bucket.CreateBucketIfNotExists([]byte(session.Name))
	if err != nil {
		return WindowEntry{}, err
	}

	// Get the entry with the name as key
	entry := sessionBucket.Get([]byte(name))
	if entry == nil {
		return WindowEntry{}, ErrWindowNotFound
	}

	var window WindowEntry
	if err := json.Unmarshal(entry, &window); err != nil {
		return WindowEntry{}, err
	}

	return window, nil
}

func GetWindows(tx *bbolt.Tx, sessionName string) ([]WindowEntry, error) {
	if tx == nil {
		return nil, ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionName)
	if err != nil {
		return nil, err
	}

	bucket, err := tx.CreateBucketIfNotExists(windowBucketName)
	if err != nil {
		return nil, err
	}

	sessionBucket, err := bucket.CreateBucketIfNotExists([]byte(session.Name))
	if err != nil {
		return nil, err
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

func DeleteWindow(tx *bbolt.Tx, sessionName, name string) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	session, err := GetSession(tx, sessionName)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(windowBucketName)
	if err != nil {
		return err
	}

	sessionBucket, err := bucket.CreateBucketIfNotExists([]byte(session.Name))
	if err != nil {
		return err
	}

	if err := sessionBucket.Delete([]byte(name)); err != nil {
		return err
	}

	return nil
}
