# Spec: 글쓰기 환경 커스터마이징 (마크다운 + 미디어 + 위치 + 투표)

- **작성일**: 2026-03-06
- **상태**: Draft

---

## 1. 개요

### 1.1 What

ComposeForm의 글쓰기 환경을 확장하여 네 가지 핵심 기능을 추가한다.

1. **마크다운 지원**: 글 작성 시 마크다운 문법(볼드, 이탤릭, 코드블록, 링크, 리스트 등)을 사용하고, 작성 중 실시간 프리뷰를 제공하며, PostCard/PostDetailPage에서 마크다운을 렌더링한다.
2. **미디어 업로드**: 이미지(최대 4장), 동영상(최대 1개), GIF(최대 1개)를 게시물에 첨부할 수 있으며, 업로드 진행률 표시와 미디어 그리드 레이아웃을 제공한다.
3. **GPS 위치 태그**: 게시물에 위치 정보를 태그할 수 있으며, 위치명이 게시물에 표시된다.
4. **투표(Polls)**: 게시물에 투표를 첨부할 수 있으며, 2~4개의 선택지 + 투표 기간을 설정할 수 있다.

### 1.2 Why

- 현재 플레인 텍스트만 지원하여 코드 공유, 강조, 링크 첨부 등 표현력이 제한적임
- 미디어 첨부 없이는 SNS로서의 핵심 사용자 경험이 불완전함
- 위치 태그와 투표는 사용자 참여도를 높이는 핵심 인터랙션 기능
- X(Twitter)의 핵심 기능 중 하나로, 클론 프로젝트의 완성도를 위해 필수

### 1.3 핵심 제약

- 마크다운은 제한된 서브셋만 허용 (HTML 직접 입력 차단)
- 미디어 업로드 시 파일 크기 제한 적용 (이미지 5MB, 동영상 50MB, GIF 15MB)
- 동영상과 이미지/GIF는 동시 첨부 불가
- XSS 방지를 위한 서버 사이드 sanitization 필수

---

## 2. 설계 결정 사항

### 2.1 글자 수 제한 기준

**결정**: 마크다운 문법 기호를 **포함한 원본 텍스트** 기준으로 계산한다. 단, 마크다운 도입에 따라 글자 수 제한을 **280자에서 500자로 상향**한다.

**이유**:
- 문법 기호 제외 시 클라이언트-서버 간 계산 불일치 위험 (파싱 구현체 차이)
- X(Twitter)도 URL 등을 특별 처리하지만, 구현 복잡도 대비 이득이 적음
- 280자에 마크다운 기호까지 포함하면 실질적 표현 공간이 너무 좁아지므로 500자로 상향
- DB 컬럼은 VARCHAR(280) -> TEXT로 변경, 서비스 레이어에서 500자 검증

**트레이드오프**: 마크다운 기호가 글자 수에 포함되어 순수 텍스트 대비 작성 가능 분량이 줄어들 수 있으나, 500자로 상향하여 완화

### 2.2 미디어 스토리지 전략

**결정**: 스토리지 인터페이스를 추상화하고, 초기 구현은 **로컬 파일 시스템**으로 한다.

**이유**:
- 학습 목적의 클론 프로젝트로 AWS 의존성 없이 빠르게 구현
- Go 인터페이스로 `MediaStorage`를 정의하여 S3 전환 시 구현체만 교체
- 로컬 저장 경로: `./uploads/{year}/{month}/{uuid}.{ext}`

**트레이드오프**: 프로덕션 환경에서는 S3/CloudFront가 필수이나, 개발 단계에서는 로컬로 충분

### 2.3 마크다운 허용 범위

**결정**: 제한된 서브셋만 허용한다.

