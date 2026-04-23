//go:build windows

package discovery

import (
	"context"
	"fmt"
	"net"
)

func listenUDP(ctx context.Context, address string) (*net.UDPConn, error) {
	listenConfig := net.ListenConfig{}
	packetConn, err := listenConfig.ListenPacket(ctx, "udp4", address)
	if err != nil {
		return nil, err
	}
	udpConn, ok := packetConn.(*net.UDPConn)
	if !ok {
		_ = packetConn.Close()
		return nil, fmt.Errorf("expected UDP connection, got %T", packetConn)
	}
	return udpConn, nil
}
