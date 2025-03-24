-- Create the extension and indexes for full-text search
-- Check article: https://niallburkley.com/blog/index-columns-for-like-in-postgres/
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_users_partner_id ON users(partner_id);