| 허용 | 불허 |
|------|------|
| 볼드 (`**bold**`) | HTML 태그 직접 입력 |
| 이탤릭 (`*italic*`) | 이미지 문법 (`![](url)`) |
| 인라인 코드 (`` `code` ``) | 제목 (`#`, `##` 등) |
| 코드 블록 (` ``` `) | 테이블 |
| 링크 (`[text](url)`) | 각주 |
| 순서 있는/없는 리스트 | 수평선 (`---`) |
| 취소선 (`~~text~~`) | |
| 인용 (`> quote`) | |

**이유**: SNS 게시물 특성상 문서 수준의 마크다운은 불필요하며, 공격 표면을 최소화

---

## 3. 기능 요구사항

### 3.1 Feature 1: 마크다운 지원

#### 수락 기준

- [ ] ComposeForm에서 마크다운 문법을 입력할 수 있다
- [ ] "작성" / "미리보기" 탭 전환으로 렌더링 결과를 확인할 수 있다
- [ ] 미리보기 탭에서 마크다운이 HTML로 렌더링된다
- [ ] PostCard에서 게시물 content가 마크다운으로 렌더링된다
- [ ] PostDetailPage에서도 마크다운이 렌더링된다
- [ ] 허용되지 않은 HTML 태그는 제거된다 (XSS 방지)
- [ ] 코드블록에 구문 강조(syntax highlighting)가 적용된다
- [ ] 링크는 `target="_blank"` + `rel="noopener noreferrer"`로 렌더링된다
- [ ] 글자 수 카운터는 원본 텍스트(마크다운 기호 포함) 기준으로 동작한다
- [ ] 글자 수 제한이 500자로 변경된다
- [ ] 빈 마크다운(공백/줄바꿈만 있는 경우)은 게시 불가

#### 엣지 케이스

1. **악의적 마크다운**: `[Click](javascript:alert(1))` -> javascript: 프로토콜 링크 차단
2. **중첩 마크다운**: `***bold italic***` -> 정상 렌더링
3. **불완전한 마크다운**: 닫히지 않은 코드블록 -> 원본 텍스트 그대로 표시
4. **극단적 길이 코드블록**: 코드블록 내 텍스트도 500자 제한에 포함
5. **XSS 페이로드**: `<script>alert(1)</script>` -> 태그 완전 제거
6. **기존 게시물 호환**: 마크다운 이전에 작성된 플레인 텍스트 게시물도 정상 표시 (마크다운 파서에 플레인 텍스트를 넣어도 동일하게 출력)

### 3.2 Feature 2: 미디어 업로드

#### 수락 기준

- [ ] ComposeForm 하단에 미디어 첨부 버튼(이미지, GIF)이 표시된다
- [ ] 이미지: jpg, png, webp, gif 파일을 최대 4장까지 첨부 가능
- [ ] 동영상: mp4, webm 파일을 최대 1개 첨부 가능
- [ ] GIF: gif 파일 1개만 첨부 가능 (애니메이션 GIF 전용)
- [ ] 동영상 첨부 시 이미지/GIF 첨부 불가 (역도 동일)
- [ ] GIF 첨부 시 이미지/동영상 첨부 불가 (역도 동일)
- [ ] 첨부된 미디어의 썸네일 미리보기가 ComposeForm에 표시된다
- [ ] 각 미디어 미리보기에 삭제(X) 버튼이 있다
- [ ] 업로드 진행률이 프로그레스 바로 표시된다
- [ ] 파일 크기 초과 시 에러 토스트가 표시된다
- [ ] 허용되지 않은 파일 형식 선택 시 에러 토스트가 표시된다
- [ ] PostCard에서 미디어가 그리드 레이아웃으로 표시된다 (1장: 전체, 2장: 2열, 3장: 1+2, 4장: 2x2)
- [ ] PostDetailPage에서 미디어가 표시된다
- [ ] 미디어 클릭 시 라이트박스(전체 화면 뷰)로 확대된다
- [ ] 미디어 없이 텍스트만으로도 게시 가능 (기존 동작 유지)
- [ ] 텍스트 없이 미디어만으로도 게시 가능

#### 엣지 케이스

1. **대용량 파일**: 제한 초과 파일은 클라이언트에서 선택 시점에 즉시 거부
2. **업로드 중 취소**: 업로드 진행 중 미디어 삭제 버튼 클릭 시 업로드 중단
3. **업로드 중 게시**: 업로드 완료 전에는 게시 버튼 비활성화
4. **네트워크 오류**: 업로드 실패 시 재시도 버튼 표시
5. **동시 업로드**: 여러 이미지를 동시에 선택해도 개별 진행률 표시
6. **MIME 타입 스푸핑**: 서버에서 매직 바이트 검증으로 실제 파일 형식 확인
7. **매우 큰 이미지**: 서버에서 리사이즈하지 않되, 최대 해상도 제한 (4096x4096)
8. **동영상 길이**: 최대 2분 20초 제한 (X와 동일)
9. **답글에서의 미디어**: 답글(Reply)에서도 미디어 첨부 가능

### 3.3 Feature 3: GPS 위치 태그

#### 수락 기준

- [ ] ComposeForm 하단 툴바에 위치 태그 버튼(MapPin 아이콘)이 표시된다
- [ ] 버튼 클릭 시 브라우저 Geolocation API로 현재 위치를 가져온다
- [ ] 위치 권한 요청 다이얼로그가 표시된다
- [ ] 위치를 가져오면 역지오코딩(Reverse Geocoding)으로 장소명을 표시한다
- [ ] 위치명은 ComposeForm에서 태그 형태로 표시되며, X 버튼으로 제거 가능
- [ ] 사용자가 위치명을 직접 수정(커스텀 텍스트)할 수 있다
- [ ] PostCard에서 위치 정보가 작성자명 아래에 MapPin 아이콘 + 위치명으로 표시된다
- [ ] PostDetailPage에서도 위치 정보가 표시된다
- [ ] 위치 정보는 선택사항 (없어도 게시 가능)

#### 엣지 케이스

1. **위치 권한 거부**: 토스트로 "위치 접근이 거부되었습니다" 안내, 수동 입력 유도
2. **Geolocation 불가**: HTTPS가 아닌 환경 등에서 API 사용 불가 시 버튼 비활성화
3. **역지오코딩 실패**: 좌표는 저장하되 위치명은 "알 수 없는 위치"로 표시
4. **매우 정밀한 좌표**: 프라이버시를 위해 소수점 2자리(~1km 정밀도)로 라운딩
5. **위치명 길이**: 최대 100자 제한

### 3.4 Feature 4: 투표(Polls)

#### 수락 기준

- [ ] ComposeForm 하단 툴바에 투표 생성 버튼(BarChart 아이콘)이 표시된다
- [ ] 버튼 클릭 시 투표 옵션 입력 UI가 ComposeForm 내에 표시된다
- [ ] 기본 2개 옵션 입력 필드 + "선택지 추가" 버튼으로 최대 4개까지 추가
- [ ] 각 옵션은 최대 25자
- [ ] 투표 기간 설정: 1시간, 6시간, 12시간, 1일, 3일, 7일 중 선택 (기본: 1일)
- [ ] 투표가 첨부된 게시물에는 미디어 첨부 불가 (역도 동일)
- [ ] PostCard에서 투표 UI가 표시된다 (선택지 버튼 리스트)
- [ ] 투표 전: 각 선택지가 클릭 가능한 버튼으로 표시
- [ ] 투표 후: 각 선택지의 득표율(프로그레스 바 + 퍼센트)이 표시되고 내 선택에 체크 표시
- [ ] 투표 종료 후: "최종 결과"로 표시, 더 이상 투표 불가
- [ ] 총 투표 수 + 남은 시간이 표시된다
- [ ] 본인 게시물의 투표에는 투표할 수 없다 (결과만 확인 가능)
- [ ] 한 번 투표하면 변경 불가

#### 엣지 케이스

1. **중복 투표**: 서버에서 유저당 1회 투표만 허용 (UNIQUE 제약조건)
2. **투표 기간 만료 체크**: 서버 시간 기준으로 만료 여부 판단 (클라이언트 시간 신뢰 불가)
3. **빈 옵션**: 최소 2개 옵션에 텍스트가 있어야 게시 가능
4. **동시 투표**: 동시성 제어를 위해 DB 레벨에서 원자적 카운트 증가
5. **투표 + 답글**: 답글에는 투표를 첨부할 수 없다 (최상위 게시물만 가능)
6. **미인증 사용자**: 투표 결과는 볼 수 있으나, 투표 참여는 로그인 필요

---

## 4. API 설계

### 4.1 미디어 업로드

```
POST /api/media/upload
Authorization: Bearer {token}
Content-Type: multipart/form-data

Form Fields:
  file: (binary)

Response 201:
{
  "data": {
    "id": "uuid",
    "url": "/uploads/2026/03/uuid.jpg",
    "type": "image",          // "image" | "video" | "gif"
    "mimeType": "image/jpeg",
    "width": 1200,
    "height": 800,
    "size": 245000,           // bytes
    "duration": null           // seconds, 동영상만 해당
  }
}

Error 400:
{
  "error": {
    "code": "INVALID_FILE_TYPE",
    "message": "허용되지 않은 파일 형식입니다. (jpg, png, webp, gif, mp4, webm)"
  }
}

Error 413:
{
  "error": {
    "code": "FILE_TOO_LARGE",
    "message": "파일 크기가 제한을 초과했습니다. (이미지: 5MB, 동영상: 50MB, GIF: 15MB)"
  }
}
```

### 4.2 게시물 생성 (수정)

기존 `POST /api/posts` 요청에 `mediaIds` 필드를 추가한다.

```
POST /api/posts
Authorization: Bearer {token}
Content-Type: application/json

Request Body:
{
  "content": "**Hello** world! `code`",   // 마크다운 원본, 최대 500자
  "visibility": "public",
  "mediaIds": ["uuid-1", "uuid-2"]         // optional, 최대 4개 (이미지) 또는 1개 (동영상/GIF)
}

Response 201:
{
  "data": {
    "id": "uuid",
    "authorId": "uuid",
    "content": "**Hello** world! `code`",
    "visibility": "public",
    "media": [
      {
        "id": "uuid-1",
        "url": "/uploads/2026/03/uuid-1.jpg",
        "type": "image",
        "width": 1200,
        "height": 800
      }
    ],
    "createdAt": "2026-03-06T12:00:00Z",
    "updatedAt": "2026-03-06T12:00:00Z"
  }
}

Error 400 (미디어 조합 위반):
{
  "error": {
    "code": "INVALID_MEDIA_COMBINATION",
    "message": "동영상/GIF는 다른 미디어와 함께 첨부할 수 없습니다."
  }
}

Error 400 (텍스트 + 미디어 모두 없음):
{
  "error": {
    "code": "EMPTY_POST",
    "message": "텍스트 또는 미디어 중 하나는 반드시 포함해야 합니다."
  }
}
```

### 4.3 게시물 조회 응답 변경

기존 `PostDetailResponse`에 `media` 배열 필드를 추가한다.

```
GET /api/posts/:id
GET /api/feed

PostDetailResponse에 추가되는 필드:
{
  ...기존 필드,
  "media": [
    {
      "id": "uuid",
      "url": "/uploads/2026/03/uuid.jpg",
      "type": "image",
      "width": 1200,
      "height": 800,
      "duration": null
    }
  ]
}
```

### 4.4 답글 생성 (수정)

기존 `POST /api/posts/:id/replies` 요청에도 `mediaIds` 필드를 추가한다.

```
POST /api/posts/:id/replies
Authorization: Bearer {token}

Request Body:
{
  "content": "reply with *markdown*",
  "mediaIds": ["uuid-1"]                  // optional
}
```

### 4.5 정적 파일 서빙

```
GET /uploads/{year}/{month}/{filename}

설명: Fiber의 Static 미들웨어로 ./uploads 디렉토리를 서빙한다.
캐싱: Cache-Control: public, max-age=31536000, immutable (파일명이 UUID이므로 불변)
```

### 4.6 게시물 생성 (위치 + 투표 필드 추가)

기존 `POST /api/posts` 요청에 `location`과 `poll` 필드를 추가한다.

```
POST /api/posts
Authorization: Bearer {token}

Request Body:
{
  "content": "서울에서 점심 뭐 먹을까?",
  "visibility": "public",
  "mediaIds": [],
  "location": {                           // optional
    "latitude": 37.57,
    "longitude": 126.98,
    "name": "서울특별시 종로구"              // optional, 역지오코딩 결과 or 사용자 입력
  },
  "poll": {                               // optional
    "options": ["한식", "중식", "양식"],     // 2~4개
    "durationMinutes": 1440               // 투표 기간 (분 단위)
  }
}

검증:
- poll과 mediaIds는 동시 사용 불가
- poll은 답글(parentId 있는 경우)에서 사용 불가
- poll.options: 최소 2개, 최대 4개, 각 최대 25자
- poll.durationMinutes: 60(1시간) ~ 10080(7일) 범위
- location.name: 최대 100자
- location.latitude: -90 ~ 90
- location.longitude: -180 ~ 180
```

### 4.7 투표 참여

```
POST /api/posts/:id/vote
Authorization: Bearer {token}

Request Body:
{
  "optionIndex": 0    // 0-based index
}

Response 200:
{
  "data": {
    "poll": {
      "options": [
        { "text": "한식", "voteCount": 15 },
        { "text": "중식", "voteCount": 8 },
        { "text": "양식", "voteCount": 12 }
      ],
      "totalVotes": 35,
      "votedIndex": 0,
      "expiresAt": "2026-03-07T12:00:00Z",
      "isExpired": false
    }
  }
}

Error 400: 이미 투표함 / 만료됨 / 본인 게시물
Error 404: 게시물 없음 / 투표 없음
```

### 4.8 게시물 조회 응답 (위치 + 투표 추가)

```
PostDetailResponse에 추가되는 필드:
{
  ...기존 필드,
  "media": [...],
  "location": {
    "latitude": 37.57,
    "longitude": 126.98,
    "name": "서울특별시 종로구"
  },
  "poll": {
    "options": [
      { "text": "한식", "voteCount": 15 },
      { "text": "중식", "voteCount": 8 },
      { "text": "양식", "voteCount": 12 }
    ],
    "totalVotes": 35,
    "votedIndex": 0,         // 현재 유저가 투표한 인덱스 (미투표 시 -1)
    "expiresAt": "2026-03-07T12:00:00Z",
    "isExpired": false
  }
}
```

---

## 5. DB 스키마

### 5.1 posts 테이블 변경

```sql
-- Migration: 008_alter_posts_content_to_text.up.sql
ALTER TABLE posts ALTER COLUMN content TYPE TEXT;
-- 기존 VARCHAR(280) -> TEXT 변경
-- 500자 제한은 서비스 레이어에서 검증

-- Migration: 008_alter_posts_content_to_text.down.sql
ALTER TABLE posts ALTER COLUMN content TYPE VARCHAR(280);
```

### 5.2 post_media 테이블 추가

```sql
-- Migration: 009_create_post_media.up.sql
CREATE TABLE post_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    media_type VARCHAR(10) NOT NULL,        -- 'image', 'video', 'gif'
    mime_type VARCHAR(50) NOT NULL,
    width INT,
    height INT,
    size_bytes BIGINT NOT NULL,
    duration_seconds FLOAT,                  -- 동영상만 해당
    sort_order SMALLINT NOT NULL DEFAULT 0,  -- 미디어 순서 보존
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_post_media_post_id ON post_media(post_id);
CREATE INDEX idx_post_media_uploader_id ON post_media(uploader_id);

-- post_id가 NULL인 경우: 업로드는 완료되었으나 아직 게시물에 연결되지 않은 상태
-- 고아 미디어 정리를 위한 배치 작업 고려 필요

-- Migration: 009_create_post_media.down.sql
DROP TABLE IF EXISTS post_media;
```

### 5.3 posts 테이블에 위치 컬럼 추가

```sql
-- Migration: 010_add_location_to_posts.up.sql
ALTER TABLE posts
  ADD COLUMN location_lat DOUBLE PRECISION,
  ADD COLUMN location_lng DOUBLE PRECISION,
  ADD COLUMN location_name VARCHAR(100);

-- Migration: 010_add_location_to_posts.down.sql
ALTER TABLE posts
  DROP COLUMN location_lat,
  DROP COLUMN location_lng,
  DROP COLUMN location_name;
```

### 5.4 polls 테이블 추가

```sql
-- Migration: 011_create_polls.up.sql
CREATE TABLE polls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL UNIQUE REFERENCES posts(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    total_votes INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE poll_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    option_index SMALLINT NOT NULL,
    text VARCHAR(25) NOT NULL,
    vote_count INT NOT NULL DEFAULT 0,
    UNIQUE(poll_id, option_index)
);

CREATE TABLE poll_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option_index SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(poll_id, user_id)  -- 유저당 1회 투표만 허용
);

