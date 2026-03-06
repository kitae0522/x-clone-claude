# Database Schema Design

## ER Diagram (Mermaid)

```mermaid
erDiagram
    users ||--o{ posts : "작성"
    users ||--o{ follows : "팔로워"
    users ||--o{ follows : "팔로잉"
    users ||--o{ likes : "좋아요"
    users ||--o{ bookmarks : "북마크"
    users ||--o{ post_media : "업로드"
    users ||--o{ poll_votes : "투표"
    posts ||--o{ posts : "답글 (parent_id)"
    posts ||--o{ likes : "좋아요"
    posts ||--o{ bookmarks : "북마크"
    posts ||--o{ post_media : "미디어"
    posts ||--o| polls : "투표"
    polls ||--o{ poll_options : "선택지"
    polls ||--o{ poll_votes : "투표"

    users {
        uuid id PK
        varchar email UK
        varchar password_hash
        varchar username UK
        varchar display_name
        text bio
        text profile_image_url
        text header_image_url
        timestamptz created_at
        timestamptz updated_at
    }

    posts {
        uuid id PK
        uuid author_id FK
        uuid parent_id FK "nullable, self-ref"
        text content
        varchar visibility "public|follower|private"
        int like_count
        int reply_count
        int view_count
        float location_lat "nullable"
        float location_lng "nullable"
        varchar location_name "nullable"
        timestamptz deleted_at "nullable, soft delete"
        timestamptz created_at
        timestamptz updated_at
    }

    follows {
        uuid follower_id PK_FK
        uuid following_id PK_FK
        timestamptz created_at
    }

    likes {
        uuid user_id PK_FK
        uuid post_id PK_FK
        timestamptz created_at
    }

    bookmarks {
        uuid user_id PK_FK
        uuid post_id PK_FK
        timestamptz created_at
    }

    post_media {
        uuid id PK
        uuid post_id FK "nullable"
        uuid uploader_id FK
        text url
        varchar media_type
        varchar mime_type
        int width "nullable"
        int height "nullable"
        bigint size_bytes
        float duration_seconds "nullable"
        smallint sort_order
        timestamptz created_at
    }

    polls {
        uuid id PK
        uuid post_id FK_UK
        timestamptz expires_at
        int total_votes
        timestamptz created_at
    }

    poll_options {
        uuid id PK
        uuid poll_id FK
        smallint option_index
        varchar text
        int vote_count
    }

    poll_votes {
        uuid id PK
        uuid poll_id FK
        uuid user_id FK
        smallint option_index
        timestamptz created_at
    }
```

## Tables

### `users`

사용자 계정 정보.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `uuid_generate_v4()` | 고유 식별자 |
| `email` | `VARCHAR(255)` | UNIQUE, NOT NULL | 이메일 주소 |
| `password_hash` | `VARCHAR(255)` | NOT NULL | bcrypt 해시 비밀번호 |
| `username` | `VARCHAR(50)` | UNIQUE, NOT NULL | 로그인 및 멘션용 핸들 |
| `display_name` | `VARCHAR(100)` | NOT NULL, DEFAULT `''` | 화면 표시 이름 |
| `bio` | `TEXT` | NOT NULL, DEFAULT `''` | 자기소개 |
| `profile_image_url` | `TEXT` | NOT NULL, DEFAULT `''` | 프로필 이미지 URL |
| `header_image_url` | `TEXT` | NOT NULL, DEFAULT `''` | 헤더 이미지 URL |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 가입일 |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 수정일 |

### `posts`

게시글 및 답글. `parent_id`가 있으면 답글, 없으면 일반 게시글.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | 고유 식별자 |
| `author_id` | `UUID` | FK → `users.id`, NOT NULL | 작성자 |
| `parent_id` | `UUID` | FK → `posts.id`, nullable | 부모 게시글 (답글일 때) |
| `content` | `TEXT` | NOT NULL | 게시글 내용 (500자 제한, 마크다운 지원) |
| `visibility` | `VARCHAR(20)` | NOT NULL, DEFAULT `'public'` | 공개 범위 (public/follower/private) |
| `like_count` | `INT` | NOT NULL, DEFAULT `0` | 좋아요 수 |
| `reply_count` | `INT` | NOT NULL, DEFAULT `0` | 답글 수 |
| `view_count` | `INT` | NOT NULL, DEFAULT `0` | 조회수 |
| `location_lat` | `DOUBLE PRECISION` | nullable | 위치 위도 |
| `location_lng` | `DOUBLE PRECISION` | nullable | 위치 경도 |
| `location_name` | `VARCHAR(100)` | nullable | 위치 이름 |
| `deleted_at` | `TIMESTAMPTZ` | nullable | Soft delete 시각 |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 작성일 |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 수정일 |

### `follows`

팔로우 관계. 단방향 (follower → following).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `follower_id` | `UUID` | PK, FK → `users.id` | 팔로우 하는 사용자 |
| `following_id` | `UUID` | PK, FK → `users.id` | 팔로우 당하는 사용자 |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 팔로우 시각 |

