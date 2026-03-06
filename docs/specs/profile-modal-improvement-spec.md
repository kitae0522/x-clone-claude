# Profile Modal Improvement Spec

## 요약
프로필 수정 모달의 이미지 삽입 방식을 URL 입력에서 직접 파일 업로드로 변경하고, 우측 상단 버튼 겹침 문제를 해결한다.

## 문제 분석

### 1. 이미지 삽입 방식
- **현재**: URL 텍스트 입력 (`<Input type="url">`)
- **문제**: 사용자가 외부 이미지 URL을 직접 알아야 함, UX 불편
- **목표**: 파일 선택 → 업로드 → URL 자동 설정

### 2. 버튼 겹침
- **현재**: `DialogContent`가 `absolute right-4 top-4`에 X(닫기) 버튼 자동 렌더, `DialogHeader`에 "저장" 버튼이 우측 배치
- **문제**: 닫기(X) 버튼과 저장 버튼이 같은 위치에 겹침
- **목표**: 닫기 버튼을 DialogHeader 좌측으로 이동, 저장 버튼은 우측 유지

## 구현 계획

### Task 1: 버튼 겹침 해결
**파일**: `frontend/src/components/EditProfileModal.tsx`
- DialogHeader에 X 닫기 버튼을 직접 배치 (좌측)
- DialogContent의 기본 X 버튼 숨기기 (빈 `<DialogClose>` 사용 또는 CSS로 숨김)
- 레이아웃: `[X 닫기] [프로필 수정 타이틀] ... [저장 버튼]`

### Task 2: 이미지 파일 업로드 UI
**파일**: `frontend/src/components/EditProfileModal.tsx`
- 기존 URL Input 2개 제거
- 프로필 이미지: 아바타 클릭 → 파일 선택 → 업로드
- 헤더 이미지: 배너 영역 클릭 → 파일 선택 → 업로드
- 업로드 중 로딩 표시
- 기존 `useMediaUpload` 훅의 `uploadFile` 로직 재활용 (단, 훅 전체가 아닌 업로드 API만 사용)

### Task 3: 업로드 후 URL 설정
- 미디어 업로드 성공 → 반환된 URL을 profileImageUrl/headerImageUrl state에 설정
- 저장 시 기존 `useUpdateProfile` 그대로 사용 (URL 문자열로 서버에 전달)

## 백엔드 변경 사항
- **없음**: 기존 `/api/media/upload` 엔드포인트와 `PUT /api/users/profile` (URL 문자열) 그대로 활용

## 영향 범위
- `frontend/src/components/EditProfileModal.tsx` (주요 변경)
- `frontend/src/components/ui/dialog.tsx` (변경 없음 - 모달별로 숨김 처리)
