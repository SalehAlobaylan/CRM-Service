DROP INDEX IF EXISTS idx_audit_logs_tenant_id;
DROP INDEX IF EXISTS idx_customer_tags_tenant_id;
DROP INDEX IF EXISTS idx_tags_tenant_id;
DROP INDEX IF EXISTS idx_notes_tenant_id;
DROP INDEX IF EXISTS idx_activities_tenant_id;
DROP INDEX IF EXISTS idx_deals_tenant_id;
DROP INDEX IF EXISTS idx_pipeline_stages_tenant_id;
DROP INDEX IF EXISTS idx_contacts_tenant_id;
DROP INDEX IF EXISTS idx_customers_tenant_id;

ALTER TABLE audit_logs DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE customer_tags DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE tags DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE notes DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE activities DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE deals DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE pipeline_stages DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE contacts DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE customers DROP COLUMN IF EXISTS tenant_id;
