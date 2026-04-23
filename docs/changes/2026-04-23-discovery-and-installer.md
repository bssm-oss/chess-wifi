# 2026-04-23 discovery and installer

## 배경

기존 설치 안내는 `go install` 이후 `GOPATH/bin` 또는 `GOBIN`이 `PATH`에 잡혀 있어야만 `chess-wifi` 명령을 바로 사용할 수 있었습니다. 또한 `chess-wifi match`는 Host 주소를 수동으로 입력하는 흐름만 제공해, 같은 Wi-Fi에서 현재 열린 매치를 바로 보기 어렵습니다.

## 문제 또는 목표

- 설치 직후 `chess-wifi match`를 바로 실행할 수 있는 설치 경로 제공
- `chess-wifi match` 첫 화면에서 현재 LAN에 열린 매치 표시
- 중앙 서버 없이 기존 직접 TCP 연결 모델 유지
- discovery가 실패해도 수동 `IP:PORT` 입력 경로 유지

## 변경 내용

- 루트 `install.sh` 추가
- Host 대기 중 UDP discovery announcement 송신
- TUI 첫 화면과 Join 화면에 `열려있는 LAN 매치` 목록 추가
- 발견된 매치를 선택하면 Join 주소를 자동으로 채우고 바로 연결
- discovery 패키지와 TUI 회귀 테스트 추가
- README, AGENTS, 아키텍처 문서 갱신

## 설계 이유

- `go install`은 설치 위치를 바꿔도 사용자의 현재 shell `PATH`를 안전하게 수정할 수 없으므로, `/usr/local/bin`에 설치하는 스크립트를 제공하는 편이 실제 사용성이 높습니다.
- discovery는 UDP broadcast만 사용해 중앙 서버, 계정, 외부 API를 추가하지 않습니다.
- TCP 게임 세션은 기존 Host authoritative 구조를 그대로 유지합니다.

## 영향 범위

- 설치 경로: `install.sh`
- discovery: `internal/discovery`
- Host lifecycle: `internal/session`
- TUI 메뉴/Join 화면: `internal/tui`
- 문서: `README.md`, `AGENTS.md`, `docs/architecture/lan-match.md`

## 검증 방법

- `go test ./...`
- `go build ./...`
- `CHESS_WIFI_INSTALL_DIR=/tmp/chess-wifi-install ./install.sh`
- PTY 두 개에서 Host를 띄운 뒤 다른 `chess-wifi match` 첫 화면에 열린 매치가 나타나는지 확인

## 남아 있는 한계

- UDP broadcast가 네트워크 정책이나 방화벽에서 막히면 자동 목록에 표시되지 않을 수 있습니다.
- discovery는 같은 LAN/Wi-Fi 범위만 대상으로 하며 인터넷 매치메이킹은 지원하지 않습니다.

## 후속 과제

- 실제 두 물리 장치와 여러 AP 환경에서 discovery 안정성 추가 검증
- 릴리스 아티팩트 자동 배포 추가 여부 검토
