-- Migration: Create CRM projects table
-- Description: Stores sales pipeline deals/projects

CREATE TABLE IF NOT EXISTS godplan.crm_projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES godplan.tenants(id),
    division_id UUID REFERENCES godplan.divisions(id),
    crm_phase_id UUID REFERENCES godplan.crm_phases(id),
    
    title VARCHAR(200) NOT NULL,
    client VARCHAR(200) NOT NULL,
    value DECIMAL(15, 2) DEFAULT 0,
    urgency VARCHAR(20) DEFAULT 'medium', -- low, medium, high
    deadline TIMESTAMP,
    contact_person VARCHAR(200),
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    manager_id BIGINT REFERENCES godplan.users(id),
    
    -- Legacy fields for backward compatibility (optional, but good to have if frontend sends them)
    category VARCHAR(50),
    stage VARCHAR(50),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_crm_projects_tenant ON godplan.crm_projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_crm_projects_division ON godplan.crm_projects(division_id);
CREATE INDEX IF NOT EXISTS idx_crm_projects_phase ON godplan.crm_projects(crm_phase_id);
CREATE INDEX IF NOT EXISTS idx_crm_projects_manager ON godplan.crm_projects(manager_id);

COMMENT ON TABLE godplan.crm_projects IS 'Sales pipeline deals/projects';
