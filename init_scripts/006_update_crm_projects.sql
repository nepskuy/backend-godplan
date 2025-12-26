-- Migration: Update CRM projects table to use dynamic divisions and phases
-- Description: Replace hardcoded category/stage with references to divisions and CRM phases

-- Add new columns for dynamic configuration
ALTER TABLE godplan.crm_projects
ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES godplan.tenants(id),
ADD COLUMN IF NOT EXISTS division_id UUID REFERENCES godplan.divisions(id),
ADD COLUMN IF NOT EXISTS crm_phase_id UUID REFERENCES godplan.crm_phases(id);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_crm_projects_tenant ON godplan.crm_projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_crm_projects_division ON godplan.crm_projects(division_id);
CREATE INDEX IF NOT EXISTS idx_crm_projects_phase ON godplan.crm_projects(crm_phase_id);

-- Assign existing CRM projects to default tenant
UPDATE godplan.crm_projects
SET tenant_id = '00000000-0000-0000-0000-000000000001'::UUID
WHERE tenant_id IS NULL;

-- Note: category and stage columns kept for backward compatibility
-- They will be deprecated once all data is migrated to division_id and crm_phase_id
-- Future migration will drop these columns

COMMENT ON COLUMN godplan.crm_projects.division_id IS 'Reference to user-defined business division';
COMMENT ON COLUMN godplan.crm_projects.crm_phase_id IS 'Reference to CRM pipeline phase (external sales tracking)';
