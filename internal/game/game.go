package game

import (
	"errors"
	"fmt"
	"strings"

	"github.com/notnil/chess"
)

type Side string

const (
	White Side = "white"
	Black Side = "black"
)

type Player struct {
	Name string `json:"name"`
	Side Side   `json:"side"`
}

type Snapshot struct {
	Version      int      `json:"version"`
	FEN          string   `json:"fen"`
	Turn         Side     `json:"turn"`
	Players      []Player `json:"players"`
	LastMoveUCI  string   `json:"last_move_uci,omitempty"`
	MoveHistory  []string `json:"move_history"`
	Status       string   `json:"status"`
	Result       string   `json:"result,omitempty"`
	Check        bool     `json:"check"`
	Message      string   `json:"message,omitempty"`
	BoardFlipped bool     `json:"board_flipped,omitempty"`
}

type Match struct {
	game       *chess.Game
	version    int
	players    []Player
	moveUCI    []string
	lastMove   string
	resultText string
	message    string
}

func New(hostName, guestName string) *Match {
	return &Match{
		game: chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		players: []Player{
			{Name: hostName, Side: White},
			{Name: guestName, Side: Black},
		},
	}
}

func SideFromColor(color chess.Color) Side {
	if color == chess.Black {
		return Black
	}
	return White
}

func ColorFromSide(side Side) chess.Color {
	if side == Black {
		return chess.Black
	}
	return chess.White
}

func Opponent(side Side) Side {
	if side == White {
		return Black
	}
	return White
}

func (m *Match) Snapshot() Snapshot {
	pos := m.game.Position()
	status := statusString(m.game.Outcome(), m.game.Method())
	result := resultString(m.game.Outcome())
	msg := m.message
	if msg == "" {
		msg = messageForState(m.game)
	}
	return Snapshot{
		Version:      m.version,
		FEN:          pos.String(),
		Turn:         SideFromColor(pos.Turn()),
		Players:      append([]Player(nil), m.players...),
		LastMoveUCI:  m.lastMove,
		MoveHistory:  append([]string(nil), m.moveUCI...),
		Status:       status,
		Result:       result,
		Check:        false,
		Message:      msg,
		BoardFlipped: false,
	}
}

func (m *Match) ApplyMoveUCI(uci string) (Snapshot, error) {
	uci = strings.TrimSpace(strings.ToLower(uci))
	move, err := chess.UCINotation{}.Decode(m.game.Position(), uci)
	if err != nil {
		return Snapshot{}, fmt.Errorf("decode move %q: %w", uci, err)
	}
	if err := m.game.Move(move); err != nil {
		return Snapshot{}, fmt.Errorf("apply move %q: %w", uci, err)
	}
	m.version++
	m.lastMove = uci
	m.moveUCI = append(m.moveUCI, uci)
	m.message = messageForState(m.game)
	return m.Snapshot(), nil
}

func (m *Match) Resign(side Side) Snapshot {
	m.version++
	if side == White {
		m.resultText = "0-1"
		m.message = "White resigned"
	} else {
		m.resultText = "1-0"
		m.message = "Black resigned"
	}
	return m.SnapshotWithOverride("resigned", m.resultText, m.message)
}

func (m *Match) SnapshotWithOverride(status, result, message string) Snapshot {
	s := m.Snapshot()
	s.Status = status
	s.Result = result
	s.Message = message
	return s
}

type MoveOption struct {
	UCI       string `json:"uci"`
	Target    string `json:"target"`
	Promotion string `json:"promotion,omitempty"`
}

func LegalMovesForSquare(fen string, side Side, from string) ([]MoveOption, error) {
	g, err := gameFromFEN(fen)
	if err != nil {
		return nil, err
	}
	if SideFromColor(g.Position().Turn()) != side {
		return nil, nil
	}
	var options []MoveOption
	for _, move := range g.ValidMoves() {
		uci := chess.UCINotation{}.Encode(g.Position(), move)
		if len(uci) < 4 || uci[:2] != from {
			continue
		}
		option := MoveOption{UCI: uci, Target: uci[2:4]}
		if len(uci) == 5 {
			option.Promotion = string(uci[4])
		}
		options = append(options, option)
	}
	return options, nil
}

func PieceAt(fen, square string) (rune, Side, bool, error) {
	g, err := gameFromFEN(fen)
	if err != nil {
		return 0, White, false, err
	}
	sq, err := squareFromString(square)
	if err != nil {
		return 0, White, false, err
	}
	if sq == chess.NoSquare {
		return 0, White, false, fmt.Errorf("invalid square %q", square)
	}
	piece := g.Position().Board().Piece(sq)
	if piece == chess.NoPiece {
		return 0, White, false, nil
	}
	return pieceSymbol(piece), SideFromColor(piece.Color()), true, nil
}

func gameFromFEN(fen string) (*chess.Game, error) {
	opt, err := chess.FEN(fen)
	if err != nil {
		return nil, fmt.Errorf("parse fen: %w", err)
	}
	return chess.NewGame(chess.UseNotation(chess.UCINotation{}), opt), nil
}

func ParseSquareName(file, rank int) (string, error) {
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return "", errors.New("square out of bounds")
	}
	return string(rune('a'+file)) + string(rune('1'+rank)), nil
}

func squareFromString(square string) (chess.Square, error) {
	if len(square) != 2 {
		return chess.NoSquare, fmt.Errorf("invalid square %q", square)
	}
	file := square[0]
	rank := square[1]
	if file < 'a' || file > 'h' || rank < '1' || rank > '8' {
		return chess.NoSquare, fmt.Errorf("invalid square %q", square)
	}
	return chess.NewSquare(chess.File(file-'a'), chess.Rank(rank-'1')), nil
}

func pieceSymbol(piece chess.Piece) rune {
	symbol := map[chess.PieceType]rune{
		chess.Pawn:   'P',
		chess.Knight: 'N',
		chess.Bishop: 'B',
		chess.Rook:   'R',
		chess.Queen:  'Q',
		chess.King:   'K',
	}[piece.Type()]
	if piece.Color() == chess.Black {
		return rune(strings.ToLower(string(symbol))[0])
	}
	return symbol
}

func messageForState(g *chess.Game) string {
	if g.Outcome() != chess.NoOutcome {
		return statusString(g.Outcome(), g.Method())
	}
	return fmt.Sprintf("%s to move", strings.Title(string(SideFromColor(g.Position().Turn()))))
}

func statusString(outcome chess.Outcome, method chess.Method) string {
	if outcome == chess.NoOutcome {
		return "active"
	}
	return strings.ReplaceAll(strings.ToLower(method.String()), "_", " ")
}

func resultString(outcome chess.Outcome) string {
	if outcome == chess.NoOutcome {
		return ""
	}
	return outcome.String()
}
