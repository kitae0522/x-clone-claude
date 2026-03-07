# Profile Image Upload Bug Fix Spec

**Issue**: #55
**Status**: Draft
**Date**: 2026-03-07

---

## 1. 기능 설명 (What)

프로필 수정 모달에서 프로필/헤더 이미지를 파일 업로드하면, 업로드 자체는 성공하지만 이후 프로필 저장(PUT /api/users/profile) 시 이미지 URL이 실제 프로필에 반영되지 않는 버그를 수정한다.

## 2. 구현 이유 (Why)

현재 플로우에서 이미지 업로드 후 프로필 저장까지 3가지 독립적인 버그가 존재하여, 사용자가 프로필 이미지를 변경할 수 없다. 이는 기본적인 사용자 경험을 심각하게 저해한다.

## 3. 버그 분석 결과

### Bug 1 (Critical): Validator `url` 태그가 상대 경로를 거부

**위치**: `backend/internal/dto/user_dto.go:9-10`

```go
ProfileImageURL string `json:"profileImageUrl" validate:"omitempty,url"`
HeaderImageURL  string `json:"headerImageUrl"  validate:"omitempty,url"`
```

**문제**: `go-playground/validator`의 `url` 태그는 **절대 URL**(예: `http://...` 또는 `https://...`)만 유효로 판단한다. 그런데 `LocalStorage.Upload()`이 반환하는 URL은 `/uploads/2026/03/uuid.jpg` 형식의 **상대 경로**이다. 따라서 프로필 저장 요청 시 validation 단계에서 400 Bad Request로 거부된다.

**증거**:
- `local_storage.go:43-48` -- `filepath.ToSlash(filePath)` 후 `/` prefix 추가 -> `/uploads/2026/03/uuid.jpg`
- `user_dto.go:9` -- `validate:"omitempty,url"` -> 이 값은 `url` 검증 실패

### Bug 2 (Critical): 프론트엔드 `/uploads` 경로에 대한 프록시 미설정

**위치**: `frontend/vite.config.ts:15-23`

```typescript
proxy: {
  "/api": { target: "http://backend:8080", changeOrigin: true },
  "/media": { target: "http://media-service:8081", changeOrigin: true },
  // "/uploads" 프록시 없음!
}
```

**문제**: 백엔드가 `/uploads/...` 상대 경로를 반환하더라도, 프론트엔드 개발 서버에서 해당 경로로 접근하면 Vite가 404를 반환한다. `/uploads` 경로를 백엔드(port 8080)로 프록시하는 설정이 누락되어 있다.

**증거**:
- `router.go:78-80` -- `p.App.Static("/uploads", "./uploads", ...)` 로 백엔드에서 정적 파일 서빙은 설정됨
- `vite.config.ts` -- `/uploads` 프록시 없음

### Bug 3 (Medium): `UpdateProfile` 서비스에서 이미지 URL 빈 문자열 처리 오류

**위치**: `backend/internal/service/user_service.go:82-83`

```go
user.ProfileImageURL = req.ProfileImageURL
user.HeaderImageURL = req.HeaderImageURL
```

**문제**: `DisplayName`은 빈 문자열이면 기존 값 유지(`if req.DisplayName != "" { ... }`), `Bio`는 무조건 덮어쓰기(의도적). 그런데 `ProfileImageURL`과 `HeaderImageURL`도 무조건 덮어쓰기한다. 사용자가 이미지를 변경하지 않고 이름만 바꿔도, 프론트엔드에서 보내는 현재 이미지 URL이 그대로 전송되므로 큰 문제는 아니지만, 프론트엔드 state 초기화 방식에 따라 의도치 않게 이미지가 지워질 수 있다.

### Bug 4 (Low): 캐시 무효화 불완전

**위치**: `frontend/src/hooks/useProfile.ts:59-62`

```typescript
onSuccess: (user) => {
  queryClient.setQueryData(['auth', 'me'], user)
  queryClient.invalidateQueries({ queryKey: ['users', user.username] })
}
```

**분석**: 현재 캐시 무효화 자체는 정상적으로 구현되어 있다.
- `['auth', 'me']` 캐시를 업데이트된 user 데이터로 직접 교체
- `['users', handle]` 쿼리를 invalidate하여 프로필 페이지 재조회 유도

**잠재적 문제**: `setQueryData(['auth', 'me'], user)`에서 `user`는 `UserResponse` 타입이지만, `useAuth`의 `['auth', 'me']` 쿼리가 기대하는 타입과 일치하는지 확인 필요. `UpdateProfile`이 반환하는 `UserResponse`와 `/api/auth/me`가 반환하는 타입이 동일하면 문제 없음.

## 4. 수락 기준 (Acceptance Criteria)

1. **AC-1**: 프로필 수정 모달에서 프로필 이미지를 업로드하고 저장하면, 프로필 페이지에 새 이미지가 즉시 반영된다.
2. **AC-2**: 프로필 수정 모달에서 헤더 이미지를 업로드하고 저장하면, 프로필 페이지에 새 헤더 이미지가 즉시 반영된다.
3. **AC-3**: 이미지를 변경하지 않고 이름만 수정해도 기존 이미지가 유지된다.
4. **AC-4**: 업로드된 이미지 파일이 브라우저에서 정상적으로 로드된다 (개발 서버, 프로덕션 모두).
5. **AC-5**: 프로필 저장 후 사이드바/네비게이션의 아바타도 즉시 갱신된다 (캐시 무효화).

## 5. 수정 방안

### 5-1. 백엔드: Validator 태그 변경

**파일**: `backend/internal/dto/user_dto.go`

`url` 태그를 제거하고, 상대 경로(`/uploads/...`)와 절대 URL(`https://...`) 모두 허용하도록 변경한다.

