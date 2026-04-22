# Manual QA

## 목표

자동 테스트만으로 확인하기 어려운 실제 CLI/TUI 흐름을 점검합니다.

## 시나리오

### 시나리오 1: 도움말 확인

```bash
go run ./cmd/chess-wifi --help
go run ./cmd/chess-wifi match --help
```

기대 결과:

- 루트 명령과 `match` 명령 설명이 출력된다.

### 시나리오 2: Host 대기 화면 진입

```bash
go run ./cmd/chess-wifi match
```

기대 결과:

- 메뉴가 보인다.
- Host 선택 후 이름/포트를 입력할 수 있다.
- 시작하면 로컬 주소 목록이 표시된다.

### 시나리오 3: 로컬 루프백 연결

두 터미널에서 각각 실행한다.

터미널 A:

```bash
go run ./cmd/chess-wifi match
```

터미널 B:

```bash
go run ./cmd/chess-wifi match
```

기대 결과:

- A는 Host 주소를 표시한다.
- B는 `127.0.0.1:8787` 또는 표시된 LAN 주소로 연결 가능하다.
- 연결 후 양쪽 모두 체스판이 렌더링된다.
- Host의 첫 수가 Join 화면에 반영된다.

### 시나리오 4: 입력 검증

- 잘못된 주소 입력 시 에러 화면으로 전환되는지 확인
- 자기 차례가 아닐 때 말을 움직일 수 없는지 확인
- `r` 기권이 양쪽 결과에 반영되는지 확인
