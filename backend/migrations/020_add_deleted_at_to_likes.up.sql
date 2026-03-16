-- likes 테이블에 soft delete 컬럼 추가
ALTER TABLE likes ADD COLUMN deleted_at TIMESTAMPTZ;

-- 활성 좋아요만 조회하기 위한 부분 인덱스
CREATE INDEX idx_likes_active ON likes(user_id, post_id) WHERE deleted_at IS NULL;

-- soft delete된 좋아요 조회용 인덱스 (사용자별 일괄 처리 시)
CREATE INDEX idx_likes_user_deleted ON likes(user_id) WHERE deleted_at IS NOT NULL;
