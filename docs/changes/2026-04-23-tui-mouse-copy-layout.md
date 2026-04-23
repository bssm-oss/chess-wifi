# 2026-04-23 TUI mouse, copy, and compact layout

## 배경

기본 터미널 크기에서 체스 UI 일부가 잘리고, Host 주소 복사와 일부 화면 조작이 키보드 중심이라 실제 사용 흐름이 불편했습니다.

## 문제 또는 목표

- `80x24` 터미널에서도 체스판과 상태 패널이 잘리지 않아야 함
- 보드 이동뿐 아니라 메뉴, 방 목록, Host 시작, 주소 복사, 기권/종료를 마우스로 조작할 수 있어야 함
- Host 대기 화면에서 주소 복사가 가능해야 함
- 자동 discovery 목록과 수동 주소 입력 fallback은 유지해야 함

## 변경 내용

- 작은 터미널에서 자동으로 compact match layout 사용
- compact layout에서 보드 셀을 1줄 높이로 줄이고 사이드바를 요약형으로 표시
- 메뉴 항목과 발견된 방 항목을 마우스로 클릭해 선택/참가 가능하게 변경
- Host/Join 입력 화면에서 필드 포커스와 실행/뒤로 버튼을 마우스로 조작 가능하게 변경
- Host 대기 화면의 각 주소 옆에 `[ Copy ]` 액션 추가
- Host 대기 화면에서 `c` 키로 첫 번째 주소 복사 추가
- Match 화면에 `[ Resign ]`, `[ Quit ]` 마우스 액션 추가
- 마우스 조작과 compact layout 회귀 테스트 추가

## 설계 이유

- TUI의 기존 warm terminal 스타일은 유지하고, 작은 화면에서만 밀도를 높여 잘림을 줄였습니다.
- 주소 복사는 별도 파일 저장이나 외부 서비스 없이 시스템 클립보드에 직접 기록합니다.
- 기존 키보드 조작을 제거하지 않고 마우스 조작을 추가해 접근 경로를 늘렸습니다.

## 영향 범위

- `internal/tui/model.go`
- `internal/tui/view.go`
- `internal/tui/board.go`
- `internal/tui/styles.go`
- `internal/tui/model_test.go`
- `README.md`

## 검증 방법

- `go test ./internal/tui -count=1 -v`
- `go test ./...`
- `go build ./...`
- PTY 두 개에서 실제 TUI 실행
- 마우스로 Host 선택, Host 시작, 주소 복사, discovery Join, 보드 `e2e4`, Quit 버튼 확인

## 남아 있는 한계

- 시스템 클립보드는 OS/터미널 환경에 따라 실패할 수 있으며, 실패 시 UI 메시지로 표시합니다.
- 실제 물리 장치 간 Wi-Fi discovery는 네트워크 정책에 영향을 받을 수 있습니다.

## 후속 과제

- 터미널 폭이 극단적으로 작은 경우를 위한 더 강한 최소 화면 안내 추가
- 실제 마우스 hover 상태 표시 여부 검토
