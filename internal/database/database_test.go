package database

import (
	"strings"
	"testing"

	"github.com/mattn/go-sqlite3"
)

func TestDatabaseWritePermissionErrorGuidance(t *testing.T) {
	err := databaseWritePermissionError("/app/data", sqlite3.Error{Code: sqlite3.ErrReadonly})
	message := err.Error()
	for _, expected := range []string{"/app/data", "UID/GID 10001:10001", "sudo chown -R 10001:10001 <mounted-data-directory>", "readonly"} {
		if !strings.Contains(message, expected) {
			t.Errorf("error %q does not contain %q", message, expected)
		}
	}
	if !isDatabaseWritePermissionError(err) {
		t.Error("wrapped read-only SQLite error was not recognized")
	}
	if isDatabaseWritePermissionError(sqlite3.Error{Code: sqlite3.ErrCorrupt}) {
		t.Error("database corruption must not be reported as an ownership problem")
	}
}
