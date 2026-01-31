package storage

// Package storage implements persistent session storage for Ira using BoltDB.
//
// Sessions are stored by UUID (internal, stable identifier) and indexed by
// name (user-facing, mutable identifier).
//
// BoltDB layout:
//
//   SESSION (bucket)
//     ├── <session-id-uuid> → JSON(SessionEntry)
//     └── __session_lookup__ (bucket)
//           └── <session-name> → <session-id-uuid>
//
// Invariants:
//   - Session UUIDs are the primary keys.
//   - Session names are unique and resolved via the lookup bucket.
//   - Renames and deletes update both buckets atomically.
//   - All operations must run inside a BoltDB transaction.

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/cchirag/ira/internal/enums"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

var (
	ErrEmptySessionName      = errors.New("empty session name")
	ErrInvalidSessionName    = errors.New("invalid name: must be 1–64 characters: letters, _, - only")
	ErrSessionAlreadyExists  = errors.New("session with the name already exists")
	ErrSessionNotFound       = errors.New("session not found")
	ErrTxnNotFound           = errors.New("db txn not found")
	ErrSessionBucketNotFound = errors.New("session bucket not found")
	ErrLookupBucketNotFound  = errors.New("lookup bucket not found")
)

var (
	sessionBucketName = []byte("SESSION")
	namePattern       = regexp.MustCompile(`^[A-Za-z_-]{1,64}$`)
	lookupBucketName  = []byte("__session_lookup__")
)

type SessionEntry struct {
	ID        uuid.UUID           `json:"id"`
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

	bucket, err := tx.CreateBucketIfNotExists(sessionBucketName)
	if err != nil {
		return SessionEntry{}, err
	}

	lookupBucket, err := bucket.CreateBucketIfNotExists(lookupBucketName)
	if err != nil {
		return SessionEntry{}, err
	}

	if _, exists, err := sessionWithNameExists(tx, name); err != nil {
		return SessionEntry{}, err
	} else if exists {
		return SessionEntry{}, ErrSessionAlreadyExists
	}

	uid, err := uuid.NewRandom()
	if err != nil {
		return SessionEntry{}, err
	}

	session := SessionEntry{
		ID:        uid,
		Name:      name,
		Status:    enums.Inactive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	bytes, err := json.Marshal(session)
	if err != nil {
		return SessionEntry{}, err
	}

	if err := bucket.Put([]byte(session.ID.String()), bytes); err != nil {
		return SessionEntry{}, err
	}

	if err := lookupBucket.Put([]byte(session.Name), []byte(session.ID.String())); err != nil {
		return SessionEntry{}, err
	}

	return session, nil
}

func sessionWithNameExists(tx *bbolt.Tx, name string) (uuid.UUID, bool, error) {
	if tx == nil {
		return uuid.UUID{}, false, ErrTxnNotFound
	}

	name, err := validateName(name)
	if err != nil {
		return uuid.UUID{}, false, err
	}

	bucket := tx.Bucket(sessionBucketName)
	if bucket == nil {
		return uuid.UUID{}, false, ErrSessionBucketNotFound
	}

	lookupBucket := bucket.Bucket(lookupBucketName)
	if lookupBucket == nil {
		return uuid.UUID{}, false, ErrLookupBucketNotFound
	}

	sessionId := lookupBucket.Get([]byte(name))

	if sessionId == nil {
		return uuid.UUID{}, false, nil
	}

	uid, err := uuid.ParseBytes(sessionId)
	if err != nil {
		return uuid.UUID{}, false, err
	}

	return uid, true, nil
}

func GetSession(tx *bbolt.Tx, id uuid.UUID) (SessionEntry, error) {
	if tx == nil {
		return SessionEntry{}, ErrTxnNotFound
	}

	bucket := tx.Bucket(sessionBucketName)
	if bucket == nil {
		return SessionEntry{}, ErrSessionBucketNotFound
	}

	entry := bucket.Get([]byte(id.String()))
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

	bucket := tx.Bucket(sessionBucketName)
	if bucket == nil {
		return nil, ErrSessionBucketNotFound
	}

	stat := bucket.Stats()
	sessions := make([]SessionEntry, 0, stat.KeyN)

	if err := bucket.ForEach(func(k, v []byte) error {
		if v == nil {
			return nil
		}

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

func UpdateSessionName(tx *bbolt.Tx, id uuid.UUID, name string) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	name, err := validateName(name)
	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists(sessionBucketName)
	if err != nil {
		return err
	}

	lookupBucket, err := bucket.CreateBucketIfNotExists(lookupBucketName)
	if err != nil {
		return err
	}

	if existing := lookupBucket.Get([]byte(name)); existing != nil {
		if string(existing) != id.String() {
			return ErrSessionAlreadyExists
		}
	}

	old := bucket.Get([]byte(id.String()))
	if old == nil {
		return ErrSessionNotFound
	}

	var session SessionEntry
	if err = json.Unmarshal(old, &session); err != nil {
		return err
	}
	oldName := session.Name

	session.Name, session.UpdatedAt = name, time.Now()

	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	if err := bucket.Put([]byte(id.String()), bytes); err != nil {
		return err
	}

	if err := lookupBucket.Put([]byte(session.Name), []byte(session.ID.String())); err != nil {
		return err
	}

	if err := lookupBucket.Delete([]byte(oldName)); err != nil {
		return err
	}

	return nil
}

func UpdateSessionStatus(tx *bbolt.Tx, id uuid.UUID, status enums.SessionStatus) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	bucket, err := tx.CreateBucketIfNotExists(sessionBucketName)
	if err != nil {
		return err
	}

	old := bucket.Get([]byte(id.String()))
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

	if err := bucket.Put([]byte(id.String()), bytes); err != nil {
		return err
	}

	return nil
}

func DeleteSession(tx *bbolt.Tx, id uuid.UUID) error {
	if tx == nil {
		return ErrTxnNotFound
	}

	bucket, err := tx.CreateBucketIfNotExists(sessionBucketName)
	if err != nil {
		return err
	}

	lookupBucket, err := bucket.CreateBucketIfNotExists(lookupBucketName)
	if err != nil {
		return err
	}

	session, err := GetSession(tx, id)
	if err != nil {
		return err
	}

	if err := DeleteWindows(tx, session.ID); err != nil {
		return err
	}

	if err := bucket.Delete([]byte(session.ID.String())); err != nil {
		return err
	}

	if err := lookupBucket.Delete([]byte(session.Name)); err != nil {
		return err
	}

	return nil
}
