package storage

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

var ErrPaneNotFound = errors.New("pane not found")

var paneBucketName = []byte("PANE")

type PaneEntry struct {
	ID          uuid.UUID `json:"id"`
	SessionName string    `json:"sessionName"`
	WindowName  string    `json:"windowName"`
	Width       int32     `json:"width"`
	Height      int32     `json:"height"`
	X           int32     `json:"x"`
	Y           int32     `json:"y"`
	Cwd         string    `json:"cwd"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewPane(tx *bbolt.Tx, sessionName, windowName string, width, height, x, y int32, cwd string) (PaneEntry, error) {
	if tx == nil {
		return PaneEntry{}, ErrTxnNotFound
	}

	window, err := GetWindow(tx, sessionName, windowName)
	if err != nil {
		return PaneEntry{}, err
	}

	bucket, err := tx.CreateBucketIfNotExists(paneBucketName)
	if err != nil {
		return PaneEntry{}, err
	}

	windowBucket, err := bucket.CreateBucketIfNotExists([]byte(window.Name))
	if err != nil {
		return PaneEntry{}, err
	}

	pane := PaneEntry{
		ID:          uuid.New(),
		SessionName: sessionName,
		WindowName:  window.Name,
		Width:       width,
		Height:      height,
		X:           x,
		Y:           y,
		Cwd:         cwd,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	bytes, err := json.Marshal(pane)
	if err != nil {
		return PaneEntry{}, err
	}

	err = windowBucket.Put([]byte(pane.ID.String()), bytes)
	if err != nil {
		return PaneEntry{}, err
	}

	return pane, nil
}

func GetPane(tx *bbolt.Tx, sessionName, windowName string, id uuid.UUID) (PaneEntry, error) {
	if tx == nil {
		return PaneEntry{}, ErrTxnNotFound
	}

	window, err := GetWindow(tx, sessionName, windowName)
	if err != nil {
		return PaneEntry{}, err
	}

	bucket, err := tx.CreateBucketIfNotExists(paneBucketName)
	if err != nil {
		return PaneEntry{}, err
	}

	windowBucket, err := bucket.CreateBucketIfNotExists([]byte(window.Name))
	if err != nil {
		return PaneEntry{}, err
	}

	bytes := windowBucket.Get([]byte(id.String()))
	if bytes == nil {
		return PaneEntry{}, ErrPaneNotFound
	}

	var pane PaneEntry

	if err := json.Unmarshal(bytes, &pane); err != nil {
		return PaneEntry{}, err
	}

	return pane, nil
}

func GetPanes(tx *bbolt.Tx, sessionName, windowName string) ([]PaneEntry, error) {
	if tx == nil {
		return nil, ErrTxnNotFound
	}

	window, err := GetWindow(tx, sessionName, windowName)
	if err != nil {
		return nil, err
	}

	bucket, err := tx.CreateBucketIfNotExists(paneBucketName)
	if err != nil {
		return nil, err
	}

	windowBucket, err := bucket.CreateBucketIfNotExists([]byte(window.Name))
	if err != nil {
		return nil, err
	}

	panes := make([]PaneEntry, 0, windowBucket.Stats().KeyN)

	if err = windowBucket.ForEach(func(k, v []byte) error {
		var pane PaneEntry
		if err = json.Unmarshal(v, &pane); err != nil {
			return err
		}
		panes = append(panes, pane)
		return nil
	}); err != nil {
		return nil, err
	}

	return panes, nil
}

func DeletePane(tx *bbolt.Tx, sessionName, windowName string, id uuid.UUID) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	window, err := GetWindow(tx, sessionName, windowName)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(paneBucketName)
	if err != nil {
		return err
	}

	windowBucket, err := bucket.CreateBucketIfNotExists([]byte(window.Name))
	if err != nil {
		return err
	}

	return windowBucket.Delete([]byte(id.String()))
}

func UpdatePaneSize(tx *bbolt.Tx, sessionName, windowName string, id uuid.UUID, width, height int32) error {
	pane, err := GetPane(tx, sessionName, windowName, id)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(paneBucketName)
	if err != nil {
		return err
	}

	windowBucket, err := bucket.CreateBucketIfNotExists([]byte(pane.WindowName))
	if err != nil {
		return err
	}

	pane.Width, pane.Height, pane.UpdatedAt = width, height, time.Now()

	bytes, err := json.Marshal(pane)
	if err != nil {
		return err
	}

	return windowBucket.Put([]byte(id.String()), bytes)
}

func UpdatePanePosition(tx *bbolt.Tx, sessionName, windowName string, id uuid.UUID, x, y int32) error {
	pane, err := GetPane(tx, sessionName, windowName, id)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(paneBucketName)
	if err != nil {
		return err
	}

	windowBucket, err := bucket.CreateBucketIfNotExists([]byte(pane.WindowName))
	if err != nil {
		return err
	}

	pane.X, pane.Y, pane.UpdatedAt = x, y, time.Now()

	bytes, err := json.Marshal(pane)
	if err != nil {
		return err
	}

	return windowBucket.Put([]byte(id.String()), bytes)
}

func UpdatePaneCwd(tx *bbolt.Tx, sessionName, windowName string, id uuid.UUID, cwd string) error {
	pane, err := GetPane(tx, sessionName, windowName, id)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(paneBucketName)
	if err != nil {
		return err
	}

	windowBucket, err := bucket.CreateBucketIfNotExists([]byte(pane.WindowName))
	if err != nil {
		return err
	}

	pane.Cwd, pane.UpdatedAt = cwd, time.Now()

	bytes, err := json.Marshal(pane)
	if err != nil {
		return err
	}

	return windowBucket.Put([]byte(id.String()), bytes)
}
