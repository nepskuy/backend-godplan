-- Migration: Create project phases table for internal project execution stages
-- Description: Separate from CRM phases - tracks INTERNAL project delivery timeline
-- Business Logic: Project phases track internal execution (task completion, delivery milestones)

CREATE TABLE IF NOT EXISTS godplan.project_phases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES godplan.tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6B46C1',
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    is_final BOOLEAN DEFAULT false,
    created_by BIGINT REFERENCES godplan.users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_project_phases_tenant ON godplan.project_phases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_project_phases_active ON godplan.project_phases(is_active);
CREATE INDEX IF NOT EXISTS idx_project_phases_order ON godplan.project_phases(tenant_id, display_order);
CREATE INDEX IF NOT EXISTS idx_project_phases_final ON godplan.project_phases(is_final);

COMMENT ON TABLE godplan.project_phases IS 'User-defined project execution phases for internal delivery tracking';
COMMENT ON COLUMN godplan.project_phases.is_final IS 'Marks phase as project completion stage';
COMMENT ON COLUMN godplan.project_phases.display_order IS 'Order for project timeline visualization';
