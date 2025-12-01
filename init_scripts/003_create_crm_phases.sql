-- Migration: Create CRM phases table for sales pipeline stages
-- Description: Replaces hardcoded stages (new, qualified, proposal, won) with user-defined CRM phases
-- Business Logic: CRM phases track EXTERNAL sales pipeline (closing deals)

CREATE TABLE IF NOT EXISTS godplan.crm_phases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES godplan.tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#8559F0',
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    is_final BOOLEAN DEFAULT false,
    success_indicator BOOLEAN DEFAULT false,
    created_by BIGINT REFERENCES godplan.users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_crm_phases_tenant ON godplan.crm_phases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_crm_phases_active ON godplan.crm_phases(is_active);
CREATE INDEX IF NOT EXISTS idx_crm_phases_order ON godplan.crm_phases(tenant_id, display_order);
CREATE INDEX IF NOT EXISTS idx_crm_phases_final ON godplan.crm_phases(is_final);

COMMENT ON TABLE godplan.crm_phases IS 'User-defined CRM pipeline phases for external sales tracking';
COMMENT ON COLUMN godplan.crm_phases.is_final IS 'Marks phase as final stage (won/lost)';
COMMENT ON COLUMN godplan.crm_phases.success_indicator IS 'true = won, false = lost (only for final phases)';
COMMENT ON COLUMN godplan.crm_phases.display_order IS 'Order for pipeline visualization';
