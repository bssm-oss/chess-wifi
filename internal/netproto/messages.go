package netproto

import "github.com/bssm-oss/chess-wifi/internal/game"

const ProtocolVersion = "1"

const (
	TypeHello      = "hello"
	TypeWelcome    = "welcome"
	TypeSnapshot   = "snapshot"
	TypeMoveIntent = "move_intent"
	TypeAction     = "action_intent"
	TypeError      = "error"
	TypePing       = "ping"
)

type Envelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

type Hello struct {
	ProtocolVersion string `json:"protocol_version"`
	PlayerName      string `json:"player_name"`
}

type Welcome struct {
	ProtocolVersion string      `json:"protocol_version"`
	Self            game.Player `json:"self"`
	Peer            game.Player `json:"peer"`
}

type Snapshot struct {
	State game.Snapshot `json:"state"`
}

type MoveIntent struct {
	ExpectedVersion int    `json:"expected_version"`
	MoveUCI         string `json:"move_uci"`
}

type ActionIntent struct {
	ExpectedVersion int    `json:"expected_version"`
	Action          string `json:"action"`
}

type Error struct {
	Message string        `json:"message"`
	State   game.Snapshot `json:"state,omitempty"`
}

type Ping struct{}
