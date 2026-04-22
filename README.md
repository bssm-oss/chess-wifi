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
- Bubble Tea 기반 TUI
- 마우스 클릭과 키보드 둘 다 지원
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

## 가장 쉬운 설치 방법

릴리스가 배포된 뒤에는 아래 명령 하나로 바로 설치할 수 있습니다.

```bash
go install github.com/bssm-oss/chess-wifi/cmd/chess-wifi@latest
```

설치 후 실행 파일 경로가 `PATH`에 잡혀 있으면 바로 아래처럼 실행할 수 있습니다.

```bash
chess-wifi match
```

## 로컬 개발 설치

```bash
git clone https://github.com/bssm-oss/chess-wifi.git
cd chess-wifi
go mod tidy
go run ./cmd/chess-wifi match
```

## 사용 방법

### 1. Host 측

```bash
chess-wifi match
```

1. `Host a match` 선택
2. 이름과 포트 입력
3. 화면에 표시된 `192.168.x.x:8787` 같은 주소를 상대에게 전달

### 2. Join 측

```bash
chess-wifi match
```

1. `Join a match` 선택
2. 이름 입력
3. Host가 알려준 `IP:PORT` 입력
4. 연결되면 체스판이 열림

## 조작 방법

- 마우스 왼쪽 클릭: 말 선택 / 이동
- 방향키 또는 `h j k l`: 커서 이동
- `Enter` / `Space`: 선택 / 이동 확정
- `Esc`: 현재 선택 취소
- `r`: 기권
- `q`: 프로그램 종료

## 테스트 실행 방법

```bash
go test ./...
```

빌드 확인:

```bash
go build ./...
```

## 주요 디렉터리 구조

```text
cmd/chess-wifi/        실제 CLI 엔트리포인트
internal/cli/          Cobra 명령 구성
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
- 자동 LAN discovery 없음
- 재연결 / 이어두기 없음
- 저장/불러오기 없음
- 로컬 네트워크에서 포트 접근이 막히면 연결되지 않음

## 로드맵

- 선택적 LAN discovery (기본 연결 방식은 유지)
- 더 풍부한 상태 표시
- 패키징과 릴리스 자동화 강화
