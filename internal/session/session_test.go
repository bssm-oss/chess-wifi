package session

import (
	"context"
	"testing"
	"time"
)

func TestHostJoinAndMoveSync(t *testing.T) {
	hostListener, err := StartHost("Host", 9101)
	if err != nil {
		t.Fatalf("StartHost returned error: %v", err)
	}
	t.Cleanup(func() { _ = hostListener.Close() })

	acceptedCh := make(chan *PeerSession, 1)
	errCh := make(chan error, 1)
	go func() {
		select {
		case peer := <-hostListener.Accepted():
			acceptedCh <- peer
		case err := <-hostListener.Errors():
			errCh <- err
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	guest, err := Join(ctx, "127.0.0.1:9101", "Guest")
	if err != nil {
		t.Fatalf("Join returned error: %v", err)
	}
	t.Cleanup(func() { _ = guest.Close() })

	var host *PeerSession
	select {
	case host = <-acceptedCh:
	case err := <-errCh:
		t.Fatalf("host listener returned error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for host acceptance")
	}
	t.Cleanup(func() { _ = host.Close() })

	if host.Role() != "white" || guest.Role() != "black" {
		t.Fatalf("unexpected roles: host=%s guest=%s", host.Role(), guest.Role())
	}
	if err := host.SubmitMove("e2e4"); err != nil {
		t.Fatalf("host move failed: %v", err)
	}
	select {
	case event := <-guest.Events():
		if event.Type != EventSnapshot {
			t.Fatalf("expected snapshot event, got %s", event.Type)
		}
		if event.Snapshot.LastMoveUCI != "e2e4" {
			t.Fatalf("expected synced move e2e4, got %q", event.Snapshot.LastMoveUCI)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for guest snapshot")
	}
}
