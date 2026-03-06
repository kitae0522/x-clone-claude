CREATE TABLE post_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    media_type VARCHAR(10) NOT NULL,
    mime_type VARCHAR(50) NOT NULL,
    width INT,
    height INT,
    size_bytes BIGINT NOT NULL,
    duration_seconds FLOAT,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_post_media_post_id ON post_media(post_id);
CREATE INDEX idx_post_media_uploader_id ON post_media(uploader_id);
