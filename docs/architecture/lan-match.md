# LAN match architecture

## 개요

`chess-wifi`는 중앙 서버 없이 두 프로세스가 직접 TCP 연결하는 구조를 사용합니다. Host가 단일 권위(authority) 역할을 맡고, Join 측은 의도(intent)만 전송합니다.

## 구성 요소

- `internal/cli`: `chess-wifi match` 명령 정의
- `internal/tui`: 화면 전환, 보드 렌더링, 마우스/키보드 입력
- `internal/discovery`: Host 대기 상태를 UDP broadcast로 알리고 Join 화면에서 열린 매치를 찾음
- `internal/session`: Host/Join 연결, heartbeat, 프로토콜 송수신
- `internal/netproto`: JSON envelope 및 메시지 구조
- `internal/game`: 체스 스냅샷과 이동 검증 보조 로직

## 연결 모델

1. Host가 TCP 리스너를 띄움
2. Host는 사설 IPv4 주소 목록을 UI에 표시
3. Host는 UDP `18787` 로 discovery announcement를 주기적으로 broadcast
4. Join 화면은 열린 매치를 스캔해서 첫 화면과 Join 화면에 표시
5. Join은 discovery 목록에서 선택하거나 `IP:PORT`를 직접 입력해 연결
6. `hello` / `welcome` / `snapshot` 순서로 핸드셰이크

## Discovery 모델

- 중앙 서버 없음
- UDP broadcast만 사용
- announcement에는 서비스명, discovery 프로토콜 버전, Host 이름, TCP 매치 포트만 포함
- Join 측은 UDP 패킷의 source IP와 announcement의 TCP 포트를 조합해 접속 주소를 만듦
- discovery가 실패해도 수동 `IP:PORT` 입력 경로는 유지

## 상태 동기화 모델

- Host가 합법 수를 검증
- Client는 `move_intent` 또는 `action_intent`만 보냄
- Host는 성공 시 전체 `snapshot`을 다시 전송
- Client는 스냅샷을 받아 UI를 갱신

## 왜 이 구조를 택했는가

- LAN 환경에서 가장 단순하고 예측 가능함
- mDNS보다 단순한 UDP broadcast가 디버깅과 수동 fallback에 유리함
- 체스 상태가 작아서 전체 스냅샷 비용이 낮음
- 양쪽 보드 상태가 어긋나기 어렵고 로그 추적이 쉬움

## 실패 처리

- heartbeat와 read deadline으로 반쯤 끊긴 연결을 감지
- 연결이 끊기면 UI를 에러 상태로 전환
- 자동 재연결은 v1 범위 밖
