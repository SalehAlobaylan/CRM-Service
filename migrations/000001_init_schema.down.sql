-- Drop all tables in reverse order of creation (respecting foreign key constraints)
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS customer_tags CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS notes CASCADE;
DROP TABLE IF EXISTS activities CASCADE;
DROP TABLE IF EXISTS deals CASCADE;
DROP TABLE IF EXISTS pipeline_stages CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
