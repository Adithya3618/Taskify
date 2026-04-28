-- Add is_final column to stages table for proper completion tracking
-- Run this migration to enable final stage tracking

ALTER TABLE stages ADD COLUMN is_final INTEGER DEFAULT 0;

-- Index for stats query optimization
CREATE INDEX IF NOT EXISTS idx_stages_project_final ON stages(project_id, is_final);
