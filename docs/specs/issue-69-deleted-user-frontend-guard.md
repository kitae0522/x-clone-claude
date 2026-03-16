# Issue #69: 탈퇴 사용자 익명화 프론트엔드 가드 누락

- **Parent Issue**: #65 (탈퇴 사용자 게시글 작성자 익명 처리)
- **유형**: Bug Fix
- **우선순위**: High (UX 결함 + 잠재적 OP 뱃지 오표시)
- **대상 파일**:
  - `frontend/src/components/PostCard.tsx`
  - `frontend/src/components/ParentPostCard.tsx`
  - `frontend/src/components/ReplyCard.tsx`
  - `frontend/src/pages/PostDetailPage.tsx`

---

## 1. 기능 설명 (What)

탈퇴한 사용자(`isDeleted === true`)의 게시글이 피드/상세 페이지에 표시될 때, 일부 UI 요소에서 익명화 가드가 누락되어 있다. 구체적으로 세 가지 문제가 존재한다:

1. **PostCard "replying to" 링크**: 탈퇴 사용자의 username이 클릭 가능한 파란색 링크로 렌더링됨
2. **ParentPostCard avatar 클릭**: avatar 영역에 isDeleted 체크 없이 프로필 네비게이션 발생
3. **ReplyCard/PostDetailPage OP 뱃지 false positive**: username 기반 비교로 인해 서로 다른 탈퇴 사용자에게 모두 OP 뱃지가 표시됨

## 2. 구현 이유 (Why)

- 탈퇴 사용자 프로필 페이지는 존재하지 않으므로, 클릭 시 404 또는 오류 상태로 이동하게 되어 UX가 깨진다.
- 탈퇴 사용자의 username은 모두 "deleted"로 설정되므로, username 기반 OP 비교는 모든 탈퇴 사용자를 동일인으로 판별하는 논리적 오류를 만든다.
- #65에서 백엔드 익명화를 도입했으나 프론트엔드 일부 분기에서 가드가 빠져 있어 완전한 익명화가 달성되지 않은 상태다.

---

## 3. 수락 기준 (Acceptance Criteria)

| # | 기준 | 검증 방법 |
|---|------|-----------|
| AC-1 | PostCard에서 `post.parent.author.isDeleted === true`이면 "replying to @deleted" 텍스트가 `text-muted-foreground`로 렌더링되며 클릭/hover 인터랙션이 없다 | 탈퇴 사용자의 답글 카드 육안 확인 |
| AC-2 | ParentPostCard에서 `post.author.isDeleted === true`이면 avatar 영역 클릭 시 프로필로 이동하지 않으며, cursor가 default이다 | 탈퇴 사용자 부모 게시글 avatar 클릭 확인 |
| AC-3 | PostDetailPage에서 OP 판별이 `authorId` 기반으로 수행되어, 서로 다른 탈퇴 사용자 답글에 OP 뱃지가 표시되지 않는다 | 2명 이상의 탈퇴 사용자가 답글을 달았을 때 OP 뱃지 확인 |
| AC-4 | `bun run check` (typecheck + lint) 통과 | CI |
| AC-5 | 기존 isDeleted 가드가 적용된 영역(PostCard author row, ReplyCard displayName 등)의 동작이 변경되지 않는다 | 회귀 테스트 |

---

## 4. 컴포넌트별 변경 상세

### 4.1 PostCard.tsx — "replying to" 링크 가드

**현재 코드** (라인 237-255):
```tsx
{post.parent && (
  <div
    className="mt-0.5 flex items-center gap-1 text-[13px] text-muted-foreground"
    onClick={(e) => {
      e.stopPropagation();
      navigate(`/post/${post.parent!.id}`);
    }}
  >
    <span>
      <span className="text-muted-foreground">replying to </span>
      <span className="cursor-pointer text-primary hover:underline">
        @{post.parent.author.username}
      </span>
    </span>
    ...
  </div>
)}
```

**문제**: `post.parent.author.isDeleted` 여부와 관계없이 항상 `text-primary hover:underline cursor-pointer` 스타일이 적용된다.

**변경 내용**:
- `@{post.parent.author.username}` 스팬에 조건부 스타일 적용
- `isDeleted === true`일 때: `text-muted-foreground` (기존 텍스트와 동일 톤), cursor/hover/underline 제거
- `isDeleted === false`일 때: 기존 `text-primary hover:underline cursor-pointer` 유지