CREATE INDEX idx_polls_post_id ON polls(post_id);
CREATE INDEX idx_poll_options_poll_id ON poll_options(poll_id);
CREATE INDEX idx_poll_votes_poll_id ON poll_votes(poll_id);
CREATE INDEX idx_poll_votes_user_id ON poll_votes(user_id);

-- Migration: 011_create_polls.down.sql
DROP TABLE IF EXISTS poll_votes;
DROP TABLE IF EXISTS poll_options;
DROP TABLE IF EXISTS polls;
```

### 5.5 데이터 모델 관계

```
posts (1) --- (N) post_media
  - post_media.post_id -> posts.id (nullable, 업로드 후 게시 전 상태)
  - post_media.uploader_id -> users.id (업로드한 사용자)

posts (1) --- (0..1) polls
  - polls.post_id -> posts.id (1:1 관계, UNIQUE 제약)

polls (1) --- (N) poll_options
  - poll_options.poll_id -> polls.id

polls (1) --- (N) poll_votes
  - poll_votes.poll_id -> polls.id
  - poll_votes.user_id -> users.id
  - UNIQUE(poll_id, user_id) 로 중복 투표 방지
```

---

## 6. 백엔드 구현 설계

### 6.1 레이어 구조

```
handler/
  media_handler.go        -- 미디어 업로드 핸들러
  poll_handler.go         -- 투표 참여 핸들러

