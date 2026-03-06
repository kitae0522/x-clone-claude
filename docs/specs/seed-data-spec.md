# Seed Data Spec

## 목표
레포지토리를 클론하고 `docker compose up`하면 초기 데모 데이터가 포함된 상태로 시작되도록 한다.

## 접근 방식
- `015_seed_data.up.sql` 마이그레이션으로 관리
- `schema_migrations` 테이블에서 추적하므로 **1회만 실행**
- `docker compose down -v && docker compose up`으로 완전 초기화 가능

## Seed 데이터 구성

### 사용자 (5명)
| handle | display_name | 비밀번호 | 설명 |
|--------|-------------|---------|------|
| alice | Alice Kim | password123 | 메인 데모 사용자 |
| bob | Bob Park | password123 | 일반 사용자 |
| charlie | Charlie Lee | password123 | 일반 사용자 |
| diana | Diana Choi | password123 | 일반 사용자 |
| eve | Eve Jung | password123 | 일반 사용자 |

### 팔로우 관계
- alice → bob, charlie, diana
- bob → alice, charlie
- charlie → alice
- diana → alice, bob, charlie, eve
- eve → alice

### 게시물 (10개+)
- 공개 게시물 7개 (다양한 마크다운, 텍스트)
- 팔로워 전용 게시물 2개
- 답글 5개 (중첩 포함)

### 좋아요
- 주요 게시물에 분산 좋아요

### 북마크
- alice가 2-3개 게시물 북마크

## 롤백
`015_seed_data.down.sql`로 seed 데이터만 삭제 가능 (UUID 기반 DELETE)