**제약:** `CHECK(follower_id <> following_id)` — 셀프 팔로우 불가

### `likes`

게시글 좋아요.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `user_id` | `UUID` | PK, FK → `users.id` | 좋아요 한 사용자 |
| `post_id` | `UUID` | PK, FK → `posts.id` | 대상 게시글 |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 좋아요 시각 |

### `bookmarks`

게시글 북마크.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `user_id` | `UUID` | PK, FK → `users.id` | 북마크 한 사용자 |
| `post_id` | `UUID` | PK, FK → `posts.id` | 대상 게시글 |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 북마크 시각 |

### `post_media`

게시글 첨부 미디어.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | 고유 식별자 |
| `post_id` | `UUID` | FK → `posts.id`, nullable | 연결된 게시글 |
| `uploader_id` | `UUID` | FK → `users.id`, NOT NULL | 업로더 |
| `url` | `TEXT` | NOT NULL | 미디어 URL |
| `media_type` | `VARCHAR(10)` | NOT NULL | 타입 (image/video) |
| `mime_type` | `VARCHAR(50)` | NOT NULL | MIME 타입 |
| `width` | `INT` | nullable | 가로 크기 |
| `height` | `INT` | nullable | 세로 크기 |
| `size_bytes` | `BIGINT` | NOT NULL | 파일 크기 (bytes) |
| `duration_seconds` | `FLOAT` | nullable | 동영상 길이 |
| `sort_order` | `SMALLINT` | NOT NULL, DEFAULT `0` | 정렬 순서 |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 업로드 시각 |

### `polls`

게시글 투표.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | 고유 식별자 |
| `post_id` | `UUID` | FK → `posts.id`, UNIQUE, NOT NULL | 연결된 게시글 |
| `expires_at` | `TIMESTAMPTZ` | NOT NULL | 투표 마감일 |
| `total_votes` | `INT` | NOT NULL, DEFAULT `0` | 총 투표 수 |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 생성일 |

### `poll_options`

투표 선택지.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | 고유 식별자 |
| `poll_id` | `UUID` | FK → `polls.id`, NOT NULL | 소속 투표 |
| `option_index` | `SMALLINT` | NOT NULL, UNIQUE(poll_id, option_index) | 선택지 순서 |
| `text` | `VARCHAR(25)` | NOT NULL | 선택지 텍스트 |
| `vote_count` | `INT` | NOT NULL, DEFAULT `0` | 득표 수 |

### `poll_votes`

사용자 투표 기록.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | 고유 식별자 |
| `poll_id` | `UUID` | FK → `polls.id`, NOT NULL | 대상 투표 |
| `user_id` | `UUID` | FK → `users.id`, NOT NULL, UNIQUE(poll_id, user_id) | 투표한 사용자 |
| `option_index` | `SMALLINT` | NOT NULL | 선택한 옵션 |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | 투표 시각 |

## Indexes

```sql
-- users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- posts
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_parent_id ON posts(parent_id);
CREATE INDEX idx_posts_deleted_at ON posts(deleted_at) WHERE deleted_at IS NULL;

-- follows
CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_following_id ON follows(following_id);

-- likes
CREATE INDEX idx_likes_post_id ON likes(post_id);

-- bookmarks
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC);

-- post_media
CREATE INDEX idx_post_media_post_id ON post_media(post_id);
CREATE INDEX idx_post_media_uploader_id ON post_media(uploader_id);

-- polls
CREATE INDEX idx_polls_post_id ON polls(post_id);
CREATE INDEX idx_poll_options_poll_id ON poll_options(poll_id);
CREATE INDEX idx_poll_votes_poll_id ON poll_votes(poll_id);
CREATE INDEX idx_poll_votes_user_id ON poll_votes(user_id);
```

## 게시글 권한 조회 로직

`visibility`에 따른 게시글 조회 필터링:

```sql
SELECT p.*
FROM posts p
WHERE p.deleted_at IS NULL
  AND (
    -- 전체 공개
    p.visibility = 'public'
    -- 본인 게시글
    OR p.author_id = :current_user_id
    -- 팔로워 공개 + 팔로우 관계 확인
    OR (
      p.visibility = 'follower'
      AND EXISTS (
        SELECT 1 FROM follows f
        WHERE f.follower_id = :current_user_id
          AND f.following_id = p.author_id
      )
    )
  )
ORDER BY p.created_at DESC;
```

## Seed Data

데모용 초기 데이터는 `migrations/015_seed_data.up.sql`에서 관리.
모든 테스트 사용자 비밀번호: `password123`

| Handle | Display Name | 설명 |
|--------|-------------|------|
| alice | Alice Kim | 메인 데모 사용자 |
| bob | Bob Park | 디자이너 |
| charlie | Charlie Lee | Go 개발자 |
| diana | Diana Choi | PM |
| eve | Eve Jung | 데이터 사이언티스트 |
