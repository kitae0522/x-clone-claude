CREATE INDEX IF NOT EXISTS idx_posts_author_deleted
ON posts (author_id, deleted_at DESC)
WHERE deleted_at IS NOT NULL;
