# Spec: 게시물 조회수(View Count) 기능

- **작성일**: 2026-03-06
- **상태**: Draft
- **마이그레이션 번호**: 013

---

## 1. 개요

### 1.1 What

모든 post/reply에 조회수(view count) 수치를 추가한다. 게시글 상세 페이지 진입 시 조회수가 1 증가하며, 피드 리스트/PostCard/ReplyCard/PostDetailPage 등 모든 UI에서 Eye 아이콘과 함께 숫자를 표시한다.

### 1.2 Why

- **사용자 참여도 지표**: X(Twitter)는 2022년 말부터 모든 트윗에 조회수를 공개 표시. 사용자가 콘텐츠의 도달 범위를 파악할 수 있는 핵심 메트릭
- **콘텐츠 가치 판단**: 좋아요/답글 수 대비 조회수를 통해 "암묵적 참여(lurking)" 정도를 파악 가능
- **기존 패턴 활용**: `like_count`, `reply_count`와 동일한 패턴으로 구현 가능하여 아키텍처 변경 최소화

### 1.3 핵심 제약

- `handler -> service -> repository` 레이어 구조 준수
- `ctx context.Context`가 모든 Go 메서드의 첫 번째 파라미터
- 조회수 증가는 **GetPostByID (상세 조회)에서만** 발생 (피드 리스트 조회에서는 증가 안 함)
- Visibility 접근 제어(`checkVisibilityAccess`)를 통과한 후에만 조회수 증가
- 본인 게시물 조회 시에도 조회수 증가 (X 동작과 동일)

---

## 2. 설계 결정 사항

### 2.1 조회수 증가 시점

**결정**: `PostService.GetPostByID` 내에서 visibility 접근 제어를 통과한 후, 응답 반환 전에 비동기적으로 조회수를 증가시킨다.

**이유**:
- 피드 리스트에서 증가시키면 스크롤만 해도 조회수가 폭증하여 의미 없는 수치가 됨
- 상세 페이지 진입 = 사용자가 해당 콘텐츠에 관심을 가졌다는 의미 있는 신호
- X(Twitter)도 트윗 상세 보기(impression) 기준으로 카운트

**트레이드오프**: 실시간 정확도보다 응답 속도를 우선. 카운트 증가 실패 시 에러를 무시하고 응답은 정상 반환

### 2.2 카운트 저장 방식

**결정**: `posts` 테이블에 `view_count INTEGER NOT NULL DEFAULT 0` 컬럼을 직접 추가. 별도의 view 기록 테이블은 생성하지 않음.

**이유**:
- `like_count`, `reply_count`와 동일한 패턴으로 일관성 유지
- 별도 테이블(post_views)을 두면 집계 쿼리 비용 증가, MVP 수준에서는 과도한 설계
- 고유 조회수(unique view) 추적은 현재 스코프 밖 (추후 Redis 기반 HyperLogLog 등으로 확장 가능)

**트레이드오프**: 중복 조회도 카운트되지만, X도 동일 사용자 재방문을 별도 impression으로 카운트하므로 문제 없음

### 2.3 동시성 처리

**결정**: `UPDATE posts SET view_count = view_count + 1 WHERE id = $1` — 원자적 증가 쿼리 사용

**이유**:
- PostgreSQL의 `view_count = view_count + 1`은 row-level lock으로 동시성 안전
- 별도 트랜잭션이나 mutex 불필요
- 트래픽이 극도로 높아질 경우 Redis 카운터 -> 주기적 flush로 변경 가능하나 현재 불필요

### 2.4 조회수 증가와 응답 순서

**결정**: 조회수 증가를 먼저 실행한 후 결과를 반환. 단, 증가 실패 시 에러를 로그만 남기고 정상 응답.

**이유**:
- 조회수 증가 실패가 사용자 경험을 해치면 안 됨
- 반환되는 `viewCount` 값은 증가 전 DB에서 읽어온 값이므로, 현재 요청의 조회수는 포함되지 않을 수 있음 (eventually consistent)
- 이는 X의 실제 동작과도 일치 (표시되는 조회수가 약간의 지연이 있음)

---

## 3. 수락 기준 (Acceptance Criteria)

