package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"
)

const (
	DefaultPort     = 18787
	ServiceName     = "chess-wifi"
	ProtocolVersion = "1"
	KindAnnounce    = "announce"
	KindQuery       = "query"

	defaultInterval = 1 * time.Second
	defaultTimeout  = 1200 * time.Millisecond
)

type Announcement struct {
	Kind            string `json:"kind,omitempty"`
	Service         string `json:"service"`
	ProtocolVersion string `json:"protocol_version"`
	PlayerName      string `json:"player_name"`
	MatchPort       int    `json:"match_port"`
}

type Match struct {
	PlayerName string
	Address    string
	LastSeen   time.Time
}

type AnnounceOptions struct {
	DiscoveryPort int
	Destinations  []string
	Interval      time.Duration
}

type ScanOptions struct {
	ListenAddress string
	Timeout       time.Duration
	Now           func() time.Time
}

func StartAnnouncer(ctx context.Context, playerName string, matchPort int) (func(), error) {
	return StartAnnouncerWithOptions(ctx, playerName, matchPort, AnnounceOptions{})
}

func StartAnnouncerWithOptions(ctx context.Context, playerName string, matchPort int, opts AnnounceOptions) (func(), error) {
	if matchPort <= 0 || matchPort > 65535 {
		return nil, fmt.Errorf("invalid match port %d", matchPort)
	}
	port := opts.DiscoveryPort
	if port == 0 {
		port = DefaultPort
	}
	interval := opts.Interval
	if interval <= 0 {
		interval = defaultInterval
	}
	destinations := opts.Destinations
	if len(destinations) == 0 {
		destinations = defaultDestinations(port)
	}
	if playerName == "" {
		playerName = "Host"
	}

	conn, err := listenUDP(ctx, fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("start discovery announcer: %w", err)
	}
	_ = conn.SetWriteBuffer(2048)
	childCtx, cancel := context.WithCancel(ctx)
	announcement := Announcement{
		Kind:            KindAnnounce,
		Service:         ServiceName,
		ProtocolVersion: ProtocolVersion,
		PlayerName:      playerName,
		MatchPort:       matchPort,
	}
	payload, err := json.Marshal(announcement)
	if err != nil {
		_ = conn.Close()
		cancel()
		return nil, fmt.Errorf("encode discovery announcement: %w", err)
	}
	addresses, err := resolveDestinations(destinations)
	if err != nil {
		_ = conn.Close()
		cancel()
		return nil, err
	}

	go func() {
		defer conn.Close()
		sendAnnouncement(conn, addresses, payload)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		buf := make([]byte, 2048)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			n, remote, err := conn.ReadFromUDP(buf)
			if err == nil && isQuery(buf[:n]) {
				_, _ = conn.WriteToUDP(payload, remote)
			}
			select {
			case <-childCtx.Done():
				return
			case <-ticker.C:
				sendAnnouncement(conn, addresses, payload)
			default:
			}
		}
	}()

	return cancel, nil
}

func Scan(ctx context.Context) ([]Match, error) {
	return ScanWithOptions(ctx, ScanOptions{})
}

func ScanWithOptions(ctx context.Context, opts ScanOptions) ([]Match, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	listenAddress := opts.ListenAddress
	if listenAddress == "" {
		listenAddress = fmt.Sprintf(":%d", DefaultPort)
	}
	conn, err := listenUDP(ctx, listenAddress)
	if err != nil {
		return nil, fmt.Errorf("listen for discovery announcements: %w", err)
	}
	defer conn.Close()

	deadline := now().Add(timeout)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, fmt.Errorf("set discovery deadline: %w", err)
	}
	port := DefaultPort
	if _, rawPort, err := net.SplitHostPort(listenAddress); err == nil {
		if parsed, parseErr := strconv.Atoi(rawPort); parseErr == nil && parsed > 0 {
			port = parsed
		}
	}
	sendQuery(conn, defaultDestinations(port))
	matches := map[string]Match{}
	buf := make([]byte, 2048)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				return sortedMatches(matches), nil
			}
			select {
			case <-ctx.Done():
				return sortedMatches(matches), nil
			default:
				return sortedMatches(matches), fmt.Errorf("read discovery announcement: %w", err)
			}
		}
		match, ok := parseAnnouncement(buf[:n], remote, now())
		if ok {
			matches[match.Address] = match
		}
	}
}

