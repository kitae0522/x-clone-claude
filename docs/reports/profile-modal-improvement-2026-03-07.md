# Profile Modal Improvement Report

**Issue**: #51
**Branch**: feat/issue-51-profile-modal-upload
**Date**: 2026-03-07

## 변경 요약

### 1. 버튼 겹침 수정
- DialogContent의 기본 X 닫기 버튼을 CSS `[&>button:last-child]:hidden`으로 숨김
- DialogHeader에 직접 X 닫기 버튼 배치 (좌측)
- 레이아웃: `[X 닫기] [프로필 수정] ... [저장]`

### 2. 이미지 파일 업로드
- URL 텍스트 입력 필드 2개 제거
- 헤더 배너 영역 클릭 -> 파일 선택 -> 업로드
- 프로필 아바타 클릭 -> 파일 선택 -> 업로드
- hover 시 Camera 아이콘 오버레이, 업로드 중 Loader2 스피너

### 3. 코드 품질 개선 (리뷰 반영)
- `uploadImage` raw fetch -> `useUploadProfileImage` 커스텀 훅 (useMutation + apiFetch)
- 로딩 오버레이 중복 렌더링 제거
- username 필드에 `maxLength={30}` 추가
- form onSubmit 이중 호출 경로 제거

## 변경 파일
| 파일 | 변경 내용 |
|------|-----------|
| `frontend/src/components/EditProfileModal.tsx` | 전면 리팩토링 |
| `frontend/src/hooks/useProfile.ts` | `useUploadProfileImage` 훅 추가 |
| `docs/specs/profile-modal-improvement-spec.md` | 스펙 문서 |
| `docs/TODO.md` | Phase 14 추가 |

## 백엔드 변경
없음. 기존 `/api/media/upload` + `PUT /api/users/profile` API 재활용.

## 테스트
- TypeScript 타입체크 + ESLint: PASS
- 프론트엔드 단위 테스트 인프라 미구축 (별도 이슈 필요)
