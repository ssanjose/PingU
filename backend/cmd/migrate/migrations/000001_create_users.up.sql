CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  email CITEXT UNIQUE NOT NULL,
  password BYTEA NOT NULL,
  pinged BOOLEAN DEFAULT FALSE,
  last_pinged_at TIMESTAMP(0) WITH TIME ZONE,
  verified BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
  pinged_partner_count INT DEFAULT 0,
  partner_id BIGSERIAL REFERENCES users(id)
)
