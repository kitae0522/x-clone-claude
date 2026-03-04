#!/bin/bash
# .claude/hooks/self-check.sh

cat << 'EOF'
{"feedback": "[셀프 체크] 작업 완료 전 확인:\n1. Go: ctx가 첫 번째 파라미터인가?\n2. Go: 인터페이스 포인터(*I)를 쓰지 않았나?\n3. Go: panic 대신 fmt.Errorf를 썼나?\n4. React: API 호출이 커스텀 훅으로 분리되었나?\n5. ReBAC: 접근 제어 검증이 service에 있나?\n6. docs/TODO.md를 업데이트했나?\n7. /clear가 필요한 시점인가?"}
EOF
