# chess-wifi

`chess-wifi`는 같은 Wi-Fi에 접속한 두 사람이 중앙 서버 없이 직접 연결해서 체스를 둘 수 있게 만드는 Go 기반 CLI 도구입니다. 핵심 명령은 `chess-wifi match` 하나이며, 실행하면 터미널 안에서 Host 또는 Join을 선택하고 마우스로 체스판을 클릭해 플레이할 수 있습니다.

## 무엇을 해결하나요?

- 같은 공간, 같은 네트워크에서 빠르게 1:1 체스를 두고 싶을 때
- 별도 계정, 중앙 서버, 웹 브라우저 없이 바로 실행하고 싶을 때
- 터미널에서도 보기 좋고 조작 가능한 체스 UI가 필요할 때

## 핵심 기능

- Go 단일 바이너리 CLI
- `chess-wifi match` 인터랙티브 실행
- 같은 Wi-Fi에서 직접 TCP 연결
- Host가 자신의 LAN 주소를 표시하고 Join이 직접 접속
- Host 대기 중 UDP LAN discovery로 열린 매치를 자동 표시
- Bubble Tea 기반 TUI
- 메뉴, 방 목록, 주소 복사, 체스판 이동, 기권/종료를 마우스 클릭으로 조작
- 작은 터미널에서는 체스판을 compact layout으로 자동 전환
- 체스 규칙은 `github.com/notnil/chess`로 검증
- 호스트 권위(authoritative host) 스냅샷 동기화로 양쪽 보드 일관성 유지

## 기술 스택

- Go 1.26+
- Cobra
- Bubble Tea
- Bubbles
- Lip Gloss
- notnil/chess
- GitHub Actions

## 요구 사항

- Go 1.26 이상
- 같은 LAN/Wi-Fi에 있는 두 대의 장치, 또는 로컬 테스트용 두 터미널
- TCP 포트 `8787` 접근 가능 환경
- 자동 매치 목록을 쓰려면 UDP discovery 포트 `18787` 접근 가능 환경

## 가장 쉬운 설치 방법

PATH 설정 없이 바로 `chess-wifi` 명령을 쓰고 싶다면 설치 스크립트를 사용합니다. 기본 설치 위치는 `/usr/local/bin/chess-wifi` 입니다.

```bash
curl -fsSL https://raw.githubusercontent.com/bssm-oss/chess-wifi/main/install.sh | sh
```

설치 후 바로 실행합니다.

```bash
chess-wifi match
```

설치 위치를 직접 지정하고 싶으면 `CHESS_WIFI_INSTALL_DIR`를 사용합니다.

```bash
CHESS_WIFI_INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://raw.githubusercontent.com/bssm-oss/chess-wifi/main/install.sh)"
```

Go 기본 설치 방식을 선호한다면 아래 명령도 사용할 수 있습니다. 이 경우 Go의 `GOBIN` 또는 `GOPATH/bin`이 `PATH`에 있어야 합니다.

```bash
go install github.com/bssm-oss/chess-wifi/cmd/chess-wifi@latest
```

## 로컬 개발 설치

```bash
git clone https://github.com/bssm-oss/chess-wifi.git
cd chess-wifi
go mod tidy
go run ./cmd/chess-wifi match
```

## 사용 방법

### 1. Host가 방 만들기

```bash
chess-wifi match
```

1. `Host a match` 선택
2. 이름과 포트 입력
3. 화면에 표시된 `192.168.x.x:8787` 같은 주소를 확인
4. 주소 옆 `[ Copy ]`를 클릭하거나 `c`를 눌러 첫 주소 복사
5. 상대가 자동 discovery 목록에서 선택할 때까지 대기
6. 자동 discovery가 안 되면 복사한 주소를 상대에게 직접 전달

Host가 대기 중이면 다른 사용자의 첫 화면에 아래와 같은 항목이 자동으로 나타납니다.

```text
열려있는 LAN 매치
  1. Host · 10.129.57.46:8787 · 0s 전
목록에서 Enter를 누르면 바로 연결합니다.
```

### 2. Join이 자동 목록으로 참가하기

```bash
chess-wifi match
```

1. 첫 화면의 `열려있는 LAN 매치` 목록에서 Host를 선택해 바로 연결
2. 방향키로 `Join Host · <주소>` 항목 선택
3. `Enter`
4. 연결되면 체스판이 열림

### 3. 자동 목록이 안 보일 때 직접 참가하기

```bash
chess-wifi match
```

1. `Join by address` 선택
2. 이름 입력
3. Host 화면에 표시된 `IP:PORT` 입력
4. 연결되면 체스판이 열림

자동 목록은 UDP broadcast를 사용합니다. 학교/회사 Wi-Fi, 게스트 네트워크, 방화벽, AP isolation 설정이 UDP `18787`을 막으면 목록에 안 뜰 수 있습니다. 이 경우에도 직접 주소 입력은 계속 사용할 수 있습니다.

## 조작 방법

