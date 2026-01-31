package storage

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

func openTestDB(t *testing.T) *bbolt.DB {
	t.Helper()

	db, err := bbolt.Open("test.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
		_ = os.Remove("test.db")
	})

	return db
}

func withTx(t *testing.T, db *bbolt.DB, fn func(tx *bbolt.Tx) error) {
	t.Helper()

	if err := db.Update(fn); err != nil {
		t.Fatal(err)
	}
}

func TestStorageLifecycle(t *testing.T) {
	db := openTestDB(t)

	var (
		sessionID uuid.UUID
		windowID  uuid.UUID
		paneIDs   []uuid.UUID
	)

	// ---- create session, window, panes ----
	withTx(t, db, func(tx *bbolt.Tx) error {
		session, err := NewSession(tx, "test-session")
		if err != nil {
			t.Fatal(err)
		}
		sessionID = session.ID

		window, err := NewWindow(tx, sessionID)
		if err != nil {
			t.Fatal(err)
		}
		windowID = window.ID

		for i := range 3 {
			pane, err := NewPane(
				tx,
				sessionID,
				windowID,
				80,
				24,
				int32(i),
				0,
				"/tmp",
			)
			if err != nil {
				t.Fatal(err)
			}
			paneIDs = append(paneIDs, pane.ID)
		}

		return nil
	})

	// ---- verify reads ----
	withTx(t, db, func(tx *bbolt.Tx) error {
		session, err := GetSession(tx, sessionID)
		if err != nil {
			t.Fatal(err)
		}
		if session.Name != "test-session" {
			t.Fatalf("unexpected session name: %s", session.Name)
		}

		window, err := GetWindow(tx, sessionID, windowID)
		if err != nil {
			t.Fatal(err)
		}
		if window.Index != 0 {
			t.Fatalf("unexpected window index: %d", window.Index)
		}

		panes, err := GetPanes(tx, sessionID, windowID)
		if err != nil {
			t.Fatal(err)
		}
		if len(panes) != 3 {
			t.Fatalf("expected 3 panes, got %d", len(panes))
		}

		return nil
	})

	// ---- update pane ----
	withTx(t, db, func(tx *bbolt.Tx) error {
		if err := UpdatePaneSize(tx, sessionID, windowID, paneIDs[0], 120, 40); err != nil {
			t.Fatal(err)
		}

		if err := UpdatePanePosition(tx, sessionID, windowID, paneIDs[0], 5, 6); err != nil {
			t.Fatal(err)
		}

		if err := UpdatePaneCwd(tx, sessionID, windowID, paneIDs[0], "/home"); err != nil {
			t.Fatal(err)
		}

		pane, err := GetPane(tx, sessionID, windowID, paneIDs[0])
		if err != nil {
			t.Fatal(err)
		}

		if pane.Width != 120 || pane.Height != 40 {
			t.Fatal("pane size not updated")
		}
		if pane.X != 5 || pane.Y != 6 {
			t.Fatal("pane position not updated")
		}
		if pane.Cwd != "/home" {
			t.Fatal("pane cwd not updated")
		}

		return nil
	})

	// ---- delete single pane ----
	withTx(t, db, func(tx *bbolt.Tx) error {
		if err := DeletePane(tx, sessionID, windowID, paneIDs[1]); err != nil {
			t.Fatal(err)
		}

		_, err := GetPane(tx, sessionID, windowID, paneIDs[1])
		if err == nil {
			t.Fatal("expected deleted pane to be missing")
		}

		return nil
	})

	// ---- delete session (cascade windows + panes) ----
	withTx(t, db, func(tx *bbolt.Tx) error {
		if err := DeleteSession(tx, sessionID); err != nil {
			t.Fatal(err)
		}

		_, err := GetSession(tx, sessionID)
		if err == nil {
			t.Fatal("expected session to be deleted")
		}

		return nil
	})
}
