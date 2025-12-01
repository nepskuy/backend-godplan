-- Migration: Create tenants table for multi-tenant support
-- Description: Each tenant represents a company/organization with their own data isolation

CREATE TABLE IF NOT EXISTS godplan.tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add tenant_id to users table for tenant association
ALTER TABLE godplan.users 
ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES godplan.tenants(id);

CREATE INDEX IF NOT EXISTS idx_users_tenant ON godplan.users(tenant_id);

-- Create default tenant for existing data
INSERT INTO godplan.tenants (id, name, slug, is_active)
VALUES (
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Default Organization',
    'default',
    true
) ON CONFLICT (slug) DO NOTHING;

-- Assign existing users to default tenant
UPDATE godplan.users 
SET tenant_id = '00000000-0000-0000-0000-000000000001'::UUID
WHERE tenant_id IS NULL;
