CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_trgm_users_username ON users
USING GIN (username gin_trgm_ops);

CREATE INDEX idx_trgm_users_description ON users
USING GIN (description gin_trgm_ops);

