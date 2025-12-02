-- Add approval tracking columns
ALTER TABLE godplan.attendances 
ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES godplan.users(id),
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

-- Update existing 'approved' records to have approval metadata
UPDATE godplan.attendances 
SET approved_by = user_id, 
    approved_at = created_at 
WHERE status = 'approved' AND approved_by IS NULL;

-- Update existing 'forced' records to 'pending_forced' status
UPDATE godplan.attendances 
SET status = 'pending_forced'
WHERE status = 'forced' AND approved_by IS NULL;

-- Create index for faster queries on pending approvals
CREATE INDEX IF NOT EXISTS idx_attendances_status_pending 
ON godplan.attendances(status) 
WHERE status IN ('pending', 'pending_forced');

-- Create index for approved_by for reporting
CREATE INDEX IF NOT EXISTS idx_attendances_approved_by 
ON godplan.attendances(approved_by) 
WHERE approved_by IS NOT NULL;

SELECT 'Migration completed successfully!' AS result;
