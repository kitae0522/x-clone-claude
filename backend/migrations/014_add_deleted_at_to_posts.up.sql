ALTER TABLE posts ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_posts_deleted_at ON posts(deleted_at) WHERE deleted_at IS NULL;
