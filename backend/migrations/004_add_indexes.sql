-- Performance indexes for project_members table
-- Run this migration to optimize query performance

-- Index for fetching all members of a project
CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id);

-- Composite index for authorization checks (project + user lookup)
CREATE INDEX IF NOT EXISTS idx_project_members_project_user ON project_members(project_id, user_id);

-- Index for role-based queries
CREATE INDEX IF NOT EXISTS idx_project_members_role ON project_members(project_id, role);

-- Index for ordering by join date
CREATE INDEX IF NOT EXISTS idx_project_members_joined_at ON project_members(project_id, joined_at);

-- Index for projects table owner lookups
CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);