### AC-1: DB 마이그레이션
- [ ] `013_add_view_count_to_posts.up.sql`이 `posts` 테이블에 `view_count INTEGER NOT NULL DEFAULT 0` 컬럼을 추가한다
- [ ] `013_add_view_count_to_posts.down.sql`이 해당 컬럼을 제거한다
- [ ] 기존 게시물의 view_count가 0으로 초기화된다

### AC-2: 백엔드 모델/DTO
- [ ] `model.Post` 구조체에 `ViewCount int` 필드가 존재한다
- [ ] `model.PostWithAuthor` 구조체에서 `Post`를 임베딩하므로 자동으로 `ViewCount` 접근 가능
- [ ] `dto.PostDetailResponse`에 `ViewCount int` 필드 (`json:"viewCount"`)가 존재한다
- [ ] `dto.ToPostDetailResponse`에서 `ViewCount`를 매핑한다

### AC-3: Repository
- [ ] 모든 SELECT 쿼리에서 `p.view_count`를 조회한다 (FindByID, FindByIDWithUser, FindAll, FindAllWithUser, FindRepliesByPostID, FindRepliesByPostIDWithUser, FindAuthorReplyByPostID, FindAuthorReplyByPostIDWithUser, FindByAuthorHandle, FindByAuthorHandleWithUser, FindRepliesByAuthorHandle, FindRepliesByAuthorHandleWithUser, FindLikedByUserHandle, FindLikedByUserHandleWithViewer)
- [ ] `scanPostRows`, `scanReplyWithParentRows` 헬퍼에서도 `view_count`를 스캔한다
- [ ] `PostRepository` 인터페이스에 `IncrementViewCount(ctx context.Context, id uuid.UUID) error` 메서드가 추가된다
- [ ] `IncrementViewCount` 구현: `UPDATE posts SET view_count = view_count + 1 WHERE id = $1`

### AC-4: Service
- [ ] `PostService.GetPostByID`에서 `checkVisibilityAccess` 통과 후 `IncrementViewCount`를 호출한다
- [ ] `IncrementViewCount` 실패 시 에러를 로그(slog)로 남기되, 사용자 응답에는 영향을 주지 않는다
- [ ] 피드 리스트 조회(`GetPosts`, `ListPostsByHandle`, `ListRepliesByHandle`, `ListLikedPostsByHandle`)에서는 `IncrementViewCount`를 호출하지 않는다

### AC-5: 프론트엔드 타입
- [ ] `PostDetail` 인터페이스에 `viewCount: number` 필드가 존재한다

### AC-6: PostCard UI
- [ ] 액션 버튼 영역에 Eye 아이콘 + 조회수 숫자가 표시된다
- [ ] 조회수가 0일 때는 숫자를 표시하지 않거나 빈 문자열을 표시한다 (like_count, reply_count와 동일 패턴)
- [ ] Eye 아이콘은 `lucide-react`의 `Eye` 컴포넌트를 사용한다
- [ ] 호버 시 배경색 변경 (primary/10)

### AC-7: PostDetailPage UI
- [ ] Stats 영역에 "N 조회" 통계가 표시된다 (like/reply count 옆)
- [ ] 조회수가 0보다 클 때만 표시

### AC-8: ReplyCard UI
- [ ] 기존 like/reply 버튼 옆에 Eye 아이콘 + 조회수가 표시된다

---

## 4. API 변경 사항

### 4.1 기존 엔드포인트 응답 변경

새로운 엔드포인트는 추가하지 않는다. 기존 엔드포인트의 응답에 `viewCount` 필드를 추가한다.

#### GET /api/posts/:id (게시글 상세)

**변경**: 응답에 `viewCount` 추가, 호출 시 view_count 1 증가

```json
{
  "data": {
    "id": "uuid",
    "authorId": "uuid",
    "content": "...",
    "visibility": "public",
    "viewCount": 142,
    "likeCount": 5,
    "replyCount": 3,
    "isLiked": false,
    "isBookmarked": false,
    "author": { ... },
    "topReplies": [
      {
        "id": "uuid",
        "viewCount": 87,
        ...
      }
    ],
    "createdAt": "...",
    "updatedAt": "..."
  }
}
```

