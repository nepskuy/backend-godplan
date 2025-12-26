# Database Migration Execution Order

## Migration Files (Execution Order)

Execute these migrations in the following order:

### Phase 1: Schema Initialization (Legacy Format)
1. `01-init-schema.sql` - Initialize godplan schema
2. `02-create-tables.sql` - Create base tables
3. `03-add-approval-tracking.sql` - Add attendance approval tracking

### Phase 2: Core Tables (New Format)
4. `001_create_tenants.sql` - Create tenants table
5. `002_create_divisions.sql` - Create divisions table

### Phase 3: CRM Setup
6. `003_create_crm_phases.sql` - Create CRM phases table
7. `005_create_crm_projects.sql` - Create CRM projects table
8. `006_update_crm_projects.sql` - Update CRM projects schema

### Phase 4: Project Management
9. `004_create_project_phases.sql` - Create project phases table
10. `007_update_projects.sql` - Update projects schema
11. `008_update_tasks.sql` - Update tasks schema

## Migration Naming Convention

**Going Forward**: Use the format `NNN_description.sql` where:
- `NNN` = 3-digit sequential number (001, 002, 003, etc.)
- `description` = snake_case description of what the migration does
- Use `create_` prefix for new tables
- Use `update_` prefix for schema changes
- Use `add_` prefix for adding columns/features

## Recent Changes

**2025-12-26**: Fixed duplicate version numbers
- Renamed `005_update_crm_projects.sql` → `006_update_crm_projects.sql`
- Renamed `006_update_projects.sql` → `007_update_projects.sql`
- Renamed `007_update_tasks.sql` → `008_update_tasks.sql`

## Notes

- Legacy files (`01-`, `02-`, `03-`) should be executed first
- New format files (`001_`, `002_`, etc.) follow after
- All migrations are idempotent (safe to run multiple times)
- Always backup database before running migrations in production

## Next Migration Number

Next migration should be: `009_description.sql`
