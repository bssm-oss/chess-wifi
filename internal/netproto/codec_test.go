package netproto

import (
	"bytes"
	"testing"

	"github.com/bssm-oss/chess-wifi/internal/game"
)

func TestCodecRoundTrip(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	codec := NewCodec(buf, buf)
	want := Envelope{
		Type: TypeSnapshot,
		Payload: Snapshot{State: game.Snapshot{
			Version:     2,
			FEN:         "8/8/8/8/8/8/4P3/4K3 w - - 0 1",
			Turn:        game.White,
			MoveHistory: []string{"e2e4"},
		}},
	}
	if err := codec.Write(want); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	got, err := codec.Read()
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if got.Type != want.Type {
		t.Fatalf("expected type %s, got %s", want.Type, got.Type)
	}
	payload, err := DecodePayload[Snapshot](got)
	if err != nil {
		t.Fatalf("DecodePayload returned error: %v", err)
	}
	if payload.State.Version != 2 || payload.State.FEN == "" {
		t.Fatalf("unexpected payload: %+v", payload.State)
	}
}
