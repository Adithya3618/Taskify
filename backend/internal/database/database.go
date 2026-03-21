package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	db, err := sql.Open("sqlite3", "./taskify.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	database := &DB{db}

	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database connected successfully")
	return database, nil
}

func (db *DB) createTables() error {
	// Create projects table
	projectsTable := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`

	// Create stages table
	stagesTable := `
	CREATE TABLE IF NOT EXISTS stages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		position INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	)
	`

	// Create tasks table
	tasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		stage_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		position INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (stage_id) REFERENCES stages(id) ON DELETE CASCADE
	)
	`

	// Create messages table
	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		project_id INTEGER NOT NULL,
		sender_name TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	)
	`

	tables := []string{projectsTable, stagesTable, tasksTable, messagesTable}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	if err := db.migrateLegacySchema(); err != nil {
		return fmt.Errorf("failed to migrate legacy schema: %v", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_stages_user ON stages(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_stages_project ON stages(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_stage ON tasks(stage_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_user ON messages(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_project ON messages(project_id)",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	log.Println("Database tables created successfully")
	return nil
}

func (db *DB) migrateLegacySchema() error {
	requiredColumns := map[string]map[string]string{
		"projects": {
			"user_id": "TEXT",
		},
		"stages": {
			"user_id": "TEXT",
		},
		"tasks": {
			"user_id": "TEXT",
		},
		"messages": {
			"user_id": "TEXT",
		},
	}

	for tableName, columns := range requiredColumns {
		for columnName, columnType := range columns {
			exists, err := db.columnExists(tableName, columnName)
			if err != nil {
				return err
			}
			if exists {
				continue
			}

			query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to add %s.%s: %v", tableName, columnName, err)
			}
			log.Printf("Migrated database: added %s.%s", tableName, columnName)
		}
	}

	return nil
}

func (db *DB) columnExists(tableName, columnName string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, fmt.Errorf("failed to inspect table %s: %v", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, fmt.Errorf("failed to scan schema for %s: %v", tableName, err)
		}

		if name == columnName {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("failed while reading schema for %s: %v", tableName, err)
	}

	return false, nil
}
