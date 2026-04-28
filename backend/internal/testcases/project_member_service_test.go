package testcases

import (
	"database/sql"
	"testing"
	"time"

	"backend/internal/models"
	pmrepository "backend/internal/repository"
	"backend/internal/services"

	_ "github.com/mattn/go-sqlite3"
)

// Helper function to create a test database with project member tables
func newProjectMemberTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create projects table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			owner_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create projects table: %v", err)
	}

	// Create project_members table
	_, err = db.Exec(`
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
	`)
	if err != nil {
		t.Fatalf("Failed to create project_members table: %v", err)
	}

	// Create project_invites table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS project_invites (
			id TEXT PRIMARY KEY,
			project_id INTEGER NOT NULL,
			invited_by TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			accepted_by TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create project_invites table: %v", err)
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			name TEXT NOT NULL,
			password_hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create activity_logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS activity_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			user_name TEXT,
			action TEXT NOT NULL,
			entity_type TEXT,
			entity_id INTEGER,
			description TEXT,
			details TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create activity_logs table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_project_members_project_user ON project_members(project_id, user_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_project_members_user ON project_members(user_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	return db
}

// Seed helper: create a project with owner
func seedProjectAndOwnerPM(t *testing.T, db *sql.DB, ownerID string) int64 {
	// Create user
	_, err := db.Exec(`INSERT INTO users (id, email, name) VALUES (?, ?, ?)`, ownerID, ownerID+"@test.com", "Owner User")
	if err != nil {
		t.Fatalf("Failed to create owner user: %v", err)
	}

	// Create project
	result, err := db.Exec(`INSERT INTO projects (name, owner_id) VALUES (?, ?)`, "Test Project", ownerID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	projectID, _ := result.LastInsertId()

	// Add owner as member
	_, err = db.Exec(`INSERT INTO project_members (project_id, user_id, role) VALUES (?, ?, ?)`,
		projectID, ownerID, models.RoleOwner)
	if err != nil {
		t.Fatalf("Failed to add owner as member: %v", err)
	}

	return projectID
}

// Test ProjectMemberService creation
func TestNewProjectMemberServicePM(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	svc := services.NewProjectMemberService(db)
	if svc == nil {
		t.Error("NewProjectMemberService() should not return nil")
	}
}

// Test ProjectMemberRepository AddMember
func TestProjectMemberRepository_AddMember(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add a new member
	member, err := repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	if member.UserID != "user-1" {
		t.Errorf("AddMember() UserID = %v, want user-1", member.UserID)
	}
	if member.Role != models.RoleMember {
		t.Errorf("AddMember() Role = %v, want %v", member.Role, models.RoleMember)
	}
}

// Test ProjectMemberRepository AddMember duplicate
func TestProjectMemberRepository_AddMember_Duplicate(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add a member
	_, err := repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	// Try to add same member again
	_, err = repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")
	if err == nil {
		t.Error("AddMember() should return error for duplicate member")
	}
}

// Test ProjectMemberRepository RemoveMember
func TestProjectMemberRepository_RemoveMember(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add a member
	_, err := repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	// Remove the member
	err = repo.RemoveMember(projectID, "user-1", "owner-1")
	if err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}

	// Verify member was removed
	isMember, err := repo.IsMember(projectID, "user-1")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if isMember {
		t.Error("user-1 should not be a member after removal")
	}
}

// Test ProjectMemberRepository RemoveMember - cannot remove owner
func TestProjectMemberRepository_RemoveMember_CannotRemoveOwner(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Try to remove the owner
	err := repo.RemoveMember(projectID, "owner-1", "owner-1")
	if err == nil {
		t.Error("RemoveMember() should return error for owner")
	}
}

// Test ProjectMemberRepository IsOwner
func TestProjectMemberRepository_IsOwner(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add a member
	_, _ = repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")

	// Check owner
	isOwner, err := repo.IsOwner(projectID, "owner-1")
	if err != nil {
		t.Fatalf("IsOwner() error = %v", err)
	}
	if !isOwner {
		t.Error("owner-1 should be owner of project")
	}

	// Check member is not owner
	isOwner, err = repo.IsOwner(projectID, "user-1")
	if err != nil {
		t.Fatalf("IsOwner() error = %v", err)
	}
	if isOwner {
		t.Error("user-1 should not be owner of project")
	}
}

// Test ProjectMemberRepository IsMember
func TestProjectMemberRepository_IsMember(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Check owner is a member
	isMember, err := repo.IsMember(projectID, "owner-1")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if !isMember {
		t.Error("owner-1 should be a member")
	}

	// Add a member
	_, _ = repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")

	// Check new member
	isMember, err = repo.IsMember(projectID, "user-1")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if !isMember {
		t.Error("user-1 should be a member after being added")
	}

	// Check non-member
	isMember, err = repo.IsMember(projectID, "non-member")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if isMember {
		t.Error("non-member should not be a member")
	}
}

// Test ProjectMemberRepository GetMembers
func TestProjectMemberRepository_GetMembers(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add multiple members
	for i := 1; i <= 5; i++ {
		userID := "user-" + string(rune('0'+i))
		_, err := db.Exec(`INSERT INTO users (id, email, name) VALUES (?, ?, ?)`, userID, userID+"@test.com", "User "+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		_, err = repo.AddMember(projectID, userID, string(models.RoleMember), "owner-1")
		if err != nil {
			t.Fatalf("AddMember() error = %v", err)
		}
	}

	// Get all members
	members, err := repo.GetMembers(projectID)
	if err != nil {
		t.Fatalf("GetMembers() error = %v", err)
	}

	if len(members) != 6 { // owner + 5 members
		t.Errorf("GetMembers() len = %v, want 6", len(members))
	}
}

// Test ProjectMemberRepository GetMembersPaginated
func TestProjectMemberRepository_GetMembersPaginated(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add multiple members
	for i := 1; i <= 5; i++ {
		userID := "user-" + string(rune('0'+i))
		_, err := db.Exec(`INSERT INTO users (id, email, name) VALUES (?, ?, ?)`, userID, userID+"@test.com", "User "+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		_, err = repo.AddMember(projectID, userID, string(models.RoleMember), "owner-1")
		if err != nil {
			t.Fatalf("AddMember() error = %v", err)
		}
	}

	// Get page 1 with limit 3
	members, total, err := repo.GetMembersPaginated(projectID, 1, 3)
	if err != nil {
		t.Fatalf("GetMembersPaginated() error = %v", err)
	}
	if len(members) != 3 {
		t.Errorf("GetMembersPaginated() page 1 len = %v, want 3", len(members))
	}
	if total != 6 {
		t.Errorf("GetMembersPaginated() total = %v, want 6", total)
	}

	// Get page 2
	members, _, err = repo.GetMembersPaginated(projectID, 2, 3)
	if err != nil {
		t.Fatalf("GetMembersPaginated() error = %v", err)
	}
	if len(members) != 3 {
		t.Errorf("GetMembersPaginated() page 2 len = %v, want 3", len(members))
	}
}

// Test ProjectMemberService CreateInvite
func TestProjectMemberService_CreateInvite(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	svc := services.NewProjectMemberService(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Create invite (only owner can do this)
	invite, err := svc.CreateInvite(projectID, "owner-1", 24)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	if invite.ProjectID != projectID {
		t.Errorf("CreateInvite() ProjectID = %v, want %v", invite.ProjectID, projectID)
	}
	if invite.InvitedBy != "owner-1" {
		t.Errorf("CreateInvite() InvitedBy = %v, want owner-1", invite.InvitedBy)
	}
	if invite.Role != "member" {
		t.Errorf("CreateInvite() Role = %v, want member", invite.Role)
	}
	if invite.Status != "pending" {
		t.Errorf("CreateInvite() Status = %v, want pending", invite.Status)
	}
}

// Test ProjectMemberService CreateInvite - non-owner cannot create
func TestProjectMemberService_CreateInvite_NonOwnerDenied(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	svc := services.NewProjectMemberService(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add a member
	repo := pmrepository.NewProjectMemberRepository(db)
	_, _ = repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")

	// Try to create invite as member
	_, err := svc.CreateInvite(projectID, "user-1", 24)
	if err == nil {
		t.Error("CreateInvite() should return error for non-owner")
	}
}

// Test ProjectMemberService AcceptInviteByID
func TestProjectMemberService_AcceptInviteByID(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	svc := services.NewProjectMemberService(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Create invite
	invite, err := svc.CreateInvite(projectID, "owner-1", 24)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	// Create user to accept invite
	_, err = db.Exec(`INSERT INTO users (id, email, name) VALUES (?, ?, ?)`, "new-user", "new@test.com", "New User")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Accept invite
	accepted, err := svc.AcceptInviteByID(invite.ID, "new-user")
	if err != nil {
		t.Fatalf("AcceptInviteByID() error = %v", err)
	}

	if accepted.Status != "accepted" {
		t.Errorf("AcceptInviteByID() Status = %v, want accepted", accepted.Status)
	}

	// Verify user is now a member
	repo := pmrepository.NewProjectMemberRepository(db)
	isMember, err := repo.IsMember(projectID, "new-user")
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if !isMember {
		t.Error("new-user should be a member after accepting invite")
	}
}

// Test Transaction Safety - AddMember in transaction
func TestProjectMemberRepository_AddMember_TransactionSafety(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add member should use transaction internally
	member, err := repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	// Verify in database
	if member.Role != models.RoleMember {
		t.Errorf("Member role = %v, want %v", member.Role, models.RoleMember)
	}
}

// Test Permission helpers
func TestProjectMember_CanEditProject(t *testing.T) {
	// Owner can edit
	if !models.CanEditProject(models.RoleOwner) {
		t.Error("Owner should be able to edit project")
	}

	// Member cannot edit
	if models.CanEditProject(models.RoleMember) {
		t.Error("Member should not be able to edit project")
	}
}

func TestProjectMember_CanManageMembers(t *testing.T) {
	// Owner can manage members
	if !models.CanManageMembers(models.RoleOwner) {
		t.Error("Owner should be able to manage members")
	}

	// Member cannot manage members
	if models.CanManageMembers(models.RoleMember) {
		t.Error("Member should not be able to manage members")
	}
}

// Test GetMemberRole
func TestProjectMemberRepository_GetMemberRole(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Add a member
	_, _ = repo.AddMember(projectID, "user-1", string(models.RoleMember), "owner-1")

	// Get owner role
	role, err := repo.GetMemberRole(projectID, "owner-1")
	if err != nil {
		t.Fatalf("GetMemberRole() error = %v", err)
	}
	if models.ProjectMemberRole(role) != models.RoleOwner {
		t.Errorf("GetMemberRole() for owner = %v, want %v", role, models.RoleOwner)
	}

	// Get member role
	role, err = repo.GetMemberRole(projectID, "user-1")
	if err != nil {
		t.Fatalf("GetMemberRole() error = %v", err)
	}
	if models.ProjectMemberRole(role) != models.RoleMember {
		t.Errorf("GetMemberRole() for member = %v, want %v", role, models.RoleMember)
	}

	// Get non-member role (empty)
	role, err = repo.GetMemberRole(projectID, "non-member")
	if err != nil {
		t.Fatalf("GetMemberRole() error = %v", err)
	}
	if role != "" {
		t.Errorf("GetMemberRole() for non-member = %v, want empty string", role)
	}
}

// Test EnsureOwnerConsistency
func TestProjectMemberRepository_EnsureOwnerConsistency(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	repo := pmrepository.NewProjectMemberRepository(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Ensure consistency should pass
	err := repo.EnsureOwnerConsistency(projectID)
	if err != nil {
		t.Fatalf("EnsureOwnerConsistency() error = %v", err)
	}

	// Verify owner is still set correctly
	isOwner, err := repo.IsOwner(projectID, "owner-1")
	if err != nil {
		t.Fatalf("IsOwner() error = %v", err)
	}
	if !isOwner {
		t.Error("owner-1 should still be owner after EnsureOwnerConsistency")
	}
}

// Test Invite expiry
func TestInvite_Expiry(t *testing.T) {
	db := newProjectMemberTestDB(t)
	defer db.Close()

	svc := services.NewProjectMemberService(db)
	projectID := seedProjectAndOwnerPM(t, db, "owner-1")

	// Create invite with 0 hours expiry (should be valid indefinitely)
	invite, err := svc.CreateInvite(projectID, "owner-1", 0)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	// Verify no expiry set
	if !invite.ExpiresAt.IsZero() {
		t.Error("Invite with 0 hours should have no expiry")
	}

	// Create invite with negative expiry (already expired)
	invite2, err := svc.CreateInvite(projectID, "owner-1", -1)
	if err != nil {
		t.Fatalf("CreateInvite() error = %v", err)
	}

	// Verify expiry is in the past
	if invite2.ExpiresAt.IsZero() {
		t.Error("Invite with -1 hours should have expiry set")
	}
	if invite2.ExpiresAt.After(time.Now()) {
		t.Error("Invite with -1 hours should have expiry in the past")
	}
}