- 마우스 왼쪽 클릭: 메뉴 선택, 방 참가, 입력 필드 포커스, Host 시작, 주소 복사, 말 선택 / 이동, 기권, 종료
- 방향키 또는 `h j k l`: 커서 이동
- `Enter` / `Space`: 선택 / 이동 확정
- `Esc`: 현재 선택 취소
- `r`: 기권
- `q`: 프로그램 종료
- Host 대기 화면에서 `c`: 첫 번째 Host 주소 복사

## 화면 크기

기본 터미널 크기인 `80x24`에서도 보드가 잘리지 않도록, 화면이 좁거나 낮으면 자동으로 compact layout을 사용합니다. 넓은 터미널에서는 기존처럼 더 큰 보드와 상세 사이드바가 표시됩니다.

## 테스트 실행 방법

```bash
go test ./...
```

빌드 확인:

```bash
go build ./...
```

설치 스크립트 검증:

```bash
CHESS_WIFI_INSTALL_DIR=/tmp/chess-wifi-install-test ./install.sh
/tmp/chess-wifi-install-test/chess-wifi --help
```

## 검증된 사용자 플로우

이번 릴리스 흐름에서 실제로 확인한 동작입니다.

```text
1. /tmp/chess-wifi-discovery match 를 PTY 두 개에서 실행
2. 첫 번째 TUI에서 마우스로 Host a match 선택
3. 마우스로 Start hosting 클릭
4. Host 대기 화면에서 주소 옆 Copy 클릭
5. 두 번째 TUI 첫 화면에 Join Host · 10.129.57.46:8787 표시 확인
6. discovery 목록을 마우스로 클릭
7. Guest 화면: Connected to Host.
8. Host 화면: Guest connected. White moves first.
9. Host 보드에서 마우스로 e2 선택 후 e4 클릭
10. 양쪽 화면에 e2e4와 Black to move 표시
11. Guest에서 Quit 버튼 클릭
12. Host 화면: connection closed
```

이전 검증 흐름:

```text
1. /tmp/chess-wifi-discovery match 를 PTY 두 개에서 실행
2. 첫 번째 TUI에서 Host a match 선택
3. 두 번째 TUI 첫 화면에 Join Host · 10.129.57.46:8787 표시 확인
4. discovery 목록에서 Enter
5. Guest 화면: Connected to Host.
6. Host 화면: Guest connected. White moves first.
7. Host에서 q 입력
8. Guest 화면: connection closed
```

실행한 검증 명령:

```bash
go test ./... -count=1
go build ./...
GOOS=linux GOARCH=amd64 go test -c ./internal/discovery -o /tmp/discovery-linux.test
CHESS_WIFI_INSTALL_DIR=/tmp/chess-wifi-install-test ./install.sh
GOBIN=/tmp/chess-wifi-gobin-final go install github.com/bssm-oss/chess-wifi/cmd/chess-wifi@latest
```

검증 결과:

```text
go test ./...                         pass
go build ./...                        pass
GitHub Actions test                   pass
CodeRabbit                            pass
go install ...@latest                 pass
설치 스크립트 smoke test              pass
PTY 2개 실제 TUI Host/Join discovery   pass
```

## 주요 디렉터리 구조

```text
cmd/chess-wifi/        실제 CLI 엔트리포인트
internal/cli/          Cobra 명령 구성
internal/discovery/    UDP 기반 LAN 매치 발견
internal/game/         체스 상태와 규칙 보조 로직
internal/lan/          사설 IPv4 주소 탐지
internal/netproto/     JSON 기반 세션 프로토콜
internal/session/      Host/Join TCP 세션 관리
internal/tui/          Bubble Tea UI
docs/                  아키텍처/변경/테스트 문서
```

## 아키텍처 개요

- 중앙 서버 없음
- 같은 Wi-Fi 안에서 Host와 Join이 직접 TCP 연결
- Host 대기 중에는 UDP broadcast로 매치 존재를 알림
- Host가 체스 상태의 단일 기준(source of truth)
- Client는 move intent만 전송
- Host가 수를 검증한 뒤 전체 스냅샷을 다시 전송

이 구조 덕분에 LAN 환경에서 구현 복잡도를 크게 늘리지 않으면서도 양쪽 상태가 쉽게 어긋나지 않습니다.

## 개발 원칙

- 정확성 우선
- 테스트 가능한 구조 유지
- UI, 네트워크, 게임 규칙 분리
- 문서와 실제 동작 일치
- 과한 기능 추가보다 명확한 동작 우선

## CI 개요

GitHub Actions에서 다음을 확인합니다.

- `go test ./...`
- `go build ./...`
- `gofmt` 포맷 체크

## 기여 방법

1. 기능 브랜치를 만듭니다.
2. 코드, 테스트, 문서를 함께 수정합니다.
3. `go test ./...` 와 `go build ./...` 를 통과시킵니다.
4. PR에 배경, 변경 이유, 검증 결과를 남깁니다.

## 알려진 제한 사항

- NAT traversal 없음
- 인터넷 매치메이킹 없음
- 재연결 / 이어두기 없음
- 저장/불러오기 없음
- 로컬 네트워크에서 TCP `8787` 또는 UDP discovery `18787` 접근이 막히면 자동 표시 또는 연결이 제한될 수 있음

## 로드맵

- 더 풍부한 상태 표시
- 패키징과 릴리스 자동화 강화
