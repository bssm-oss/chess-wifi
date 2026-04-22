# 2026-04-22 chess-wifi bootstrap

## 배경

원격 저장소가 비어 있었기 때문에, 기존 구현을 확장하는 작업이 아니라 프로젝트 자체를 처음부터 부트스트랩해야 했습니다. 동시에 사용자는 Go 기반 CLI, 같은 Wi-Fi에서의 직접 연결, 마우스 클릭이 가능한 TUI, 문서화, 테스트, CI까지 포함한 릴리스 품질을 요구했습니다.

## 문제 또는 목표

- `chess-wifi match` 한 명령으로 LAN 체스를 시작할 수 있어야 함
- 중앙 서버 없이 동작해야 함
- UI와 네트워크, 체스 규칙을 분리한 유지보수 가능한 구조가 필요함

## 변경 내용

- Go 모듈 초기화
- Cobra 기반 CLI 추가
- Bubble Tea / Lip Gloss 기반 TUI 추가
- Host/Join TCP 세션 계층 추가
- 체스 스냅샷/프로토콜 구조 추가
- 자동 테스트와 GitHub Actions CI 추가
- README, AGENTS, 아키텍처/테스트 문서 추가

## 설계 이유

- v1에서는 자동 discovery보다 수동 `IP:PORT` 입력이 더 단순하고 실패 원인이 적습니다.
- Host authoritative snapshot 구조를 택해 두 플레이어 상태가 어긋나는 문제를 줄였습니다.
- 체스 상태는 작기 때문에 패치보다 전체 스냅샷 재전송이 디버깅과 유지보수에 유리합니다.

## 영향 범위

- CLI 엔트리포인트
- LAN 세션 프로토콜
- TUI 렌더링 및 마우스 입력
- 문서와 CI

## 검증 방법

- `go test ./...`
- `go build ./...`
- CLI 도움말과 인터랙티브 실행 스모크 테스트

## 남아 있는 한계

- 자동 LAN discovery 없음
- 재연결 없음
- 인터넷 환경 지원 없음
- 저장/불러오기 없음

## 후속 과제

- 선택적 discovery 추가 여부 평가
- 릴리스 아티팩트 자동 배포
- 수동 QA 시나리오를 CI-friendly smoke test로 보강
