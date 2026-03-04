ALTER TABLE posts ADD COLUMN parent_id UUID REFERENCES posts(id) ON DELETE CASCADE;
ALTER TABLE posts ADD COLUMN reply_count INT NOT NULL DEFAULT 0;

CREATE INDEX idx_posts_parent_id ON posts(parent_id);