service/
  media_service.go        -- 미디어 비즈니스 로직 (검증, 저장 위임)
  poll_service.go         -- 투표 비즈니스 로직 (생성, 투표, 만료 체크)
  post_service.go         -- 수정: CreatePost에 mediaIds, location, poll 처리

repository/
  media_repository.go     -- post_media CRUD
  poll_repository.go      -- polls, poll_options, poll_votes CRUD

storage/
  storage.go              -- MediaStorage 인터페이스 정의
  local_storage.go        -- 로컬 파일 시스템 구현체
  # s3_storage.go         -- 향후 S3 구현체
```

### 6.2 MediaStorage 인터페이스

```go
type MediaStorage interface {
    Upload(ctx context.Context, file io.Reader, filename string, contentType string) (url string, err error)
    Delete(ctx context.Context, url string) error
}
```

### 6.3 미디어 검증 규칙

| 항목 | 이미지 | 동영상 | GIF |
|------|--------|--------|-----|
| 최대 파일 크기 | 5 MB | 50 MB | 15 MB |
| 허용 MIME | image/jpeg, image/png, image/webp | video/mp4, video/webm | image/gif |
| 최대 개수 | 4 | 1 | 1 |
| 최대 해상도 | 4096x4096 | - | 4096x4096 |
| 최대 길이 | - | 140초 | - |
| 다른 타입과 혼합 | 불가(동영상/GIF) | 불가 | 불가 |

### 6.4 마크다운 처리

- **저장**: 원본 마크다운 텍스트를 그대로 DB에 저장
- **렌더링**: 프론트엔드에서 렌더링 (서버는 원본 반환)
- **검증**: 서비스 레이어에서 500자 제한 검증 (원본 텍스트 기준)
- **sanitization**: 프론트엔드 렌더링 시 DOMPurify로 HTML sanitize

---

## 7. 프론트엔드 컴포넌트 설계

### 7.1 수정 대상 컴포넌트

#### ComposeForm.tsx 변경사항

- 글자 수 제한: 280 -> 500
- "작성" / "미리보기" 탭 추가
- 미리보기 탭에서 마크다운 렌더링
- 하단 툴바에 미디어 첨부 버튼 추가
- 미디어 미리보기 영역 추가

#### PostCard.tsx 변경사항

- `<p>{post.content}</p>` -> `<MarkdownRenderer content={post.content} />`
- 미디어 그리드 컴포넌트 추가 (content 아래)

#### PostDetailPage.tsx 변경사항

- 마크다운 렌더링 적용
- 미디어 표시 추가

### 7.2 새 컴포넌트

```
components/
  MarkdownRenderer.tsx     -- 마크다운 -> sanitized HTML 렌더링
  MediaGrid.tsx            -- 1~4장 미디어 그리드 레이아웃
  MediaUploadButton.tsx    -- 미디어 첨부 버튼 + 파일 선택
  MediaPreview.tsx         -- 첨부된 미디어 썸네일 + 삭제 버튼
  MediaLightbox.tsx        -- 미디어 전체 화면 뷰
  UploadProgressBar.tsx    -- 업로드 진행률 표시
  LocationTag.tsx          -- 위치 태그 표시 (MapPin + 위치명)
  LocationPicker.tsx       -- ComposeForm 내 위치 선택/수정 UI
  PollCreator.tsx          -- ComposeForm 내 투표 생성 UI (옵션 입력 + 기간 선택)
  PollDisplay.tsx          -- PostCard 내 투표 표시 (투표 전/후/만료 상태)