변경할 부분 (라인 247-249):
```tsx
// Before
<span className="cursor-pointer text-primary hover:underline">
  @{post.parent.author.username}
</span>

// After
<span
  className={cn(
    post.parent.author.isDeleted
      ? "text-muted-foreground"
      : "cursor-pointer text-primary hover:underline",
  )}
>
  @{post.parent.author.username}
</span>
```

### 4.2 ParentPostCard.tsx — avatar 클릭 가드

**현재 코드** (라인 20-32):
```tsx
<div className="flex gap-3">
  <div className="flex flex-col items-center">
    {post.author.profileImageUrl ? (
      <img
        src={post.author.profileImageUrl}
        alt=""
        className="h-10 w-10 rounded-full object-cover"
      />
    ) : (
      <div className="h-10 w-10 rounded-full bg-border" />
    )}
    <div className="mt-1 w-0.5 flex-1 bg-border" />
  </div>
```

**문제**: avatar 영역(`<img>` / `<div>`)을 감싸는 클릭 핸들러가 없지만, 부모 `<div>` 전체에 `cursor-pointer`가 있고, avatar를 클릭하면 카드 전체 클릭 이벤트(`navigate(/post/${post.id})`)가 발생한다. 그러나 displayName span(라인 35-49)에는 이미 isDeleted 가드가 적용되어 있으므로, 일관성을 위해 avatar에도 독립적인 클릭 핸들러 + isDeleted 가드를 추가해야 한다.

**변경 내용**:
- avatar 이미지/placeholder를 감싸는 클릭 가능 `<div>` 추가
- `isDeleted === false`일 때만 `cursor-pointer` + `navigate(/${post.author.username})` + `e.stopPropagation()`
- `isDeleted === true`일 때 클릭 무시 (카드 클릭으로 fallthrough하여 post detail로 이동 — 이 동작은 정상)

라인 21-31을 다음과 같이 변경:
```tsx
// Before
<div className="flex flex-col items-center">
  {post.author.profileImageUrl ? (
    <img ... />
  ) : (
    <div ... />
  )}
  <div className="mt-1 w-0.5 flex-1 bg-border" />
</div>

// After
<div className="flex flex-col items-center">
  <div
    className={cn(
      "shrink-0",
      !post.author.isDeleted && "cursor-pointer",
    )}
    onClick={(e) => {
      e.stopPropagation();
      if (!post.author.isDeleted) navigate(`/${post.author.username}`);
    }}
  >
    {post.author.profileImageUrl ? (
      <img ... />
    ) : (
      <div ... />
    )}
  </div>
  <div className="mt-1 w-0.5 flex-1 bg-border" />
</div>
```

**추가 import**: `cn`은 이미 import되어 있으므로 추가 변경 불필요.

### 4.3 ReplyCard.tsx + PostDetailPage.tsx — OP 뱃지 `authorId` 기반 비교

#### 4.3.1 PostDetailPage.tsx (prop 전달 변경)

**현재 코드** (라인 380-387):
```tsx
{post.topReplies?.map((reply) => (
  <ReplyCard
    key={reply.id}
    reply={reply}
    parentPostId={postId}
    opUsername={post.author.username}
  />
))}
```

**문제**: `opUsername={post.author.username}`로 전달하면 탈퇴 사용자의 경우 `"deleted"` 문자열이 전달되어, 다른 탈퇴 사용자의 답글에도 OP 뱃지가 표시된다.

**변경 내용**:
- `opUsername` prop 대신 `opAuthorId` prop으로 변경
- `post.authorId`를 전달 (UUID이므로 고유성 보장)

```tsx
// Before
opUsername={post.author.username}

// After
opAuthorId={post.authorId}
```

#### 4.3.2 ReplyCard.tsx (interface + 비교 로직 변경)

**현재 코드**:

인터페이스 (라인 44-49):
```tsx
interface ReplyCardProps {
  reply: PostDetail;
  parentPostId?: string;
  opUsername?: string;
  hasNextSibling?: boolean;
}
```

비교 로직 (라인 65-66, 74):
```tsx
const isParentAuthor =
  opUsername != null && currentUser?.username === opUsername;
const canDelete = isOwner || isParentAuthor;
// ...
const isOP = opUsername != null && reply.author.username === opUsername;
```

재귀 전달 (라인 309):
```tsx
opUsername={opUsername}
```

**변경 내용**:

1. Props interface 변경:
```tsx
// Before
opUsername?: string;

// After
opAuthorId?: string;
```

