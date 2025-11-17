package utils

import (
	"testing"
)

func TestCheckIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv4 localhost", "127.0.0.1", true},
		{"Valid IPv4 zero", "0.0.0.0", true},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"Valid IPv6 short", "::1", true},
		{"Valid IPv6 compressed", "2001:db8::1", true},
		{"Invalid IP - letters", "192.168.1.a", false},
		{"Invalid IP - out of range", "256.1.1.1", false},
		{"Invalid IP - incomplete", "192.168.1", false},
		{"Empty string", "", false},
		{"DNS name", "example.com", false},
		{"Invalid format", "not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckIP(tt.input)
			if result != tt.expected {
				t.Errorf("CheckIP(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckDNS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid DNS", "example.com", true},
		{"Valid DNS with subdomain", "www.example.com", true},
		{"Valid DNS multiple subdomains", "api.v1.example.com", true},
		{"Valid DNS with hyphen", "my-site.example.com", true},
		{"Valid DNS with numbers", "site123.example.com", true},
		{"Valid DNS long TLD", "example.technology", true},
		{"Invalid DNS - IP address", "192.168.1.1", false},
		{"Invalid DNS - IPv6", "::1", false},
		{"Invalid DNS - starts with hyphen", "-example.com", false},
		{"Invalid DNS - ends with hyphen", "example-.com", false},
		{"Invalid DNS - no TLD", "example", false},
		{"Invalid DNS - single char TLD", "example.c", false},
		{"Invalid DNS - starts with dot", ".example.com", false},
		{"Invalid DNS - ends with dot", "example.com.", false},
		{"Invalid DNS - double dot", "example..com", false},
		{"Invalid DNS - special chars", "example!.com", false},
		{"Invalid DNS - underscore", "example_site.com", false},
		{"Invalid DNS - space", "example .com", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckDNS(tt.input)
			if result != tt.expected {
				t.Errorf("CheckDNS(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetIPFromDNS(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		minIPs    int // minimum number of IPs expected
	}{
		{"Valid DNS - localhost", "localhost", false, 1},
		{"Valid DNS - google.com", "google.com", false, 1},
		{"Invalid DNS - nonexistent", "this-domain-does-not-exist-12345.com", true, 0},
		{"Invalid DNS - empty string", "", true, 0},
		{"Invalid DNS - invalid format", "invalid..domain", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetIPFromDNS(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("GetIPFromDNS(%q) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("GetIPFromDNS(%q) unexpected error: %v", tt.input, err)
				}
				if len(result) < tt.minIPs {
					t.Errorf("GetIPFromDNS(%q) returned %d IPs, expected at least %d", tt.input, len(result), tt.minIPs)
				}
				// Verify all returned values are valid IPv4 addresses
				for _, ip := range result {
					if !CheckIP(ip) {
						t.Errorf("GetIPFromDNS(%q) returned invalid IP: %s", tt.input, ip)
					}
				}
			}
		})
	}
}
