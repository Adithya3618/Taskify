package testcases

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewDB_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_taskify_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("sql.Open() should not return nil")
	}

	if err := db.Ping(); err != nil {
		t.Errorf("db.Ping() error = %v", err)
	}
}

func TestSQLiteOperations(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_ops_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	t.Run("exec creates table", func(t *testing.T) {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS test_table (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL
			)
		`)
		if err != nil {
			t.Errorf("Failed to create table: %v", err)
		}
	})

	t.Run("exec inserts row", func(t *testing.T) {
		result, err := db.Exec("INSERT INTO test_table (name) VALUES (?)", "test name")
		if err != nil {
			t.Errorf("Failed to insert row: %v", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			t.Errorf("RowsAffected = %v, want 1", rowsAffected)
		}
	})

	t.Run("query selects row", func(t *testing.T) {
		var id int
		var name string
		err := db.QueryRow("SELECT id, name FROM test_table LIMIT 1").Scan(&id, &name)
		if err != nil {
			t.Errorf("Failed to query row: %v", err)
		}
		if name != "test name" {
			t.Errorf("name = %v, want 'test name'", name)
		}
	})
}

func TestTransactionSupport(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_tx_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	// Create table
	db.Exec("CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY)")

	// Test transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin() error = %v", err)
	}

	_, err = tx.Exec("INSERT INTO test VALUES (1)")
	if err != nil {
		tx.Rollback()
		t.Fatalf("tx.Exec() error = %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("tx.Commit() error = %v", err)
	}

	// Verify
	var count int
	db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if count != 1 {
		t.Errorf("count = %v, want 1", count)
	}
}
