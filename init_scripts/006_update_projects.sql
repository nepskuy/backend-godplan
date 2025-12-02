-- Migration: Update projects table for CRM → Project conversion
-- Description: Link projects to CRM deals and add project execution tracking

-- Add columns for CRM conversion and project management
ALTER TABLE godplan.projects
ADD COLUMN IF NOT EXISTS crm_id UUID REFERENCES godplan.crm_projects(id),
ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES godplan.tenants(id),
ADD COLUMN IF NOT EXISTS division_id UUID REFERENCES godplan.divisions(id),
ADD COLUMN IF NOT EXISTS assigned_to UUID REFERENCES godplan.users(id),
ADD COLUMN IF NOT EXISTS team_members TEXT[],
ADD COLUMN IF NOT EXISTS current_phase_id UUID REFERENCES godplan.project_phases(id),
ADD COLUMN IF NOT EXISTS progress INT DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
ADD COLUMN IF NOT EXISTS expected_completion_date TIMESTAMP;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_projects_crm ON godplan.projects(crm_id);
CREATE INDEX IF NOT EXISTS idx_projects_tenant ON godplan.projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_division ON godplan.projects(division_id);
CREATE INDEX IF NOT EXISTS idx_projects_assigned ON godplan.projects(assigned_to);
CREATE INDEX IF NOT EXISTS idx_projects_phase ON godplan.projects(current_phase_id);

-- Assign existing projects to default tenant
UPDATE godplan.projects
SET tenant_id = '00000000-0000-0000-0000-000000000001'::UUID
WHERE tenant_id IS NULL;

COMMENT ON COLUMN godplan.projects.crm_id IS 'Link to original CRM deal (if converted from CRM)';
COMMENT ON COLUMN godplan.projects.assigned_to IS 'Project manager/owner';
COMMENT ON COLUMN godplan.projects.team_members IS 'Array of user IDs assigned to this project';
COMMENT ON COLUMN godplan.projects.current_phase_id IS 'Current project execution phase (internal tracking)';
COMMENT ON COLUMN godplan.projects.progress IS 'Overall project completion percentage (0-100)';
