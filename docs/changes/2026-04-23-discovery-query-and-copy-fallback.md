# 2026-04-23 discovery query and copy fallback

## 배경

같은 Wi-Fi에서 열린 방이 안정적으로 보이지 않고, Host 주소 복사가 터미널/OS 환경에 따라 동작하지 않는 문제가 보고되었습니다.

## 문제 또는 목표

- Join 화면이 켜졌을 때 열린 Host를 더 적극적으로 찾을 수 있어야 함
- Host discovery broadcast가 실제 네트워크로 나갈 수 있도록 socket 옵션을 명시해야 함
- Host 주소 복사는 시스템 클립보드 실패 시에도 터미널 OSC52 경로를 같이 시도해야 함
- 수동 `IP:PORT` 입력 fallback은 유지해야 함

## 변경 내용

- discovery message에 `kind` 필드 추가
- Host가 UDP `18787` 에서 query를 수신하고 announcement를 즉시 응답
- Join scan 시작 시 discovery query broadcast 송신
- Unix discovery socket에 `SO_BROADCAST` 설정 추가
- query payload가 match로 표시되지 않도록 필터링
- Host 주소 복사 시 시스템 클립보드와 OSC52 sequence를 같이 사용
- README와 LAN architecture 문서 갱신

## 설계 이유

- 기존 방식은 Host의 주기적 announcement를 Join이 우연히 수신해야 했습니다. query/response를 추가하면 Join 시작 시점에 Host가 즉시 응답할 수 있습니다.
- 일부 OS에서는 broadcast 송신에 `SO_BROADCAST`가 필요하므로 명시적으로 설정했습니다.
- 시스템 클립보드는 플랫폼 의존성이 크므로, terminal-native clipboard 표준인 OSC52를 함께 사용합니다.

## 영향 범위

- `internal/discovery`
- `internal/tui/model.go`
- `internal/tui/view.go`
- `README.md`
- `docs/architecture/lan-match.md`

## 검증 방법

- `go test ./... -count=1`
- `go build ./...`
- `GOOS=linux GOARCH=amd64 go test -c ./internal/discovery -o /tmp/discovery-linux.test`
- `GOOS=linux GOARCH=amd64 go test -c ./internal/tui -o /tmp/tui-linux.test`
- PTY 두 개에서 Host 대기 후 다른 TUI의 열린 방 표시 확인
- Host 주소 Copy 클릭 후 `pbpaste` 확인

## 남아 있는 한계

- 학교/회사 Wi-Fi, 게스트 네트워크, AP isolation, 방화벽이 UDP broadcast를 막으면 discovery가 제한될 수 있습니다.
- OSC52는 터미널 설정에 따라 차단될 수 있습니다. 이 경우 화면의 주소를 수동 복사해야 합니다.

## 후속 과제

- 실제 두 물리 장치와 여러 AP 환경에서 UDP discovery 추가 검증
- discovery 실패 시 사용자가 직접 입력할 주소를 더 눈에 띄게 안내
