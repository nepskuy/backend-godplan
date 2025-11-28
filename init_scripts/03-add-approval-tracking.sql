-- Migration: Add approval tracking to attendances table
-- Date: 2025-11-28
-- Description: Add columns for tracking approval status, approver, and rejection reason

-- Add approval tracking columns
ALTER TABLE attendances 
ADD COLUMN IF NOT EXISTS approved_by INTEGER REFERENCES users(id),
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

-- Update existing 'approved' records to have approval metadata
UPDATE attendances 
SET approved_by = user_id, 
    approved_at = created_at 
WHERE status = 'approved' AND approved_by IS NULL;

-- Update existing 'forced' records to 'pending_forced' status
UPDATE attendances 
SET status = 'pending_forced'
WHERE status = 'forced' AND approved_by IS NULL;

-- Create index for faster queries on pending approvals
CREATE INDEX IF NOT EXISTS idx_attendances_status_pending 
ON attendances(status) 
WHERE status IN ('pending', 'pending_forced');

-- Create index for approved_by for reporting
CREATE INDEX IF NOT EXISTS idx_attendances_approved_by 
ON attendances(approved_by) 
WHERE approved_by IS NOT NULL;

COMMENT ON COLUMN attendances.approved_by IS 'User ID of the supervisor/manager who approved or rejected this attendance';
COMMENT ON COLUMN attendances.approved_at IS 'Timestamp when the attendance was approved or rejected';
COMMENT ON COLUMN attendances.rejection_reason IS 'Reason provided when attendance is rejected';