hooks/
  useMediaUpload.ts        -- 미디어 업로드 mutation + 진행률 관리
  useGeolocation.ts        -- 브라우저 Geolocation API + 역지오코딩
  usePoll.ts               -- 투표 참여 mutation + 투표 상태 관리
```

### 7.3 라이브러리 선택

| 용도 | 라이브러리 | 이유 |
|------|-----------|------|
| 마크다운 파싱 | `react-markdown` | React 컴포넌트 기반, remark/rehype 플러그인 생태계 |
| 코드 하이라이팅 | `rehype-highlight` 또는 `react-syntax-highlighter` | remark 파이프라인과 통합 용이 |
| HTML sanitization | `rehype-sanitize` | rehype 파이프라인 내에서 sanitize, DOMPurify 대비 번들 크기 절약 |
| GFM 지원 | `remark-gfm` | 취소선, 테이블 등 GitHub Flavored Markdown |

### 7.4 MediaGrid 레이아웃 규칙

```
1장: 단일 이미지, 최대 높이 제한, rounded-2xl
2장: 2열 균등 분할 (gap-0.5)
3장: 좌측 1장 (2/3), 우측 2장 세로 스택 (1/3)
4장: 2x2 그리드 (gap-0.5)
동영상: 단일, 16:9 비율, 재생 컨트롤 표시
GIF: 단일, 자동 재생, 루프
```

### 7.5 업로드 플로우 (프론트엔드)

1. 사용자가 미디어 첨부 버튼 클릭 -> 파일 선택 다이얼로그
2. 파일 선택 -> 클라이언트 검증 (파일 크기, MIME 타입, 개수 제한)
3. 검증 통과 -> `POST /api/media/upload`로 업로드 시작
4. XMLHttpRequest의 `onprogress` 이벤트로 진행률 표시
5. 업로드 완료 -> mediaId를 로컬 상태에 저장 + 썸네일 미리보기 표시
6. 게시 버튼 클릭 -> `POST /api/posts`에 content + mediaIds 전송
7. 모든 미디어 업로드 완료 전까지 게시 버튼 비활성화

---

## 8. 보안 고려사항

### 8.1 XSS 방지

- 서버: 마크다운 원본을 그대로 저장, HTML 변환하지 않음
- 클라이언트: `rehype-sanitize`로 허용된 태그/속성만 통과
- `javascript:`, `data:`, `vbscript:` 프로토콜 링크 차단
- `on*` 이벤트 핸들러 속성 제거

### 8.2 파일 업로드 보안

- **MIME 타입 검증**: Content-Type 헤더 + 매직 바이트(파일 시그니처) 이중 검증
- **파일명 무시**: 원본 파일명을 사용하지 않고 UUID로 대체 (경로 순회 공격 방지)
- **저장 경로 격리**: 업로드 디렉토리는 Go 소스 코드 디렉토리 외부에 위치
- **실행 권한 제거**: 업로드된 파일에 실행 권한을 부여하지 않음 (0644)
- **크기 제한**: Fiber의 `BodyLimit` 미들웨어로 요청 본문 크기 제한

### 8.3 권한 검증

- 업로드된 미디어를 게시물에 연결할 때 `uploader_id == 요청자 ID` 검증
- 타인이 업로드한 미디어를 자신의 게시물에 첨부할 수 없음
- 고아 미디어(게시물에 연결되지 않은 업로드)는 24시간 후 배치 삭제 예정

### 8.4 Rate Limiting

- 미디어 업로드: 사용자당 분당 10회 제한
- 게시물 생성: 기존 제한 유지

---

## 9. 의존성 및 제약사항

### 9.1 백엔드 의존성

- 새 Go 패키지 불필요 (파일 I/O는 표준 라이브러리로 충분)
- 동영상 메타데이터(길이, 해상도) 추출이 필요하면 `ffprobe` 또는 Go 라이브러리 검토 필요

### 9.2 프론트엔드 의존성 (신규)

- `react-markdown` (마크다운 렌더링)
- `remark-gfm` (GFM 지원)
- `rehype-sanitize` (XSS 방지)
- `rehype-highlight` 또는 `react-syntax-highlighter` (코드 하이라이팅)

### 9.3 기존 코드 영향 범위

| 파일 | 변경 내용 |
|------|----------|
| `backend/internal/dto/post_dto.go` | CreatePostRequest에 MediaIds/Location/Poll 추가, PostDetailResponse에 Media/Location/Poll 추가, 글자 수 제한 280->500 |
| `backend/internal/model/post.go` | Post에 위치 필드 추가, PostWithAuthor에 Media/Poll 추가 |
| `backend/internal/service/post_service.go` | CreatePost에 미디어/위치/투표 처리, 글자 수 검증 500자 |
| `backend/internal/repository/post_repository.go` | 게시물 조회 시 post_media/polls JOIN, 위치 컬럼 |
| `backend/internal/handler/post_handler.go` | CreatePost 요청 파싱 변경 |
| `backend/main.go` | 미디어/투표 라우트 등록, 정적 파일 서빙 |
| `frontend/src/components/ComposeForm.tsx` | 마크다운 탭, 미디어/위치/투표 UI, 글자 수 500 |
| `frontend/src/components/PostCard.tsx` | MarkdownRenderer + MediaGrid + LocationTag + PollDisplay |
| `frontend/src/pages/PostDetailPage.tsx` | 마크다운/미디어/위치/투표 표시 |
| `frontend/src/types/api.ts` | Media/Location/Poll 타입, CreatePostRequest 수정 |
| `frontend/src/hooks/usePosts.ts` | createPost mutation 수정 |

---

## 10. 구현 우선순위

구현을 3단계로 나누어 점진적으로 진행한다.

### Phase A: 마크다운 지원 (예상 소요: 1일)

1. posts.content 컬럼 VARCHAR -> TEXT 마이그레이션
2. 글자 수 제한 280 -> 500 변경 (백엔드 DTO + 서비스, 프론트엔드 ComposeForm)
3. `react-markdown` + `remark-gfm` + `rehype-sanitize` 설치
4. `MarkdownRenderer` 컴포넌트 구현
5. ComposeForm에 "작성/미리보기" 탭 추가
6. PostCard, PostDetailPage에 MarkdownRenderer 적용
7. 코드 하이라이팅 추가

### Phase B: 미디어 업로드 백엔드 (예상 소요: 1~2일)

1. post_media 테이블 마이그레이션
2. MediaStorage 인터페이스 + LocalStorage 구현체
3. MediaRepository 구현
4. MediaService 구현 (파일 검증, 저장, 메타데이터 추출)
5. MediaHandler 구현 (업로드 API)
6. PostService 수정 (CreatePost에 mediaIds 처리)
7. PostRepository 수정 (조회 시 미디어 JOIN)
8. 정적 파일 서빙 설정

### Phase C: 미디어 업로드 프론트엔드 (예상 소요: 1~2일)

1. `useMediaUpload` 훅 구현 (업로드 + 진행률)
2. MediaUploadButton, MediaPreview 컴포넌트
3. ComposeForm에 미디어 UI 통합
4. MediaGrid 컴포넌트 (1~4장 레이아웃)
5. PostCard, PostDetailPage에 MediaGrid 적용
6. MediaLightbox (전체 화면 뷰)
7. API 타입 업데이트

### Phase D: GPS 위치 태그 (예상 소요: 0.5~1일)

1. posts 테이블에 위치 컬럼 마이그레이션
2. PostService CreatePost에 location 처리 추가
3. PostRepository 조회 쿼리에 위치 컬럼 추가
4. DTO/Model에 위치 필드 추가
5. `useGeolocation` 훅 구현 (Geolocation API + 역지오코딩)
6. LocationPicker 컴포넌트 (ComposeForm 통합)
7. LocationTag 컴포넌트 (PostCard/PostDetailPage 표시)

### Phase E: 투표(Polls) 백엔드 (예상 소요: 1일)

1. polls, poll_options, poll_votes 테이블 마이그레이션
2. PollRepository 구현 (생성, 투표, 조회)
3. PollService 구현 (투표 검증, 만료 체크, 결과 계산)
4. PollHandler 구현 (POST /posts/:id/vote)
5. PostService 수정 (CreatePost에 poll 처리, 조회 시 poll JOIN)

### Phase F: 투표(Polls) 프론트엔드 (예상 소요: 1일)

1. `usePoll` 훅 구현 (투표 mutation + optimistic UI)
2. PollCreator 컴포넌트 (옵션 입력 + 기간 선택)
3. ComposeForm에 투표 UI 통합 (미디어와 상호 배타)
4. PollDisplay 컴포넌트 (투표 전/후/만료 3가지 상태)
5. PostCard, PostDetailPage에 PollDisplay 적용
6. API 타입 업데이트

---

## 11. 테스트 계획

### 11.1 백엔드

- MediaStorage 인터페이스 mock으로 MediaService 단위 테스트
- 파일 검증 로직 테이블 드리븐 테스트 (허용/불허 MIME, 크기 초과, 개수 초과)
- 미디어 조합 검증 테스트 (이미지+동영상 동시 첨부 거부 등)
- PostService CreatePost 통합 테스트 (mediaIds 연결)

### 11.2 프론트엔드

- MarkdownRenderer: 각 마크다운 문법의 렌더링 검증
- MarkdownRenderer: XSS 페이로드 차단 검증
- MediaGrid: 1~4장 레이아웃 스냅샷 테스트
- ComposeForm: 미디어 첨부/삭제 상호작용 테스트
- PollCreator: 옵션 추가/삭제/제한 검증
- PollDisplay: 투표 전/후/만료 3가지 상태 렌더링 검증
- LocationPicker: Geolocation API 모킹 테스트

---

## 12. 미결정 사항

1. **동영상 트랜스코딩**: 현재 스펙에서는 업로드된 원본을 그대로 서빙. 프로덕션에서는 HLS 변환 필요하나 클론 프로젝트 범위에서는 제외. 추후 결정 필요.
2. **이미지 리사이즈/썸네일 생성**: 현재는 원본 서빙. 대역폭 최적화가 필요하면 Sharp(Node) 또는 imaging(Go) 라이브러리로 리사이즈 파이프라인 추가 검토.
3. **고아 미디어 정리**: 업로드 후 게시하지 않은 미디어의 정리 배치 작업. 구현 시점은 Phase B 이후 별도 이슈로 관리.
4. **Alt 텍스트**: 접근성을 위한 이미지 대체 텍스트 입력. 현재 스펙에서는 제외하나, 향후 추가 고려.
5. **역지오코딩 서비스**: 무료 API(Nominatim/OpenStreetMap) vs 유료 API(Google Maps). 클론 프로젝트이므로 Nominatim 우선, 프론트엔드에서 호출.
6. **투표 결과 실시간 업데이트**: WebSocket으로 투표 수 실시간 반영할지, 새로고침 시에만 업데이트할지. Phase 4(WebSocket)와 연계하여 결정.
