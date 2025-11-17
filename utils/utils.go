package utils

import (
	"net"
	"regexp"
)

func CheckIP(data string) bool {
	ip := net.ParseIP(data)
	return ip != nil
}

func CheckDNS(data string) bool {
	if CheckIP(data) {
		return false
	}

	dnsPattern := `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(dnsPattern, data)
	return matched
}

func GetIPFromDNS(data string) ([]string, error) {
	ips, err := net.LookupIP(data)
	if err != nil {
		return nil, err
	}
	var ipv4s []string
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			ipv4s = append(ipv4s, ipv4.String())
		}
	}
	return ipv4s, nil
}