#### GET /api/posts (피드 리스트)

**변경**: 각 게시글에 `viewCount` 추가 (증가 없음, 읽기만)

```json
{
  "data": [
    {
      "id": "uuid",
      "viewCount": 142,
      "likeCount": 5,
      ...
    }
  ]
}
```

#### GET /api/users/:handle/posts, /api/users/:handle/replies, /api/users/:handle/likes

**변경**: 각 게시글에 `viewCount` 추가 (증가 없음, 읽기만)

---

## 5. ReBAC 고려사항

### 5.1 조회수 증가 접근 제어

- `checkVisibilityAccess`를 통과한 사용자만 조회수가 증가되어야 한다
- **public**: 모든 사용자 (비로그인 포함) -> 조회수 증가
- **follower**: 팔로워 + 작성자 본인 -> 조회수 증가, 그 외는 404 반환 (증가 안 됨)
- **private**: 작성자 본인만 -> 조회수 증가, 그 외는 404 반환 (증가 안 됨)

### 5.2 조회수 표시 접근 제어

- 조회수 자체는 별도의 접근 제어가 불필요. 게시글을 볼 수 있으면 조회수도 볼 수 있음
- 피드 리스트에서 visibility 필터링이 이미 적용되므로, 표시되는 게시글의 조회수는 자연스럽게 접근 권한이 있는 것

---

## 6. 엣지 케이스 목록

| # | 케이스 | 예상 동작 |
|---|--------|-----------|
| 1 | 비로그인 사용자가 public 게시글 상세 조회 | view_count 증가, viewCount 표시 |
| 2 | 비로그인 사용자가 follower/private 게시글 조회 시도 | 404, view_count 증가 안 됨 |
| 3 | 작성자 본인이 자기 게시글 조회 | view_count 증가 (X 동작과 동일) |
| 4 | 같은 사용자가 새로고침으로 반복 조회 | 매번 view_count 증가 (unique view 아님) |
| 5 | 게시글 상세 페이지에서 답글의 viewCount 표시 | 답글의 view_count를 표시하지만, 부모 게시글 상세 진입 시 답글의 view_count는 증가하지 않음 |
| 6 | 답글 상세 조회 (답글 클릭으로 해당 답글의 detail 페이지 진입) | 해당 답글의 view_count 증가 |
| 7 | 동시 다발적 조회 (race condition) | `view_count = view_count + 1` 원자적 연산으로 안전 |
| 8 | IncrementViewCount DB 에러 | 에러 로그만 남기고, 게시글 응답은 정상 반환 |
| 9 | 기존 게시물 (마이그레이션 적용 후) | view_count = 0으로 표시 |
| 10 | viewCount가 매우 큰 수 (1M+) | 프론트엔드에서 포맷팅 고려 (1.2M 등) - 현재 스코프 밖, 추후 formatCompactNumber 유틸 추가 |

---

## 7. 구현 상세

### 7.1 DB 마이그레이션

**파일**: `backend/migrations/013_add_view_count_to_posts.up.sql`
```sql
ALTER TABLE posts ADD COLUMN view_count INTEGER NOT NULL DEFAULT 0;
```

**파일**: `backend/migrations/013_add_view_count_to_posts.down.sql`
```sql
ALTER TABLE posts DROP COLUMN view_count;
```

### 7.2 백엔드 변경 범위

#### model/post.go
- `Post` 구조체에 `ViewCount int` 필드 추가 (`LikeCount`, `ReplyCount` 바로 뒤)

#### dto/post_dto.go
- `PostDetailResponse`에 `ViewCount int` 필드 추가 (`json:"viewCount"`, `likeCount`/`replyCount` 옆)
- `ToPostDetailResponse`에서 `ViewCount: p.ViewCount` 매핑 추가

