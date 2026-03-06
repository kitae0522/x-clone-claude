# View Count 기능 구현 보고서

- **작성일**: 2026-03-06
- **브랜치**: feat/issue-37-compose-customization
- **스펙**: docs/specs/view-count-spec.md

---

## 1. 구현 요약

모든 post/reply에 조회수(view_count) 수치를 추가. 게시글 상세 조회 시 조회수 증가, 모든 UI에서 Eye 아이콘으로 표시.

## 2. 변경 파일

### 백엔드 (6 files)
| 파일 | 변경 내용 |
|------|-----------|
| `migrations/013_add_view_count_to_posts.up.sql` | ALTER TABLE - view_count 컬럼 추가 |
| `migrations/013_add_view_count_to_posts.down.sql` | 롤백 마이그레이션 |
| `internal/model/post.go` | Post.ViewCount 필드 추가 |
| `internal/dto/post_dto.go` | PostDetailResponse.ViewCount + 매핑 |
| `internal/repository/post_repository.go` | 14개 SELECT에 view_count 추가, IncrementViewCount 메서드 |
| `internal/service/post_service.go` | GetPostByID에서 비본인 조회 시 IncrementViewCount 호출 |

### 프론트엔드 (5 files)
| 파일 | 변경 내용 |
|------|-----------|
| `src/types/api.ts` | PostDetail.viewCount 추가 |
| `src/lib/formatTime.ts` | formatCompactNumber 유틸 (1.2K, 3.4M) |
| `src/components/PostCard.tsx` | Eye 아이콘 + 조회수 표시 |
| `src/pages/PostDetailPage.tsx` | Stats에 조회수 통계 추가 |
| `src/components/ReplyCard.tsx` | Eye 아이콘 + 조회수 표시 |

### 테스트 (3 files)
| 파일 | 변경 내용 |
|------|-----------|
| `internal/service/post_service_test.go` | IncrementViewCount mock + 2 테스트 (4 서브케이스) |
| `internal/service/bookmark_service_test.go` | IncrementViewCount mock 추가 |
| `internal/service/poll_service_test.go` | IncrementViewCount mock 추가 |

## 3. 테스트 결과

| 테스트 | 결과 |
|--------|------|
| TestGetPostByID_IncrementsViewCount (0->1) | PASS |
| TestGetPostByID_IncrementsViewCount (5->6) | PASS |
| TestGetPosts_DoesNotIncrementViewCount (stays 0) | PASS |
| TestGetPosts_DoesNotIncrementViewCount (stays 10) | PASS |
| 전체 Go 테스트 | PASS |
| 프론트엔드 빌드 (tsc + vite) | PASS |

## 4. 코드 리뷰 결과

### 수정된 이슈
| 심각도 | 이슈 | 수정 내용 |
|--------|------|-----------|
| Critical | 본인 조회 시 조회수 증가 | 비로그인 + 비본인만 증가하도록 가드 추가 |
| Warning | stale viewCount 반환 | increment 성공 시 result.ViewCount++ 반영 |
| Warning | ReplyCard 0 표시 비일관성 | PostCard와 동일하게 빈 문자열로 통일 |

### 수용된 이슈 (현재 스코프 밖)
| 심각도 | 이슈 | 사유 |
|--------|------|------|
| Critical | Redis 비동기 배치 처리 | MVP 수준에서 과도한 설계. PostgreSQL 원자적 UPDATE로 충분 |
| Warning | errors.Is 미사용 | 기존 코드 이슈, 이번 변경 스코프 밖 |
| Warning | time.Now DI | 기존 코드 이슈 |

## 5. 설계 결정 기록

1. **조회수 증가 조건**: 비로그인 사용자 + 본인이 아닌 로그인 사용자만 (봇/자기 조회 방지)
2. **카운트 저장**: posts 테이블 직접 저장 (like_count 패턴 일관성)
3. **큰 숫자 포맷팅**: formatCompactNumber 유틸 추가 (1.2K, 3.4M)
4. **에러 처리**: IncrementViewCount 실패 시 slog 로그만, 정상 응답 유지
