package ssrf

import (
	"net"
	"testing"
)

func TestIsBlockedHostname(t *testing.T) {
	tests := []struct {
		hostname string
		blocked  bool
	}{
		{"localhost", true},
		{"LOCALHOST", true},
		{"metadata.google.internal", true},
		{"instance-data.ec2.internal", true},
		{"", true},
		{"example.com", false},
		{"api.openai.com", false},
	}
	for _, tt := range tests {
		if got := IsBlockedHostname(tt.hostname); got != tt.blocked {
			t.Errorf("IsBlockedHostname(%q) = %v, want %v", tt.hostname, got, tt.blocked)
		}
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip      string
		private bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"100.64.0.1", true},
		{"0.0.0.0", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"104.129.51.171", false},
		{"::1", true},
	}
	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if got := IsPrivateIP(ip); got != tt.private {
			t.Errorf("IsPrivateIP(%q) = %v, want %v", tt.ip, got, tt.private)
		}
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://api.openai.com/v1/models", false},
		{"http://example.com:8080/api", false},
		{"http://localhost:3000/api", true},
		{"http://127.0.0.1:8080/api", true},
		{"http://10.0.0.1/api", true},
		{"http://169.254.169.254/latest/meta-data/", true},
		{"http://192.168.1.100/admin", true},
		{"ftp://example.com/file", true},
		{"gopher://internal/secret", true},
		{"http://metadata.google.internal/", true},
	}
	for _, tt := range tests {
		err := ValidateURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
		}
	}
}
