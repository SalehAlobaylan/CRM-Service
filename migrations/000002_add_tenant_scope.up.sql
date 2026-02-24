ALTER TABLE customers ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE pipeline_stages ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE deals ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE activities ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE notes ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE tags ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE customer_tags ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64);

UPDATE customers SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE contacts SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE pipeline_stages SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE deals SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE activities SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE notes SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE tags SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE customer_tags SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';
UPDATE audit_logs SET tenant_id = 'default' WHERE tenant_id IS NULL OR tenant_id = '';

ALTER TABLE customers ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE contacts ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE pipeline_stages ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE deals ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE activities ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE notes ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE tags ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE customer_tags ALTER COLUMN tenant_id SET DEFAULT 'default';
ALTER TABLE audit_logs ALTER COLUMN tenant_id SET DEFAULT 'default';

ALTER TABLE customers ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE contacts ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE pipeline_stages ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE deals ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE activities ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE notes ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE tags ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE customer_tags ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE audit_logs ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_contacts_tenant_id ON contacts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_stages_tenant_id ON pipeline_stages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_deals_tenant_id ON deals(tenant_id);
CREATE INDEX IF NOT EXISTS idx_activities_tenant_id ON activities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notes_tenant_id ON notes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tags_tenant_id ON tags(tenant_id);
CREATE INDEX IF NOT EXISTS idx_customer_tags_tenant_id ON customer_tags(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
