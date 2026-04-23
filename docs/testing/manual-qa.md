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
go test ./internal/tui -run TestMouseClickMoveSyncsAcrossPeerSessions -v
```

관찰 결과:

- 전체 테스트 통과
- 전체 빌드 통과
- 클릭 가능한 보드 선택 경로 검증 테스트 통과
- 마우스 클릭으로 `e2e4` 수를 전송하고 peer 세션까지 동기화하는 TUI 통합 테스트 통과

테스트 출력 요약:

```text
=== RUN   TestHandleMouseSelectsPieceFromRenderedBoard
--- PASS: TestHandleMouseSelectsPieceFromRenderedBoard (0.00s)
PASS

=== RUN   TestMouseClickMoveSyncsAcrossPeerSessions
--- PASS: TestMouseClickMoveSyncsAcrossPeerSessions (0.00s)
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

### 4. 세션 종료 안정성 회귀 확인

실행 방법:

- `/tmp/chess-wifi-local match` 바이너리를 PTY 두 개에서 동시에 실행
- Host는 기본 포트 `8787`로 대기
- Guest는 `127.0.0.1:8787` 로 Join
- 연결 후 Host에서 `q` 입력

관찰 결과:

- Host 프로세스가 panic 없이 정상 종료됨
- Guest 화면이 `상태 알림` / `connection closed` 로 전환됨
- 이전에 재현되던 `send on closed channel` panic이 더 이상 발생하지 않음

실행 출력 요약:

```text
HOST_EXIT=0
GUEST_CLOSE_SCREEN=True
GUEST_SNIP=...connection closed...
PANIC=False
```

### 5. Discovery 목록과 설치 스크립트 확인

실행 명령:

```bash
go build -o /tmp/chess-wifi-discovery ./cmd/chess-wifi
CHESS_WIFI_INSTALL_DIR=/tmp/chess-wifi-install-test ./install.sh
/tmp/chess-wifi-install-test/chess-wifi --help
```

관찰 결과:

- 설치 스크립트가 지정한 디렉터리에 `chess-wifi` 바이너리를 설치함
- 설치된 바이너리의 `--help` 출력이 정상 표시됨
- `/tmp/chess-wifi-discovery match`를 PTY 두 개에서 실행하고 한쪽을 Host 대기 상태로 만들면, 다른 쪽 첫 화면에 열린 매치가 자동 표시됨
- discovery 목록에서 `Join Host · <주소>` 항목을 선택하면 Guest가 즉시 연결되고 Host 화면이 매치 화면으로 전환됨

실행 출력 요약:

```text
INSTALL_OK=True
HELP_OK=True
DISCOVERY_VISIBLE=True
DISCOVERED_ITEM=Join Host · 10.129.57.46:8787
JOIN_FROM_DISCOVERY=True
HOST_CONNECTED=True
GUEST_CONNECTED=True
```

### 6. 마우스 조작, 주소 복사, compact layout 확인

실행 명령:

```bash
go build -o /tmp/chess-wifi-ui-fix ./cmd/chess-wifi
/tmp/chess-wifi-ui-fix match
```

실행 방법:

- `/tmp/chess-wifi-ui-fix match`를 PTY 두 개에서 실행
- 첫 번째 TUI에서 마우스로 `Host a match` 클릭
- Host 설정 화면에서 마우스로 `[ Start hosting ]` 클릭
- Host 대기 화면에서 주소 옆 `[ Copy ]` 클릭
- 두 번째 TUI 첫 화면에서 discovery로 표시된 `Join Host · <주소>` 항목을 마우스로 클릭
- Host 보드에서 마우스로 `e2`를 클릭한 뒤 `e4`를 클릭
- Guest 화면에서 `[ Quit ]` 버튼을 마우스로 클릭

관찰 결과:

- `80x24` PTY에서 match 화면이 compact layout으로 표시되어 보드와 상태 패널이 잘리지 않음
- Host 주소 복사 후 `복사됨: <주소>` 메시지가 표시됨
- Host 주소 복사 후 `pbpaste` 결과가 `10.0.0.6:8787` 로 확인됨
- discovery 방 목록 클릭으로 Guest가 즉시 연결됨
- 마우스 클릭으로 `e2e4` 수가 적용되고 Guest 화면에도 동기화됨
- `[ Quit ]` 클릭으로 한쪽이 종료되고 상대는 `connection closed` 화면으로 전환됨

실행 출력 요약:

```text
COMPACT_LAYOUT_VISIBLE=True
COPY_MESSAGE=True
PBPASTE=10.0.0.6:8787
DISCOVERY_MOUSE_JOIN=True
MOUSE_MOVE_E2E4=True
MOUSE_QUIT=True
```

### 7. Discovery query와 OSC52 복사 fallback 확인

실행 명령:

```bash
go build -o /tmp/chess-wifi-discovery-fix ./cmd/chess-wifi
/tmp/chess-wifi-discovery-fix match
pbpaste
```

관찰 결과:

- Host 대기 후 다른 TUI에서 `Join Host · 10.129.57.46:8787` 항목이 표시됨
- Host 주소 `[ Copy ]` 클릭 시 OSC52 sequence가 출력에 포함됨
- `pbpaste` 결과가 `10.0.0.6:8787` 로 확인됨

실행 출력 요약:

```text
DISCOVERY_QUERY_VISIBLE=True
OSC52_SEQUENCE_VISIBLE=True
SYSTEM_CLIPBOARD_VALUE=10.0.0.6:8787
```

## 해석

- 비인터랙티브 파이프 환경에서는 Bubble Tea가 `/dev/tty` 를 요구하므로 단순 파이프 대신 PTY 기반 검증이 필요했습니다.
- 실제 마우스 입력 전체를 터미널 escape sequence로 재현하는 대신, 클릭 가능한 선택 로직은 TUI 테스트로 검증했고, Host/Join 실사용 흐름은 PTY smoke test로 검증했습니다.

## 남은 수동 QA 권장 항목

- 실제 같은 Wi-Fi의 두 장치에서 Host/Join 검증
- `r` 기권 결과가 양쪽 화면에 반영되는지 확인
- 네트워크 정책이 다른 Wi-Fi/AP에서 UDP discovery `18787` 이 차단되지 않는지 확인
- OS/터미널별 클립보드 권한 차이 확인

## 현재 한계

- 이번 세션은 단일 개발 환경과 PTY 기반 시뮬레이션에서 수행되었기 때문에, 두 대의 실제 물리 장치가 같은 Wi-Fi에 붙은 상태의 검증은 별도로 다시 수행하는 것이 가장 안전합니다.
