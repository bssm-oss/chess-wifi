package discovery

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestAnnouncerAndScan(t *testing.T) {
	port := freeUDPPort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop, err := StartAnnouncerWithOptions(ctx, "Host", 8787, AnnounceOptions{
		DiscoveryPort: port,
		Destinations:  []string{fmt.Sprintf("127.0.0.1:%d", port)},
		Interval:      25 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("StartAnnouncerWithOptions returned error: %v", err)
	}
	t.Cleanup(stop)

	matches, err := ScanWithOptions(t.Context(), ScanOptions{
		ListenAddress: fmt.Sprintf("127.0.0.1:%d", port),
		Timeout:       300 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("ScanWithOptions returned error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 discovered match, got %d: %+v", len(matches), matches)
	}
	if matches[0].PlayerName != "Host" {
		t.Fatalf("expected Host player, got %q", matches[0].PlayerName)
	}
	if matches[0].Address != "127.0.0.1:8787" {
		t.Fatalf("expected 127.0.0.1:8787, got %q", matches[0].Address)
	}
}

func TestScanIgnoresInvalidAnnouncements(t *testing.T) {
	port := freeUDPPort(t)
	conn, err := net.Dial("udp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("dial UDP returned error: %v", err)
	}
	defer conn.Close()

	done := make(chan []Match, 1)
	errCh := make(chan error, 1)
	go func() {
		matches, err := ScanWithOptions(context.Background(), ScanOptions{
			ListenAddress: fmt.Sprintf("127.0.0.1:%d", port),
			Timeout:       150 * time.Millisecond,
		})
		if err != nil {
			errCh <- err
			return
		}
		done <- matches
	}()

	time.Sleep(25 * time.Millisecond)
	if _, err := conn.Write([]byte(`{"service":"other","protocol_version":"1","match_port":8787}`)); err != nil {
		t.Fatalf("write invalid announcement returned error: %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("ScanWithOptions returned error: %v", err)
	case matches := <-done:
		if len(matches) != 0 {
			t.Fatalf("expected invalid announcement to be ignored, got %+v", matches)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for scan")
	}
}

func TestQueryPayloadIsNotParsedAsMatch(t *testing.T) {
	query := []byte(`{"kind":"query","service":"chess-wifi","protocol_version":"1"}`)
	if match, ok := parseAnnouncement(query, &net.UDPAddr{IP: net.ParseIP("192.168.0.2")}, time.Now()); ok {
		t.Fatalf("expected query to be ignored, got %+v", match)
	}
	if !isQuery(query) {
		t.Fatal("expected query payload to be recognized")
	}
}

func freeUDPPort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ResolveUDPAddr returned error: %v", err)
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		t.Fatalf("ListenUDP returned error: %v", err)
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).Port
}
