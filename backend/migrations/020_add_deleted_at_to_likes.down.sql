DROP INDEX IF EXISTS idx_likes_user_deleted;
DROP INDEX IF EXISTS idx_likes_active;
ALTER TABLE likes DROP COLUMN IF EXISTS deleted_at;
