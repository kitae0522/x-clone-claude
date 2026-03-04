DROP INDEX IF EXISTS idx_posts_parent_id;
ALTER TABLE posts DROP COLUMN IF EXISTS reply_count;
ALTER TABLE posts DROP COLUMN IF EXISTS parent_id;