**방안 A (권장)**: `url` 태그를 제거하고 커스텀 validator 등록
- `url_or_path` 커스텀 태그: `strings.HasPrefix(v, "/")` || `url` 통과
- 장점: 명시적, 안전

**방안 B (단순)**: `url` 태그 제거, `omitempty`만 유지
- `validate:"omitempty"` -- 빈 문자열이 아니면 무조건 통과
- 장점: 간단
- 단점: 임의 문자열도 통과 (보안 리스크 낮음 -- DB에 저장만 하고 서빙은 Static으로 처리)

**추천**: 방안 B. 현재 MVP 단계에서 과도한 검증보다 동작이 우선. 향후 S3 전환 시 URL 형식이 다시 바뀔 수 있으므로 유연하게 유지.

### 5-2. 프론트엔드: Vite 프록시 추가

**파일**: `frontend/vite.config.ts`

```typescript
proxy: {
  "/api": { target: "http://backend:8080", changeOrigin: true },
  "/uploads": { target: "http://backend:8080", changeOrigin: true },  // 추가
  "/media": { target: "http://media-service:8081", changeOrigin: true },
}
```

### 5-3. 백엔드: UpdateProfile 이미지 URL 보존 로직

**파일**: `backend/internal/service/user_service.go`

이미지 URL이 빈 문자열이면 기존 값을 유지하도록 수정:

```go
if req.ProfileImageURL != "" {
    user.ProfileImageURL = req.ProfileImageURL
}
if req.HeaderImageURL != "" {
    user.HeaderImageURL = req.HeaderImageURL
}
```

**주의**: 이미지를 명시적으로 제거하는 기능이 필요하면 별도 플래그(`ClearProfileImage bool`)를 추가해야 하나, 현재 UI에는 이미지 제거 기능이 없으므로 불필요.

## 6. API 엔드포인트

기존 엔드포인트 변경 없음. 동작만 수정.

| Method | Path | 변경 내용 |
|--------|------|-----------|
| PUT | `/api/users/profile` | validator 태그 수정으로 `/uploads/...` 경로 허용 |
| POST | `/api/media/upload` | 변경 없음 |
| GET | `/uploads/*` | 변경 없음 (정적 파일 서빙 이미 설정됨) |

## 7. ReBAC 고려사항

해당 없음. 프로필 이미지 업로드/수정은 본인만 가능하며, 기존 `AuthRequired` 미들웨어 + `userID` 검증으로 충분.

## 8. 엣지 케이스 목록

| # | 엣지 케이스 | 예상 동작 |
|---|-------------|-----------|
| 1 | 이미지 업로드 후 저장 버튼을 누르지 않고 모달 닫기 | 이미지 파일은 서버에 남지만 프로필에 반영 안 됨 (orphan 파일). MVP에서 허용. |
| 2 | 동시에 프로필/헤더 이미지 업로드 | `uploadImage` mutation이 하나이므로, 두 업로드가 동시에 진행되면 `isUploading` 상태가 공유됨. 하나의 업로드 완료 시 다른 것의 로딩 표시가 사라질 수 있음. (기존 이슈, 이번 스펙 범위 밖) |
| 3 | 매우 큰 파일 업로드 시도 | 프론트엔드에서 5MB 제한 체크 + 백엔드에서도 5MB 검증. 양쪽 모두 에러 메시지 표시. |
| 4 | 지원하지 않는 파일 형식 업로드 | 프론트엔드 `accept` 속성 + `validateImageFile()`로 차단. 백엔드에서도 MIME 검증. |
| 5 | 이미지 URL에 path traversal 공격 (`../../etc/passwd`) | `LocalStorage.Delete()`에 path traversal 방어 구현됨. 이미지 URL은 서버가 생성하므로 직접 주입 불가. |
| 6 | 백엔드 실행 위치가 다른 경우 (`./uploads` 상대 경로 문제) | Docker 환경에서 workdir이 고정되므로 현재는 문제 없음. 프로덕션에서는 S3 전환으로 해결. |

## 9. 의존성 및 제약사항

- **의존성**: 없음. 기존 코드 수정만 필요.
- **제약사항**:
  - LocalStorage 방식은 서버 재시작/컨테이너 재생성 시 업로드 파일 손실 위험. Docker volume 마운트 필요.
  - 프로덕션에서는 S3/CloudFront 전환이 권장되나, 이번 버그 수정 범위에서는 다루지 않음.

## 10. 수정 파일 목록 (총 3개)

| 파일 | 변경 유형 | 설명 |
|------|-----------|------|
| `backend/internal/dto/user_dto.go` | 수정 | `validate:"omitempty,url"` -> `validate:"omitempty"` (2곳) |
| `backend/internal/service/user_service.go` | 수정 | 이미지 URL 빈 문자열 시 기존 값 보존 |
| `frontend/vite.config.ts` | 수정 | `/uploads` 프록시 추가 |

## 11. 테스트 계획

| 테스트 | 유형 | 설명 |
|--------|------|------|
| UpdateProfile with relative image URL | Go unit test | `/uploads/2026/03/uuid.jpg` 형식 URL로 프로필 업데이트 성공 확인 |
| UpdateProfile with empty image URL preserves existing | Go unit test | 이미지 URL 빈 문자열 전송 시 기존 이미지 유지 확인 |
| UpdateProfile with absolute URL | Go unit test | `https://example.com/img.jpg` 형식도 여전히 동작 확인 |
| E2E: 프로필 이미지 업로드 -> 저장 -> 반영 | 수동 테스트 | 모달에서 이미지 업로드, 저장 후 프로필 페이지 확인 |
