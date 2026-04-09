package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	// Use absolute path for database
	dbPath := "./taskify.db"

	// Try to get absolute path
	if absPath, err := filepath.Abs(dbPath); err == nil {
		dbPath = absPath
	}

	log.Printf("Database path: %s", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
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
		owner_id TEXT,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`

	// Create project_members table for multi-user collaboration
	projectMembersTable := `
	CREATE TABLE IF NOT EXISTS project_members (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		invited_by TEXT,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		UNIQUE(project_id, user_id)
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

	// Create activity_logs table for audit trail
	activityLogsTable := `
	CREATE TABLE IF NOT EXISTS activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		action TEXT NOT NULL,
		target_user TEXT,
		details TEXT,
		ip_address TEXT,
		user_agent TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	)
	`

	// Create project_invites table for invite links
	projectInvitesTable := `
	CREATE TABLE IF NOT EXISTS project_invites (
		id TEXT PRIMARY KEY,
		project_id INTEGER NOT NULL,
		invited_by TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		accepted_by TEXT,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE
	)
	`

	tables := []string{projectsTable, stagesTable, tasksTable, messagesTable, projectMembersTable, activityLogsTable, projectInvitesTable}

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
		"CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id)",
		"CREATE INDEX IF NOT EXISTS idx_stages_user ON stages(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_stages_project ON stages(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_tasks_stage ON tasks(stage_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_user ON messages(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_project ON messages(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_project ON project_members(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_user ON project_members(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_members_project_user ON project_members(project_id, user_id)",
		"CREATE INDEX IF NOT EXISTS idx_activity_logs_project ON activity_logs(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_activity_logs_user ON activity_logs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON activity_logs(action)",
		"CREATE INDEX IF NOT EXISTS idx_project_invites_project ON project_invites(project_id)",
		"CREATE INDEX IF NOT EXISTS idx_project_invites_id ON project_invites(id)",
		"CREATE INDEX IF NOT EXISTS idx_project_invites_status ON project_invites(status)",
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
			"owner_id": "TEXT",
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

	// Backfill owner_id from user_id for existing projects
	if err := db.backfillOwnerID(); err != nil {
		return fmt.Errorf("failed to backfill owner_id: %v", err)
	}

	// Backfill project_members table for existing projects
	if err := db.backfillProjectMembers(); err != nil {
		return fmt.Errorf("failed to backfill project_members: %v", err)
	}

	return nil
}

// backfillOwnerID migrates existing projects by setting owner_id = user_id
func (db *DB) backfillOwnerID() error {
	// Check if user_id column exists (legacy schema)
	userIDExists, err := db.columnExists("projects", "user_id")
	if err != nil {
		return fmt.Errorf("failed to check user_id column: %v", err)
	}

	// If user_id doesn't exist, this is a fresh database - skip backfill
	if !userIDExists {
		log.Println("Database migration: fresh database, skipping owner_id backfill")
		return nil
	}

	// Check if owner_id needs backfilling (user_id exists but owner_id is empty)
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM projects 
		WHERE user_id IS NOT NULL 
		AND user_id != '' 
		AND (owner_id IS NULL OR owner_id = '')
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check owner_id backfill status: %v", err)
	}

	if count == 0 {
		log.Println("Database migration: owner_id already backfilled")
		return nil
	}

	// Backfill owner_id (safe because we checked user_id exists above)
	result, err := db.Exec(`
		UPDATE projects 
		SET owner_id = user_id 
		WHERE user_id IS NOT NULL 
		AND user_id != '' 
		AND (owner_id IS NULL OR owner_id = '')
	`)
	if err != nil {
		return fmt.Errorf("failed to backfill owner_id: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Database migration: backfilled owner_id for %d projects", rowsAffected)
	return nil
}

// backfillProjectMembers creates project_members entries for existing projects
func (db *DB) backfillProjectMembers() error {
	// Check if project_members table exists and has entries
	var tableExists int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='project_members'
	`).Scan(&tableExists)
	if err != nil || tableExists == 0 {
		// Table doesn't exist yet, will be created in createTables
		return nil
	}

	// Check if backfill is needed
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM projects p
		WHERE p.owner_id IS NOT NULL 
		AND p.owner_id != ''
		AND NOT EXISTS (
			SELECT 1 FROM project_members pm 
			WHERE pm.project_id = p.id AND pm.user_id = p.owner_id AND pm.role = 'owner'
		)
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check project_members backfill status: %v", err)
	}

	if count == 0 {
		log.Println("Database migration: project_members already backfilled")
		return nil
	}

	// Backfill project_members with owner as member
	result, err := db.Exec(`
		INSERT INTO project_members (project_id, user_id, role, invited_by, joined_at)
		SELECT p.id, p.owner_id, 'owner', p.owner_id, p.created_at
		FROM projects p
		WHERE p.owner_id IS NOT NULL 
		AND p.owner_id != ''
		AND NOT EXISTS (
			SELECT 1 FROM project_members pm 
			WHERE pm.project_id = p.id AND pm.user_id = p.owner_id AND pm.role = 'owner'
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to backfill project_members: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Database migration: backfilled %d project_members", rowsAffected)
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
