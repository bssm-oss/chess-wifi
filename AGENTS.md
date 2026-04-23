# AGENTS.md

## 프로젝트 목적

`chess-wifi`는 같은 Wi-Fi 안에서 두 사용자가 중앙 서버 없이 직접 연결해서 체스를 두는 Go CLI/TUI 프로젝트입니다. 핵심 사용자 흐름은 `chess-wifi match`를 실행한 뒤 Host/Join을 선택하고, 연결 후 마우스로 보드를 클릭해 플레이하는 것입니다.

## 빠른 시작

```bash
go mod tidy
go test ./...
go build ./...
go run ./cmd/chess-wifi match
CHESS_WIFI_INSTALL_DIR=/tmp/chess-wifi-bin ./install.sh
```

## 설치 / 실행 / 테스트 명령

- 의존성 정리: `go mod tidy`
- 테스트: `go test ./...`
- 빌드: `go build ./...`
- 실행: `go run ./cmd/chess-wifi match`
- 설치 스크립트 검증: `CHESS_WIFI_INSTALL_DIR=/tmp/chess-wifi-bin ./install.sh`
- 설치형 사용: `curl -fsSL https://raw.githubusercontent.com/bssm-oss/chess-wifi/main/install.sh | sh`
- Go 직접 설치: `go install github.com/bssm-oss/chess-wifi/cmd/chess-wifi@latest`

## 기본 작업 순서

1. README, AGENTS, docs를 먼저 읽습니다.
2. 변경 범위를 정리합니다.
3. 최소 수정 설계를 고릅니다.
4. 구현 후 테스트를 추가하거나 수정합니다.
5. 문서를 실제 동작에 맞게 갱신합니다.
6. `go test ./...` 와 `go build ./...` 를 실행합니다.

## 완료 조건

- 요청 기능이 실제 코드에 반영됨
- 관련 테스트가 존재하고 실행됨
- `go test ./...` 통과
- `go build ./...` 통과
- README/AGENTS/docs가 실제 상태와 일치

## 코드 스타일 원칙

- UI, 세션, 게임 규칙을 섞지 않습니다.
- 숨은 전역 상태보다 명시적 구조를 선호합니다.
- 네트워크 프로토콜은 버전 필드를 유지합니다.
- discovery 프로토콜도 버전 필드를 유지하고 중앙 서버를 추가하지 않습니다.
- 랜덤한 리팩터링 대신 범위에 맞는 최소 수정만 합니다.

## 파일 구조 원칙

- `cmd/`에는 CLI 엔트리포인트만 둡니다.
- `internal/cli`는 Cobra 명령 정의만 담당합니다.
- `internal/discovery`는 UDP LAN discovery만 담당합니다.
- `internal/game`은 체스 상태/규칙 관련 보조 로직을 가집니다.
- `internal/session`은 TCP 세션과 동기화를 담당합니다.
- `internal/tui`는 Bubble Tea 렌더링과 입력을 담당합니다.

## 문서화 원칙

- 의미 있는 변경이 있으면 `README.md`, `AGENTS.md`, `docs/` 중 하나 이상을 갱신합니다.
- 문서에는 배경, 변경 내용, 검증 방법, 제한 사항을 남깁니다.
- 문서와 실제 명령이 다르면 문서를 즉시 수정합니다.

## 테스트 원칙

- 버그 수정이면 회귀 테스트를 먼저 고려합니다.
- 세션 프로토콜, 게임 규칙, 보드 입력 매핑은 자동 테스트 우선입니다.
- TUI는 가능한 범위에서 순수 함수/매핑 로직을 분리해 테스트합니다.

## 브랜치 / 커밋 / PR 규칙

- 기본 브랜치에서 직접 작업하지 않습니다.
- 브랜치 형식: `feat/...`, `fix/...`, `docs/...`, `test/...`
- 커밋은 목적 단위로 나눕니다.
- PR에는 배경, 변경 내용, 테스트 결과, 수동 검증 결과를 적습니다.

## 민감한 경로 / 주의 경로

- `internal/session/`: 프로토콜 호환성에 영향이 큽니다.
- `internal/discovery/`: LAN 자동 발견과 UDP 포트 `18787` 동작에 영향이 큽니다.
- `internal/game/`: 합법 수 판정과 스냅샷 구조에 영향이 큽니다.
- `internal/tui/`: 마우스 좌표 매핑이 쉽게 깨질 수 있습니다.

## 작업 전 체크리스트

- [ ] 요청 범위를 한 문장으로 설명할 수 있는가
- [ ] 수정할 패키지와 이유를 알고 있는가
- [ ] 테스트/검증 방법을 알고 있는가

## 작업 후 체크리스트

- [ ] `go test ./...`
- [ ] `go build ./...`
- [ ] README/AGENTS/docs 반영
- [ ] 변경 이유와 한계 정리

## 절대 하면 안 되는 것

- 실행하지 않은 테스트를 통과했다고 말하지 않기
- 문서와 실제 동작을 다르게 두기
- 중앙 서버나 외부 인프라를 몰래 추가하기
- discovery를 이유로 인터넷 릴레이, 계정, 외부 API를 추가하기
- 타입/에러를 무시하는 임시 처리 남기기
- 사용자 요청과 무관한 기능 확장하기
