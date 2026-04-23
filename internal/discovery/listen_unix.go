//go:build !windows

package discovery

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func listenUDP(ctx context.Context, address string) (*net.UDPConn, error) {
	listenConfig := net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var controlErr error
			if err := conn.Control(func(fd uintptr) {
				if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
					controlErr = err
					return
				}
				if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
					controlErr = err
				}
			}); err != nil {
				return err
			}
			return controlErr
		},
	}
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
