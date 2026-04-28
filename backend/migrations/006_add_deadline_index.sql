-- Add index on tasks.deadline for overdue task queries
-- Improves performance of GET /api/projects/:id/stats
CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline);