-- Migration: Update tasks table for project task management
-- Description: Link tasks to projects and project phases with progress tracking

-- Add columns for project task management
ALTER TABLE godplan.tasks
ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES godplan.projects(id) ON DELETE CASCADE,
ADD COLUMN IF NOT EXISTS phase_id UUID REFERENCES godplan.project_phases(id),
ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES godplan.tenants(id),
ADD COLUMN IF NOT EXISTS progress INT DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
ADD COLUMN IF NOT EXISTS weight INT DEFAULT 1 CHECK (weight > 0),
ADD COLUMN IF NOT EXISTS estimated_hours DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS actual_hours DECIMAL(10,2);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_project ON godplan.tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_phase ON godplan.tasks(phase_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant ON godplan.tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assigned ON godplan.tasks(assigned_to);

-- Assign existing tasks to default tenant
UPDATE godplan.tasks
SET tenant_id = '00000000-0000-0000-0000-000000000001'::UUID
WHERE tenant_id IS NULL;

COMMENT ON COLUMN godplan.tasks.project_id IS 'Link to parent project (if task is part of a project)';
COMMENT ON COLUMN godplan.tasks.phase_id IS 'Current project phase for this task';
COMMENT ON COLUMN godplan.tasks.progress IS 'Task completion percentage (0-100)';
COMMENT ON COLUMN godplan.tasks.weight IS 'Weight for calculating project completion (higher = more important)';
COMMENT ON COLUMN godplan.tasks.estimated_hours IS 'Estimated time to complete task';
COMMENT ON COLUMN godplan.tasks.actual_hours IS 'Actual time spent on task';
