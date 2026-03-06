# Action Dropdown Unification Spec

## 기능 설명 (What)

PostCard, ReplyCard, PostDetailPage 세 컴포넌트에 걸쳐 액션 버튼(팔로우, 북마크, 공유)의 배치를 통일한다.

### 핵심 변경 요약

| 컴포넌트 | 변경 전 | 변경 후 |
|----------|---------|---------|
| **PostCard** (홈 피드) | 본인: 드롭다운(수정/삭제), 타인: 팔로우 버튼. 하단에 Bookmark/Share 독립 버튼 | 모든 사용자에게 MoreHorizontal 드롭다운 표시. 하단에서 Bookmark/Share 제거 |
| **ReplyCard** (디테일 페이지 답글) | 본인/OP만 드롭다운(수정/삭제), 타인은 아무 버튼 없음. 하단에 Bookmark/Share 없음 | 모든 로그인 사용자에게 드롭다운 표시 |
| **PostDetailPage** (글 상세) | 본인: 드롭다운(수정/삭제), 타인: 팔로우 버튼(유지). 하단에 Bookmark/Share 독립 버튼 | 팔로우 버튼 유지. 하단에서 Bookmark/Share 제거, 드롭다운에 편입 |

---

## 구현 이유 (Why)

1. **UI 일관성**: PostCard에서 본인 글은 드롭다운, 타인 글은 팔로우 버튼이 표시되어 시각적으로 불균형하다. 모든 카드에 동일한 MoreHorizontal 아이콘을 배치하면 통일감이 생긴다.
2. **팔로우 버튼 집중**: 팔로우 버튼이 피드의 모든 카드에 나타나면 시각적 노이즈가 크고, 실수로 누를 가능성도 있다. 프로필 페이지와 디테일 페이지(메인 포스트)로 한정하면 의도적 팔로우를 유도한다.
3. **하단 액션 바 간소화**: Bookmark/Share를 드롭다운에 편입하면 하단 액션 바가 핵심 인터랙션(Reply, Repost, Like, View Count)에 집중된다.
4. **ReplyCard 기능 확장**: 현재 타인의 답글에는 아무 액션 버튼이 없어 북마크/공유가 불가능하다. 드롭다운을 통해 이를 해결한다.

---

## 상세 변경 사항

### 1. PostCard.tsx

#### 드롭다운 메뉴 변경

**변경 전**: `isOwner`일 때만 드롭다운 표시, `!isOwner && currentUser`일 때 팔로우 버튼 표시

**변경 후**: 로그인 사용자에게 항상 MoreHorizontal 드롭다운 표시 (비로그인 시 드롭다운 미표시)

드롭다운 메뉴 항목:
- **본인 글**: 수정, 삭제, (구분선), 북마크 추가/제거, 공유하기
- **타인 글**: 북마크 추가/제거, 공유하기 (팔로우 없음 - PostCard에서 팔로우 제거)

```
[MoreHorizontal]
+-- 본인일 때 --------+     +-- 타인일 때 ----------+
|  [Pencil] 수정       |     |  [Bookmark] 북마크    |
|  [Trash2] 삭제       |     |  [Share] 공유하기     |
|  ─── 구분선 ───      |     +----------------------+
|  [Bookmark] 북마크   |
|  [Share] 공유하기    |
+---------------------+
```

#### 하단 액션 바 변경

**제거할 버튼**: Bookmark 독립 버튼, Share 독립 버튼

**변경 후 하단 액션 바**: Reply, Repost, Like, View Count (4개)

#### 제거할 상태/훅

- `isHoveringFollow` useState 제거
- `useProfile(post.author.username, !isOwner)` 호출은 유지 (드롭다운에서 팔로우 상태 확인 필요)
- `useFollow`, `useUnfollow` 훅 유지 (드롭다운에서 사용)
- `handleFollowClick`을 드롭다운 onClick으로 이동 (e.stopPropagation 불필요, 드롭다운이 처리)
- `handleBookmarkClick` 제거 (드롭다운 onClick 인라인으로 대체)
- `showShareModal` 상태 유지 (ShareModal은 그대로 사용, 드롭다운 항목 클릭 시 열림)

#### 공유하기 동작

