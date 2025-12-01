-- Migration: Create divisions table for dynamic business divisions
-- Description: Replaces hardcoded categories (godjah, godtive, godweb) with user-defined divisions

CREATE TABLE IF NOT EXISTS godplan.divisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES godplan.tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#8559F0',
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    created_by BIGINT REFERENCES godplan.users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_divisions_tenant ON godplan.divisions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_divisions_active ON godplan.divisions(is_active);
CREATE INDEX IF NOT EXISTS idx_divisions_order ON godplan.divisions(tenant_id, display_order);

COMMENT ON TABLE godplan.divisions IS 'User-defined business divisions for CRM categorization';
COMMENT ON COLUMN godplan.divisions.slug IS 'URL-friendly identifier, unique per tenant';
COMMENT ON COLUMN godplan.divisions.color IS 'Hex color code for UI visualization';
COMMENT ON COLUMN godplan.divisions.display_order IS 'Order for displaying divisions in UI';
