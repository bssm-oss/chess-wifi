package lan

import (
	"fmt"
	"net"
	"sort"
)

func PrivateIPv4(port int) ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}
	var addrs []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, raw := range ifaceAddrs {
			ipNet, ok := raw.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil || !ip.IsPrivate() {
				continue
			}
			addrs = append(addrs, fmt.Sprintf("%s:%d", ip.String(), port))
		}
	}
	sort.Strings(addrs)
	if len(addrs) == 0 {
		addrs = append(addrs, fmt.Sprintf("127.0.0.1:%d", port))
	}
	return addrs, nil
}