ShareModal(Dialog)은 기존 그대로 유지한다. 드롭다운의 "공유하기" 클릭 시 `setShowShareModal(true)` 호출.

주의: DropdownMenu가 닫힌 후 Dialog가 열려야 한다. shadcn/ui의 DropdownMenuItem에서 Dialog를 트리거할 때 포커스 충돌이 발생할 수 있으므로, `onSelect` 이벤트에서 `e.preventDefault()`를 호출하고 수동으로 ShareModal state를 토글한다. 또는 `setTimeout(() => setShowShareModal(true), 0)` 패턴을 사용한다.

---

### 2. ReplyCard.tsx

#### 드롭다운 메뉴 변경

**변경 전**: `canDelete` (본인 또는 OP)일 때만 드롭다운 표시

**변경 후**: 로그인 사용자에게 항상 드롭다운 표시

드롭다운 메뉴 항목:
- **본인 답글**: 수정, 삭제, (구분선), 북마크 추가/제거, 공유하기
- **OP (부모 글 작성자, 답글 작성자 아님)**: 삭제, (구분선), 북마크 추가/제거, 공유하기
- **일반 로그인 사용자**: 북마크 추가/제거, 공유하기

```
[MoreHorizontal]
+-- 본인 답글 ---------+     +-- OP -----------+     +-- 일반 사용자 --------+
|  [Pencil] 수정       |     |  [Trash2] 삭제   |     |  [Bookmark] 북마크   |
|  [Trash2] 삭제       |     |  ─── 구분선 ───  |     |  [Share] 공유하기    |
|  ─── 구분선 ───      |     |  [Bookmark] 북마크|     +---------------------+
|  [Bookmark] 북마크   |     |  [Share] 공유하기 |
|  [Share] 공유하기    |     +-----------------+
+---------------------+
```

#### 추가할 훅/상태

- `useBookmark(reply.id, reply.isBookmarked)` 추가
- `showShareModal` useState 추가
- `ShareModal` 컴포넌트 import 및 렌더링 추가

#### 하단 액션 바

변경 없음. 기존: Like, Reply, View Count 유지.

---

### 3. PostDetailPage.tsx

#### 드롭다운 메뉴 변경

**변경 전**: `isOwner`일 때만 드롭다운(수정/삭제) 표시

**변경 후**: 로그인 사용자에게 항상 드롭다운 표시. 팔로우 버튼은 **그대로 유지** (요구사항).

주의: PostDetailPage에서는 팔로우 버튼과 드롭다운이 **공존**한다.
- 본인 글: 드롭다운(수정, 삭제, 북마크, 공유). 팔로우 버튼 미표시.
- 타인 글: 팔로우 버튼 + 드롭다운(북마크, 공유).

레이아웃:
```
[Avatar] [Name/Handle]          [FollowBtn(타인만)] [MoreHorizontal]
```

드롭다운 메뉴 항목:
- **본인 글**: 수정, 삭제, (구분선), 북마크 추가/제거, 공유하기
- **타인 글**: 북마크 추가/제거, 공유하기

#### 하단 액션 바 변경

**제거할 버튼**: Bookmark 독립 버튼, Share 독립 버튼

**변경 후 하단 액션 바**: Reply, Repost, Like (3개)

---

## 수락 기준 (Acceptance Criteria)

### PostCard

- [ ] AC-1: 로그인 사용자가 피드에서 본인 글을 볼 때, MoreHorizontal 드롭다운에 "수정", "삭제", "북마크 추가", "공유하기" 메뉴가 표시된다.
- [ ] AC-2: 로그인 사용자가 피드에서 타인 글을 볼 때, MoreHorizontal 드롭다운에 "팔로우" (또는 "언팔로우"), "북마크 추가", "공유하기" 메뉴가 표시된다.
- [ ] AC-3: PostCard에서 팔로우 버튼(Button 컴포넌트)이 더 이상 표시되지 않는다.
- [ ] AC-4: PostCard 하단 액션 바에서 Bookmark 독립 버튼과 Share 독립 버튼이 제거되었다.
- [ ] AC-5: 비로그인 사용자가 피드를 볼 때, 드롭다운 버튼이 표시되지 않는다.
- [ ] AC-6: 드롭다운의 "공유하기" 클릭 시 기존 ShareModal이 정상적으로 열린다.
- [ ] AC-7: 드롭다운의 "북마크 추가/제거" 클릭 시 bookmark 토글이 정상 동작하며, 이미 북마크된 경우 "북마크 제거" 텍스트로 표시된다.
- [ ] AC-8: 드롭다운의 팔로우/언팔로우 클릭 시 팔로우 상태가 토글된다.

