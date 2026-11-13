package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the SQL database handle and provides data access methods.
type DB struct {
	*sql.DB
}

// New opens an embedded SQLite database at the specified path.
// If the parent directory does not exist and path is not in-memory, it creates the directory.
func New(dbPath string) (*DB, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	dsn := dbPath
	if dbPath == ":memory:" {
		dsn = "file:memdb1?mode=memory&cache=shared"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize SQLite performance for concurrent readers
	if _, err := db.Exec("PRAGMA journal_mode = WAL; PRAGMA synchronous = NORMAL; PRAGMA foreign_keys = ON;"); err != nil {
		// WAL might not be supported in pure memory mode, ignore error if in-memory
		if dbPath != ":memory:" {
			return nil, fmt.Errorf("failed to set pragmas: %w", err)
		}
	}

	return &DB{DB: db}, nil
}
