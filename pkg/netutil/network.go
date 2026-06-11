package netutil

import (
	"net"
	"strings"
)

// isLinkLocal checks if an IP is in the 169.254.0.0/16 range (link-local)
func isLinkLocal(ip net.IP) bool {
	return ip[0] == 169 && ip[1] == 254
}

// isVirtualInterface checks if an interface name looks like a VPN/virtual adapter
func isVirtualInterface(name string) bool {
	lower := strings.ToLower(name)
	keywords := []string{"vpn", "tun", "tap", "veth", "docker", "vmnet", "vbox", "virbr", "wg", "tailscale", "zerotier"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// GetLocalIP returns the preferred local IP address (filters out link-local and VPN interfaces)
func GetLocalIP() string {
	ips := GetAllLocalIPs()
	if len(ips) > 0 {
		return ips[0]
	}
	return "127.0.0.1"
}

// GetAllLocalIPs returns all local IP addresses, prioritizing real LAN IPs over virtual/link-local ones.
func GetAllLocalIPs() []string {
	var preferred []string
	var fallback []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"127.0.0.1"}
	}

	for _, iface := range ifaces {
		// Skip down or loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		isVirtual := isVirtualInterface(iface.Name)

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}

			// Skip link-local addresses (169.254.x.x)
			if isLinkLocal(ip) {
				continue
			}

			if isVirtual {
				fallback = append(fallback, ip.String())
			} else {
				preferred = append(preferred, ip.String())
			}
		}
	}

	// Return preferred (real LAN) first, then fallback (VPN/virtual)
	result := append(preferred, fallback...)
	if len(result) == 0 {
		return []string{"127.0.0.1"}
	}
	return result
}