### ReplyCard

- [ ] AC-9: 로그인 사용자가 타인의 답글에서도 MoreHorizontal 드롭다운이 표시된다.
- [ ] AC-10: 일반 사용자의 드롭다운에는 "북마크 추가/제거", "공유하기"만 표시된다.
- [ ] AC-11: 본인 답글의 드롭다운에는 "수정", "삭제", "북마크 추가/제거", "공유하기"가 표시된다.
- [ ] AC-12: OP의 드롭다운에는 "삭제", "북마크 추가/제거", "공유하기"가 표시된다.
- [ ] AC-13: ReplyCard에서 공유하기 클릭 시 ShareModal이 정상적으로 열린다.

### PostDetailPage

- [ ] AC-14: 타인 글 상세에서 팔로우 버튼이 **그대로 유지**된다.
- [ ] AC-15: 로그인 사용자에게 항상 MoreHorizontal 드롭다운이 표시된다.
- [ ] AC-16: 본인 글 상세의 드롭다운에 "수정", "삭제", "북마크 추가/제거", "공유하기"가 표시된다.
- [ ] AC-17: 타인 글 상세의 드롭다운에 "북마크 추가/제거", "공유하기"가 표시된다 (팔로우는 독립 버튼으로 존재).
- [ ] AC-18: 하단 액션 바에서 Bookmark, Share 독립 버튼이 제거되었다.

### 공통

- [ ] AC-19: DropdownMenuSeparator가 관리 액션(수정/삭제/팔로우)과 유틸리티 액션(북마크/공유) 사이에 표시된다.
- [ ] AC-20: 북마크 상태에 따라 아이콘과 텍스트가 동적으로 변경된다 (북마크 추가 / 북마크 제거).
- [ ] AC-21: `bun run check` (타입체크 + 린트)가 에러 없이 통과한다.

---

## API 엔드포인트 설계

백엔드 변경 없음. 기존 API를 그대로 사용한다.

| 기능 | 메서드 | 경로 | 비고 |
|------|--------|------|------|
| 북마크 추가 | POST | `/api/posts/:id/bookmark` | 기존 |
| 북마크 제거 | DELETE | `/api/posts/:id/bookmark` | 기존 |
| 팔로우 | POST | `/api/users/:handle/follow` | 기존 |
| 언팔로우 | DELETE | `/api/users/:handle/follow` | 기존 |
| 프로필 조회 | GET | `/api/users/:handle` | 팔로우 상태 확인용, 기존 |

---

## ReBAC 고려사항

이 변경은 프론트엔드 UI 재배치만 해당하며, 접근 제어 로직의 변경은 없다. 기존 백엔드 권한 검증이 그대로 적용된다.

- 북마크: 로그인 사용자만 가능 (백엔드 auth middleware)
- 팔로우/언팔로우: 로그인 사용자만 가능, 자기 자신 팔로우 불가 (백엔드 검증)
- 수정/삭제: 본인만 가능 (백엔드 검증), Reply 삭제는 OP도 가능 (백엔드 검증)

---

## 엣지 케이스 목록

1. **비로그인 사용자**: 드롭다운 버튼 자체가 렌더링되지 않아야 한다. 현재 PostCard에서 `currentUser`가 없으면 팔로우 버튼이 숨겨지는데, 동일하게 드롭다운도 `currentUser`가 있을 때만 표시한다.

2. **팔로우 상태 로딩 중**: `authorProfile`이 아직 로드되지 않은 상태에서 드롭다운의 팔로우 메뉴 텍스트가 결정되지 않는다. `authorProfile`이 없으면 "팔로우" 텍스트를 기본값으로 표시하거나, 로딩 중 상태를 표시한다.

