package scan

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"huzbackend-go/internal/oui"
)

type Device struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
	Iface    string `json:"iface"`
	State    string `json:"state,omitempty"`
	Vendor   string `json:"vendor,omitempty"`
}

type Scanner struct{}

func NewScanner() *Scanner { return &Scanner{} }

func (s *Scanner) NetworkDevices() (map[string]any, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var candidates []net.IPNet
	for _, i := range ifs {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet == nil || ipnet.IP == nil || ipnet.IP.To4() == nil || ipnet.Mask == nil {
				continue
			}
			candidates = append(candidates, *ipnet)
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no network interface")
	}

	_ = getHostRange(candidates[0])
	arpEntries, err := readARPTable()
	if err != nil {
		return nil, err
	}
	result := make([]Device, 0, len(arpEntries))
	for _, ent := range arpEntries {
		if ent.MAC == "" || ent.MAC == "ff:ff:ff:ff:ff:ff" || isMulticastIP(ent.IP) {
			continue
		}
		ent.Vendor = oui.LookupVendor(ent.MAC)
		if ent.Hostname == "" {
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancel()
			hosts, err := net.DefaultResolver.LookupAddr(ctx, ent.IP)
			if err == nil && len(hosts) > 0 {
				ent.Hostname = hosts[0]
			}
		}
		result = append(result, ent)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].IP < result[j].IP })
	return map[string]any{"count": len(result), "devices": result}, nil
}

func getHostRange(ipnet net.IPNet) []string {
	ip := ipnet.IP.To4()
	mask := ipnet.Mask
	if ip == nil || mask == nil {
		return nil
	}
	if len(mask) != net.IPv4len || len(ip) != net.IPv4len {
		return nil
	}
	network := make(net.IP, net.IPv4len)
	broadcast := make(net.IP, net.IPv4len)
	for i := 0; i < net.IPv4len; i++ {
		network[i] = ip[i] & mask[i]
		broadcast[i] = ip[i] | ^mask[i]
	}
	var out []string
	start := ipToInt(network) + 1
	end := ipToInt(broadcast) - 1
	if start >= end {
		return out
	}
	for v := start; v < end && len(out) < 1024; v++ {
		out = append(out, intToIPv4(v).String())
	}
	return out
}

func ipToInt(ip net.IP) uint32 {
	v := ip.To4()
	if v == nil { return 0 }
	return uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
}

func intToIPv4(v uint32) net.IP {
	return net.IPv4(byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func pingHost(ip string) bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "500", ip)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	}
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func readARPTable() ([]Device, error) {
	out, err := exec.Command("arp", "-a").CombinedOutput()
	if err != nil && runtime.GOOS == "linux" {
		out, err = exec.Command("ip", "neighbor", "show").CombinedOutput()
	}
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	result := make([]Device, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		item, ok := parseARPLine(line)
		if ok {
			result = append(result, item)
		}
	}
	return result, nil
}

func parseARPLine(raw string) (Device, bool) {
	line := strings.TrimSpace(raw)
	if strings.Contains(line, "FAILED") || strings.Contains(line, "incomplete") {
		return Device{}, false
	}
	if strings.Contains(line, "lladdr") || strings.Contains(line, "dev") {
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			for i := 0; i < len(parts)-3; i++ {
				if net.ParseIP(parts[i]) != nil && i+3 < len(parts) && strings.EqualFold(parts[i+1], "dev") {
					mac := strings.TrimSpace(parts[i+3])
					if isValidMAC(mac) {
						return Device{IP: parts[i], MAC: normalizeMAC(mac), Iface: parts[i+2], State: "reachable"}, true
					}
				}
			}
		}
	}
	if strings.Contains(line, "(") && strings.Contains(line, ")") {
		m := arpRegex.FindStringSubmatch(line)
		if len(m) >= 5 {
			ip := strings.TrimSpace(m[2])
			mac := strings.TrimSpace(m[3])
			if isValidMAC(mac) {
				return Device{IP: ip, MAC: normalizeMAC(mac), Hostname: strings.TrimSpace(m[1]), Iface: strings.TrimSpace(m[4])}, true
			}
		}
	}
	return Device{}, false
}

func isValidMAC(mac string) bool {
	mac = strings.TrimSpace(mac)
	if mac == "" || strings.Contains(strings.ToLower(mac), "incomplete") || strings.Contains(strings.ToLower(mac), "ff:ff:ff:ff:ff:ff") {
		return false
	}
	_, err := net.ParseMAC(mac)
	return err == nil
}

func normalizeMAC(mac string) string {
	m, err := net.ParseMAC(mac)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(mac))
	}
	return strings.ToLower(m.String())
}

func isMulticastIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	v4 := parsed.To4()
	if v4 == nil {
		return false
	}
	return v4[0] >= 224 && v4[0] <= 239
}

var arpRegex = regexp.MustCompile(`^(\S+)?\s*\((\d+\.\d+\.\d+\.\d+)\)\s+at\s+([0-9a-fA-F:]{11,17})(?:.*on\s+(\S+))?`)
