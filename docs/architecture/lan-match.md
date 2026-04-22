# LAN match architecture

## 개요

`chess-wifi`는 중앙 서버 없이 두 프로세스가 직접 TCP 연결하는 구조를 사용합니다. Host가 단일 권위(authority) 역할을 맡고, Join 측은 의도(intent)만 전송합니다.

## 구성 요소

- `internal/cli`: `chess-wifi match` 명령 정의
- `internal/tui`: 화면 전환, 보드 렌더링, 마우스/키보드 입력
- `internal/session`: Host/Join 연결, heartbeat, 프로토콜 송수신
- `internal/netproto`: JSON envelope 및 메시지 구조
- `internal/game`: 체스 스냅샷과 이동 검증 보조 로직

## 연결 모델

1. Host가 TCP 리스너를 띄움
2. Host는 사설 IPv4 주소 목록을 UI에 표시
3. Join은 `IP:PORT`로 직접 연결
4. `hello` / `welcome` / `snapshot` 순서로 핸드셰이크

## 상태 동기화 모델

- Host가 합법 수를 검증
- Client는 `move_intent` 또는 `action_intent`만 보냄
- Host는 성공 시 전체 `snapshot`을 다시 전송
- Client는 스냅샷을 받아 UI를 갱신

## 왜 이 구조를 택했는가

- LAN 환경에서 가장 단순하고 예측 가능함
- mDNS/브로드캐스트보다 디버깅이 쉬움
- 체스 상태가 작아서 전체 스냅샷 비용이 낮음
- 양쪽 보드 상태가 어긋나기 어렵고 로그 추적이 쉬움

## 실패 처리

- heartbeat와 read deadline으로 반쯤 끊긴 연결을 감지
- 연결이 끊기면 UI를 에러 상태로 전환
- 자동 재연결은 v1 범위 밖
