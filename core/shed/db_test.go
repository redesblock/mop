package shed

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/redesblock/hop/core/logging"
)

// TestNewDB constructs a new DB
// and validates if the schema is initialized properly.
func TestNewDB(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	s, err := db.getSchema()
	if err != nil {
		t.Fatal(err)
	}
	if s.Fields == nil {
		t.Error("schema fields are empty")
	}
	if len(s.Fields) != 0 {
		t.Errorf("got schema fields length %v, want %v", len(s.Fields), 0)
	}
	if s.Indexes == nil {
		t.Error("schema indexes are empty")
	}
	if len(s.Indexes) != 0 {
		t.Errorf("got schema indexes length %v, want %v", len(s.Indexes), 0)
	}
}

// TestDB_persistence creates one DB, saves a field and closes that DB.
// Then, it constructs another DB and trues to retrieve the saved value.
func TestDB_persistence(t *testing.T) {
	dir, err := ioutil.TempDir("", "shed-test-persistence")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	logger := logging.New(ioutil.Discard, 0)

	db, err := NewDB(dir, logger)
	if err != nil {
		t.Fatal(err)
	}
	stringField, err := db.NewStringField("preserve-me")
	if err != nil {
		t.Fatal(err)
	}
	want := "persistent value"
	err = stringField.Put(want)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	db2, err := NewDB(dir, logger)
	if err != nil {
		t.Fatal(err)
	}
	stringField2, err := db2.NewStringField("preserve-me")
	if err != nil {
		t.Fatal(err)
	}
	got, err := stringField2.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got string %q, want %q", got, want)
	}
}

// newTestDB is a helper function that constructs a
// temporary database and returns a cleanup function that must
// be called to remove the data.
func newTestDB(t *testing.T) (db *DB, cleanupFunc func()) {
	t.Helper()
	logger := logging.New(ioutil.Discard, 0)
	db, err := NewDB("", logger)
	if err != nil {
		t.Fatal(err)
	}
	return db, func() {
		db.Close()
	}
}
