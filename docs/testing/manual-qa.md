# Manual QA

## 목표

자동 테스트만으로 부족한 CLI/TUI 실제 흐름을 실행 결과 기준으로 남깁니다. 아래 내용은 이번 작업에서 실제로 실행한 명령과 관찰 결과입니다.

## 실행 환경

- macOS (darwin)
- Go 1.26.1
- PTY 기반 로컬 루프백 시뮬레이션 사용

## 실제 실행 기록

### 1. 도움말 확인

실행 명령:

```bash
go run ./cmd/chess-wifi --help
go run ./cmd/chess-wifi match --help
```

관찰 결과:

- 루트 명령에서 `match` 서브커맨드가 노출됨
- `match` 명령 설명이 정상 출력됨

### 2. 자동 검증

실행 명령:

```bash
go test ./...
go build ./...
go test ./internal/tui -run TestHandleMouseSelectsPieceFromRenderedBoard -v
```

관찰 결과:

- 전체 테스트 통과
- 전체 빌드 통과
- 클릭 가능한 보드 선택 경로 검증 테스트 통과

테스트 출력 요약:

```text
=== RUN   TestHandleMouseSelectsPieceFromRenderedBoard
--- PASS: TestHandleMouseSelectsPieceFromRenderedBoard (0.00s)
PASS
```

### 3. Host 대기 화면 진입과 Join 연결

실행 방법:

- `/tmp/chess-wifi match` 바이너리를 PTY 두 개에서 동시에 실행
- Host는 기본 포트 `8787`로 대기
- Guest는 `127.0.0.1:8787` 로 Join

관찰 결과:

- Host 대기 화면에서 `대기 시간` 메시지 확인
- Host 화면에서 `Guest connected. White moves first.` 확인
- Guest 화면에서 `Connected to Host.` 확인
- 양쪽 모두 `LAN Match` 화면으로 전환됨

실행 출력 요약:

```text
HOST_WAITING=True
HOST_CONNECTED=True
GUEST_CONNECTED=True
HOST_SNIP=...Guest connected. White moves first.
GUEST_SNIP=...LAN Match...
```

## 해석

- 비인터랙티브 파이프 환경에서는 Bubble Tea가 `/dev/tty` 를 요구하므로 단순 파이프 대신 PTY 기반 검증이 필요했습니다.
- 실제 마우스 입력 전체를 터미널 escape sequence로 재현하는 대신, 클릭 가능한 선택 로직은 TUI 테스트로 검증했고, Host/Join 실사용 흐름은 PTY smoke test로 검증했습니다.

## 남은 수동 QA 권장 항목

- 실제 같은 Wi-Fi의 두 장치에서 Host/Join 검증
- 마우스로 실제 말 이동 후 상대 보드 반영 확인
- `r` 기권 결과가 양쪽 화면에 반영되는지 확인