#### repository/post_repository.go
- **인터페이스에 추가**: `IncrementViewCount(ctx context.Context, id uuid.UUID) error`
- **구현 추가**: 단일 UPDATE 쿼리
- **모든 SELECT 쿼리 수정**: `p.view_count`를 SELECT/Scan 목록에 추가
  - 개별 쿼리 (FindByID, FindByIDWithUser, FindAuthorReplyByPostID, FindAuthorReplyByPostIDWithUser): 직접 수정
  - 리스트 쿼리 (FindAll, FindAllWithUser, FindRepliesByPostID, FindRepliesByPostIDWithUser): 직접 수정
  - 핸들러 헬퍼 (`scanPostRows`, `scanReplyWithParentRows`): scanArgs에 `&p.ViewCount` 추가
  - handle 기반 쿼리 6개: 헬퍼를 거치므로 자동 반영

**주의**: `view_count`는 SELECT 목록에서 `reply_count` 바로 뒤에 배치하여, 기존 Scan 호출의 인자 순서에 맞춰 `&p.ViewCount`를 `&p.ReplyCount` 다음에 추가해야 함

#### service/post_service.go
- `GetPostByID`에서 `checkVisibilityAccess` 통과 후, 응답 조립 전에 `IncrementViewCount` 호출
- 실패 시 `slog.Error("failed to increment view count", ...)` 로그만 남기고 진행

### 7.3 프론트엔드 변경 범위

#### types/api.ts
- `PostDetail` 인터페이스에 `viewCount: number` 추가

#### components/PostCard.tsx
- `lucide-react`에서 `Eye` 아이콘 import 추가
- 액션 버튼 영역에 Eye 버튼 추가 (Bookmark 버튼과 Share 버튼 사이)
- 표시 패턴: `{post.viewCount || ""}` (0이면 빈 문자열)

#### pages/PostDetailPage.tsx
- Stats 영역에 조회수 통계 추가: `<strong>{post.viewCount}</strong> 조회`
- 조건: `post.viewCount > 0` 일 때만 표시

#### components/ReplyCard.tsx
- `lucide-react`에서 `Eye` 아이콘 import 추가
- 기존 like/reply 버튼 옆에 Eye + 조회수 표시
- 클릭 이벤트 없음 (조회수 버튼은 비인터랙티브)

---

## 8. 의존성 및 제약사항

### 8.1 의존성
- `backend/migrations/012_rename_friends_to_follower.up.sql` 마이그레이션이 먼저 적용되어 있어야 함
- `lucide-react` 패키지의 `Eye` 아이콘 (이미 설치된 패키지 내 포함)
- `log/slog` 패키지 (이미 프로젝트에서 사용 중)

### 8.2 제약사항
- Cursor Pagination에는 영향 없음 (view_count는 정렬/커서 기준으로 사용하지 않음)
- Unique view 추적은 현재 스코프 밖 (별도 이슈로 분리 권장)
- 조회수 기반 "인기 게시물" 정렬은 현재 스코프 밖

### 8.3 테스트 요구사항
- **Repository 테스트**: `IncrementViewCount`가 view_count를 정확히 1 증가시키는지
- **Service 테스트**: `GetPostByID` 호출 시 `IncrementViewCount`가 호출되는지, 피드 조회 시 호출되지 않는지
- **Service 테스트**: `IncrementViewCount` 실패 시에도 정상 응답이 반환되는지

---

## 9. 구현 순서 권장

1. DB 마이그레이션 작성 및 적용
2. Model에 ViewCount 필드 추가
3. Repository: IncrementViewCount 메서드 추가 + 모든 SELECT 쿼리에 view_count 추가
4. DTO: PostDetailResponse에 viewCount 필드 + ToPostDetailResponse 매핑
5. Service: GetPostByID에서 IncrementViewCount 호출
6. Frontend types: PostDetail에 viewCount 추가
7. Frontend UI: PostCard, PostDetailPage, ReplyCard에 조회수 표시
8. 테스트 작성

---

## 10. 남은 의문점

1. **큰 숫자 포맷팅**: viewCount가 1000 이상일 때 "1.2K", "3.4M" 형태로 표시할지? -> 현재는 raw 숫자 표시, 추후 `formatCompactNumber` 유틸로 별도 처리 권장
2. **봇/크롤러 필터링**: User-Agent 기반 봇 필터링이 필요한지? -> 현재 스코프 밖
3. **Rate Limiting**: 같은 사용자의 반복 조회를 제한할지? -> 현재 스코프 밖, X도 제한하지 않음