3. **DropdownMenu + Dialog 포커스 충돌**: DropdownMenu가 닫히면서 포커스가 트리거로 돌아가고, Dialog가 열리면서 포커스를 가져가려 하면 충돌할 수 있다. `setTimeout`으로 Dialog 열기를 지연시키거나, DropdownMenuItem의 `onSelect`에서 `e.preventDefault()`를 사용한다.

4. **Optimistic UI 불일치**: 드롭다운에서 북마크를 토글한 후, 하단 액션 바에서 북마크 아이콘이 사라졌으므로 시각적 피드백이 없다. 드롭다운 메뉴 텍스트("북마크 추가" / "북마크 제거")로 상태 피드백을 제공한다. toast 알림도 고려할 수 있다.

5. **ReplyCard에서 isBookmarked 데이터 존재 여부**: 현재 백엔드 API가 reply에 대해서도 `isBookmarked` 필드를 반환하는지 확인 필요. PostDetail 타입에 `isBookmarked`가 있으므로 reply도 동일 타입을 사용하여 문제없을 것으로 예상된다.

6. **드롭다운 항목이 북마크/공유만 있는 경우**: 일반 사용자가 타인의 ReplyCard를 볼 때, 구분선(Separator) 없이 2개 항목만 표시된다. 구분선은 관리 액션이 있을 때만 렌더링한다.

7. **PostDetailPage에서 타인 글일 때 팔로우 버튼과 드롭다운 동시 표시**: 레이아웃이 가로로 충분한 공간이 있어야 한다. 기존 flex 레이아웃에서 팔로우 버튼 오른쪽에 드롭다운을 추가 배치한다.

---

## 의존성 및 제약사항

### 의존성

- `shadcn/ui`: DropdownMenu, DropdownMenuSeparator (이미 설치됨)
- `lucide-react`: UserPlus, UserMinus 아이콘 추가 import 필요 (팔로우/언팔로우 드롭다운 아이콘)
- `useBookmark` 훅: ReplyCard에 신규 도입
- `ShareModal` 컴포넌트: ReplyCard에 신규 도입
- `useFollow`, `useUnfollow` 훅: PostCard에서 계속 사용 (드롭다운으로 이동)

### 제약사항

- 백엔드 변경 없음 (프론트엔드만 변경)
- ParentPostCard는 변경 대상이 아님 (부모 글 체인 표시용, 최소 UI)
- ProfilePage의 팔로우 버튼은 변경 대상이 아님

### 변경 파일 목록

| 파일 | 변경 유형 |
|------|----------|
| `frontend/src/components/PostCard.tsx` | 수정 (드롭다운 통합, 하단 버튼 제거) |
| `frontend/src/components/ReplyCard.tsx` | 수정 (드롭다운 확장, 북마크/공유 추가) |
| `frontend/src/pages/PostDetailPage.tsx` | 수정 (드롭다운 통합, 하단 버튼 제거) |

### 추가 import 필요

| 컴포넌트 | 신규 import |
|----------|-------------|
| PostCard | `UserPlus`, `UserMinus` (lucide-react), `DropdownMenuSeparator` |
| ReplyCard | `Bookmark`, `Share`, `UserPlus`, `UserMinus` (lucide-react), `DropdownMenuSeparator`, `useBookmark`, `ShareModal` |
| PostDetailPage | `UserPlus`, `UserMinus` (lucide-react), `DropdownMenuSeparator` |

---

## 구현 순서 권장

1. **PostCard.tsx** 수정 (가장 큰 변경, 패턴 확립)
2. **ReplyCard.tsx** 수정 (PostCard 패턴 재활용, 신규 훅 추가)
3. **PostDetailPage.tsx** 수정 (팔로우 버튼 유지 주의)
4. `bun run check` 타입체크/린트 통과 확인
5. 수동 UI 테스트 (로그인/비로그인, 본인/타인, 팔로우 상태 토글, 북마크 토글, 공유 모달)

---

## 다음 단계 권장사항

이 스펙이 승인되면, **Frontend Agent**가 구현을 이어받아야 한다. 구현 시 다음 사항에 주의:
- DropdownMenu와 Dialog(ShareModal)의 포커스 충돌 패턴 테스트
- PostCard에서 `isHoveringFollow` 상태와 Button variant 관련 코드 정리
- ReplyCard의 `isBookmarked` 필드가 API 응답에 포함되는지 런타임 확인