2. `isParentAuthor` 로직 변경 (라인 65-66):
```tsx
// Before
const isParentAuthor =
  opUsername != null && currentUser?.username === opUsername;

// After
const isParentAuthor =
  opAuthorId != null && currentUser?.id === opAuthorId;
```

> 주의: `currentUser`는 `useAuth()`에서 반환되는 `User` 타입이며, `id` 필드가 존재함 (`User.id: string`).

3. `isOP` 로직 변경 (라인 74):
```tsx
// Before
const isOP = opUsername != null && reply.author.username === opUsername;

// After
const isOP = opAuthorId != null && reply.authorId === opAuthorId;
```

> `reply.authorId`는 `PostDetail.authorId: string`으로 이미 존재.

4. 재귀 prop 전달 변경 (라인 309):
```tsx
// Before
opUsername={opUsername}

// After
opAuthorId={opAuthorId}
```

5. 함수 시그니처 destructuring 변경 (라인 51-56):
```tsx
// Before
export default function ReplyCard({
  reply,
  parentPostId,
  opUsername,
  hasNextSibling = false,
}: ReplyCardProps) {

// After
export default function ReplyCard({
  reply,
  parentPostId,
  opAuthorId,
  hasNextSibling = false,
}: ReplyCardProps) {
```

---

## 5. API 변경사항

없음. 순수 프론트엔드 수정이며, 백엔드 응답의 `authorId` 필드와 `PostAuthor.isDeleted` 필드는 이미 존재한다.

---

## 6. 엣지 케이스

| # | 케이스 | 예상 동작 |
|---|--------|-----------|
| E-1 | `post.parent`가 `null`인 경우 (최상위 게시글) | "replying to" 블록 자체가 렌더링되지 않으므로 영향 없음 |
| E-2 | `post.parent.author.isDeleted`가 `undefined`인 경우 | `PostAuthor.isDeleted`는 optional (`isDeleted?: boolean`)이므로 `undefined`는 falsy, 기존 스타일 유지 (안전) |
| E-3 | 탈퇴 사용자가 OP인 게시글에 비탈퇴 사용자가 답글을 달 때 | `reply.authorId !== post.authorId`이므로 OP 뱃지 미표시 (정상) |
| E-4 | 탈퇴 사용자가 OP인 게시글에 같은 탈퇴 사용자(본인)의 답글 | 탈퇴 시 soft delete이므로 해당 사용자의 `authorId`는 동일 UUID. OP 뱃지 정상 표시 |
| E-5 | 2명 이상의 서로 다른 탈퇴 사용자가 같은 게시글에 답글 | 각자 다른 `authorId`를 가지므로 OP가 아닌 사용자에게는 뱃지 미표시 (버그 수정 핵심) |
| E-6 | ParentPostCard에서 avatar가 없는 탈퇴 사용자 | `profileImageUrl`이 빈 문자열이면 기존 placeholder `<div>` 렌더링, 클릭 가드 동일 적용 |
| E-7 | `currentUser`가 `null`(로그아웃 상태)일 때 `isParentAuthor` 비교 | `currentUser?.id`가 `undefined`이므로 `false` (안전) |

---

## 7. 의존성 및 제약사항

- **타입 의존성**: `PostDetail.authorId` (string) — 이미 존재
- **타입 의존성**: `User.id` (string) — `useAuth()` 반환값에 이미 존재
- **하위 호환성**: `ReplyCard`의 `opUsername` prop이 `opAuthorId`로 변경되므로, `ReplyCard`를 사용하는 모든 곳을 업데이트해야 함
  - `PostDetailPage.tsx` (라인 385) — 본 스펙에 포함
  - 다른 곳에서 `ReplyCard`를 사용하는지 확인 필요

---

## 8. 영향 범위 확인

`ReplyCard` import 사용처를 확인해야 한다:
- `PostDetailPage.tsx` — 본 스펙에 포함
- `ReplyCard.tsx` 자체 (재귀 호출) — 본 스펙에 포함

---

## 9. 다음 단계 권장사항

1. **구현 담당**: Frontend Agent가 4개 파일 수정
2. **검증**: `bun run check` 통과 확인
3. **수동 테스트**: 탈퇴 사용자가 포함된 시나리오에서 PostCard/ParentPostCard/PostDetailPage 동작 확인
4. **PR 생성**: `feat/issue-69-deleted-user-guard` 브랜치에서 master로 PR
