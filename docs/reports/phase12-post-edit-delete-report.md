# Phase 12: Post/Reply 수정 및 삭제 - 최종 보고서

## 개요
Post와 Reply의 수정(Edit) 및 삭제(Delete) 기능을 구현했습니다. Soft delete 패턴을 적용하여 데이터 보존성을 확보했고, 소유권 기반 권한 검증을 통해 보안을 보장합니다.

## 구현 범위

### 백엔드
| 항목 | 파일 | 설명 |
|------|------|------|
| DB 마이그레이션 | `migrations/014_add_deleted_at_to_posts` | `deleted_at TIMESTAMPTZ` 컬럼 + 부분 인덱스 |
| AppError | `apperror/apperror.go` | `Forbidden` (403) 추가 |
| Model | `model/post.go` | `DeletedAt *time.Time` 필드 |
| DTO | `dto/post_dto.go` | `UpdatePostRequest`, `DeletePostResponse` |
| Repository | `repository/post_repository.go` | `Update`, `SoftDelete`, `SoftDeleteReply` + 14개 쿼리에 `deleted_at IS NULL` 필터 |
| Repository | `repository/poll_repository.go` | `DeleteByPostID` |
| Repository | `repository/media_repository.go` | `UnlinkByPostID` |
| Service | `service/post_service.go` | `UpdatePost`, `DeletePost` (소유권 검증, ReBAC) |
| Handler | `handler/post_handler.go` | `UpdatePost` (PUT), `DeletePost` (DELETE) |
| Router | `router/router.go` | `PUT /:id`, `DELETE /:id` 라우트 |

### 프론트엔드
| 항목 | 파일 | 설명 |
|------|------|------|
| Types | `types/api.ts` | `UpdatePostRequest` 인터페이스 |
| Hooks | `hooks/usePosts.ts` | `useUpdatePost`, `useDeletePost` |
| UI 컴포넌트 | `components/ui/dropdown-menu.tsx` | Radix UI 기반 드롭다운 |
| UI 컴포넌트 | `components/ui/alert-dialog.tsx` | Radix UI 기반 확인 다이얼로그 |
| PostCard | `components/PostCard.tsx` | 드롭다운 메뉴, 인라인 수정, 삭제 확인, (edited) 표시 |
| ReplyCard | `components/ReplyCard.tsx` | 동일 패턴 적용 |
| PostDetailPage | `pages/PostDetailPage.tsx` | 동일 패턴 + 삭제 후 홈 이동 |

## 핵심 설계 결정

### 1. Soft Delete
- `deleted_at` 컬럼으로 논리적 삭제, 실제 데이터는 DB에 보존
- 모든 SELECT 쿼리(14개+)에 `WHERE deleted_at IS NULL` 조건 추가
- 부분 인덱스 `WHERE deleted_at IS NULL`로 성능 최적화

### 2. 권한 모델
- **수정**: 본인 소유 post/reply만 수정 가능
- **삭제 (Post)**: 본인만 삭제 가능
- **삭제 (Reply)**: 본인 + 부모 post 작성자 삭제 가능 (ReBAC)

### 3. Reply 삭제 시 트랜잭션
- `SoftDeleteReply`는 트랜잭션 내에서 soft delete + 부모의 `reply_count` 감소를 원자적으로 처리
- `GREATEST(reply_count - 1, 0)`으로 음수 방지

### 4. 부분 업데이트 (Partial Update)
- `UpdatePostRequest`에서 포인터 타입(`*string`)으로 nil vs 부재 구분
- `ClearLocation`, `ClearPoll` boolean 플래그로 명시적 제거 지원
- 미디어: unlink 후 재연결 방식

### 5. 프론트엔드 UX
- 인라인 편집 (Textarea 토글)
- DropdownMenu (MoreHorizontal 아이콘)로 Edit/Delete 접근
- AlertDialog로 삭제 확인
- `(edited)` 표시: `updatedAt !== createdAt`
- React Query 캐시 무효화로 즉시 반영

## 테스트 결과
- Go 백엔드: 기존 테스트 전체 통과 (mock 업데이트 포함)
- Frontend: TypeScript 빌드 성공

## 코드 리뷰 결과 (부분)
- `init()` 함수: 없음 (통과)
- `panic()` 호출: 없음 (통과)
- Sentinel errors: 패키지 상단 정의 패턴 준수
- `time.Now()` 사용: SQL `NOW()` 사용으로 서버 시간 일관성 확보
- 에러 핸들링: `fmt.Errorf` 래핑 + `apperror` 패턴 준수

## 작성일
2026-03-06
