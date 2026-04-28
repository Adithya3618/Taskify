-- Add is_final column to stages table for proper completion tracking
-- Run this migration to enable final stage tracking

ALTER TABLE stages ADD COLUMN is_final INTEGER DEFAULT 0;
