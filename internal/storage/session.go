package storage

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/cchirag/ira/internal/enums"
	"go.etcd.io/bbolt"
)

var (
	ErrEmptySessionName     = errors.New("empty session name")
	ErrInvalidSessionName   = errors.New("invalid name: must be 1–64 characters: letters, _, - only")
	ErrSessionAlreadyExists = errors.New("session with the name already exists")
	ErrSessionNotFound      = errors.New("session not found")
	ErrTxnNotFound          = errors.New("db txn not found")
)

var (
	bucketName  = []byte("SESSION")
	namePattern = regexp.MustCompile(`^[A-Za-z_-]{1,64}$`)
)

type SessionEntry struct {
	Name      string              `json:"name"`
	Status    enums.SessionStatus `json:"status"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrEmptySessionName
	}

	if ok := namePattern.MatchString(name); !ok {
		return "", ErrInvalidSessionName
	}

	return name, nil
}

func NewSession(tx *bbolt.Tx, name string) (SessionEntry, error) {
	if tx == nil {
		return SessionEntry{}, ErrTxnNotFound
	}

	name, err := validateName(name)
	if err != nil {
		return SessionEntry{}, err
	}

	bucket, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return SessionEntry{}, err
	}

	if entry := bucket.Get([]byte(name)); entry != nil {
		return SessionEntry{}, ErrSessionAlreadyExists
	}

	session := SessionEntry{
		Name:      name,
		Status:    enums.Inactive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	bytes, err := json.Marshal(session)
	if err != nil {
		return SessionEntry{}, err
	}

	if err := bucket.Put([]byte(name), bytes); err != nil {
		return SessionEntry{}, err
	}

	return session, nil
}

func GetSession(tx *bbolt.Tx, name string) (SessionEntry, error) {
	if tx == nil {
		return SessionEntry{}, ErrTxnNotFound
	}

	name, err := validateName(name)
	if err != nil {
		return SessionEntry{}, err
	}

	bucket, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return SessionEntry{}, err
	}

	entry := bucket.Get([]byte(name))
	if entry == nil {
		return SessionEntry{}, ErrSessionNotFound
	}

	var session SessionEntry

	if err := json.Unmarshal(entry, &session); err != nil {
		return SessionEntry{}, err
	}

	return session, nil
}

func GetSessions(tx *bbolt.Tx) ([]SessionEntry, error) {
	if tx == nil {
		return nil, ErrTxnNotFound
	}

	bucket, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return nil, err
	}

	stat := bucket.Stats()
	sessions := make([]SessionEntry, 0, stat.KeyN)

	if err := bucket.ForEach(func(k, v []byte) error {
		var session SessionEntry

		err := json.Unmarshal(v, &session)
		if err != nil {
			return err
		}

		sessions = append(sessions, session)

		return nil
	}); err != nil {
		return nil, err
	}

	return sessions, nil
}

func UpdateSessionName(tx *bbolt.Tx, name, newName string) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	name, err := validateName(name)
	if err != nil {
		return err
	}

	newName, err = validateName(newName)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return err
	}

	old := bucket.Get([]byte(name))
	if old == nil {
		return ErrSessionNotFound
	}

	var session SessionEntry
	if err = json.Unmarshal(old, &session); err != nil {
		return err
	}

	session.Name = newName
	session.UpdatedAt = time.Now()

	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	if err := bucket.Put([]byte(newName), bytes); err != nil {
		return err
	}

	if err := bucket.Delete([]byte(name)); err != nil {
		return err
	}

	return nil
}

func UpdateSessionStatus(tx *bbolt.Tx, name string, status enums.SessionStatus) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	name, err := validateName(name)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return err
	}

	old := bucket.Get([]byte(name))
	if old == nil {
		return ErrSessionNotFound
	}

	var session SessionEntry
	if err = json.Unmarshal(old, &session); err != nil {
		return err
	}

	session.Status = status
	session.UpdatedAt = time.Now()

	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	if err := bucket.Put([]byte(name), bytes); err != nil {
		return err
	}

	return nil
}

func DeleteSession(tx *bbolt.Tx, name string) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	name, err := validateName(name)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		return err
	}

	if err := bucket.Delete([]byte(name)); err != nil {
		return err
	}

	return nil
}
