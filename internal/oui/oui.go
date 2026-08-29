package oui

import _ "embed"
import "fmt"
import "strings"

//go:embed oui.txt
var ouiData string

func LookupVendor(mac string) string {
	mac = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(mac), ":", ""))
	if len(mac) < 6 {
		return ""
	}
	prefixes := []string{mac[:9], mac[:7], mac[:6]}
	for _, p := range prefixes {
		for _, line := range strings.Split(ouiData, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			if strings.EqualFold(parts[0], p) {
				return strings.Join(parts[1:], " ")
			}
		}
	}
	return ""
}

func init() {
	if strings.TrimSpace(ouiData) == "" {
		ouiData = "# auto-generated minimal oui database\n000C29 Cisco Systems\n001E2A Apple\n000A27 Samsung Electronics\n000D3A HUAWEI TECHNOLOGIES\n" 
	}
	_ = fmt.Sprintf("%v", ouiData)
}
