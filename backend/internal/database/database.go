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

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_stages_project ON stages(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_stage ON tasks(stage_id)",
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