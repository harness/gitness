ALTER TABLE repositories ADD COLUMN repo_size INTEGER NOT NULL DEFAULT 0;
ALTER TABLE repositories ADD COLUMN repo_size_updated BIGINT NOT NULL DEFAULT 0;