DROP INDEX IF EXISTS idx_posts_deleted_at;
ALTER TABLE posts DROP COLUMN IF EXISTS deleted_at;