func parseAnnouncement(payload []byte, remote *net.UDPAddr, seen time.Time) (Match, bool) {
	var announcement Announcement
	if err := json.Unmarshal(payload, &announcement); err != nil {
		return Match{}, false
	}
	if announcement.Kind == KindQuery {
		return Match{}, false
	}
	if announcement.Service != ServiceName || announcement.ProtocolVersion != ProtocolVersion || announcement.MatchPort <= 0 || announcement.MatchPort > 65535 {
		return Match{}, false
	}
	host := ""
	if remote != nil && remote.IP != nil {
		host = remote.IP.String()
	}
	if host == "" || host == "<nil>" {
		return Match{}, false
	}
	name := announcement.PlayerName
	if name == "" {
		name = "Host"
	}
	return Match{
		PlayerName: name,
		Address:    net.JoinHostPort(host, fmt.Sprintf("%d", announcement.MatchPort)),
		LastSeen:   seen,
	}, true
}

func isQuery(payload []byte) bool {
	var announcement Announcement
	if err := json.Unmarshal(payload, &announcement); err != nil {
		return false
	}
	return announcement.Kind == KindQuery && announcement.Service == ServiceName && announcement.ProtocolVersion == ProtocolVersion
}

func sendQuery(conn *net.UDPConn, destinations []string) {
	payload, err := json.Marshal(Announcement{
		Kind:            KindQuery,
		Service:         ServiceName,
		ProtocolVersion: ProtocolVersion,
	})
	if err != nil {
		return
	}
	addresses, err := resolveDestinations(destinations)
	if err != nil {
		return
	}
	sendAnnouncement(conn, addresses, payload)
}

func defaultDestinations(port int) []string {
	destinations := []string{
		fmt.Sprintf("255.255.255.255:%d", port),
	}
	for _, addr := range interfaceBroadcasts(port) {
		destinations = append(destinations, addr)
	}
	return destinations
}

func interfaceBroadcasts(port int) []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, raw := range addrs {
			ipNet, ok := raw.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || !ip.IsPrivate() {
				continue
			}
			mask := ipNet.Mask
			if len(mask) != net.IPv4len {
				continue
			}
			broadcast := net.IPv4(
				ip[0]|^mask[0],
				ip[1]|^mask[1],
				ip[2]|^mask[2],
				ip[3]|^mask[3],
			)
			out = append(out, fmt.Sprintf("%s:%d", broadcast.String(), port))
		}
	}
	sort.Strings(out)
	return out
}

func resolveDestinations(destinations []string) ([]net.Addr, error) {
	var addresses []net.Addr
	for _, destination := range destinations {
		addr, err := net.ResolveUDPAddr("udp4", destination)
		if err != nil {
			return nil, fmt.Errorf("resolve discovery destination %q: %w", destination, err)
		}
		addresses = append(addresses, addr)
	}
	if len(addresses) == 0 {
		return nil, errors.New("no discovery destinations")
	}
	return addresses, nil
}

func sendAnnouncement(conn net.PacketConn, addresses []net.Addr, payload []byte) {
	for _, addr := range addresses {
		_, _ = conn.WriteTo(payload, addr)
	}
}

func sortedMatches(matches map[string]Match) []Match {
	out := make([]Match, 0, len(matches))
	for _, match := range matches {
		out = append(out, match)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PlayerName == out[j].PlayerName {
			return out[i].Address < out[j].Address
		}
		return out[i].PlayerName < out[j].PlayerName
	})
	return out
}
