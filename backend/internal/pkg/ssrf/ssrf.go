// Package ssrf provides shared SSRF protection utilities.
// It validates URLs and provides safe HTTP transport that blocks requests
// to private/internal networks, cloud metadata endpoints, and loopback addresses.
package ssrf

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BlockedHostnames contains known cloud metadata and local hostnames that must be blocked.
var BlockedHostnames = map[string]struct{}{
	"localhost":                  {},
	"localhost.localdomain":      {},
	"metadata":                   {},
	"metadata.google.internal":   {},
	"metadata.goog":              {},
	"instance-data":              {},
	"instance-data.ec2.internal": {},
}

// BlockedCIDRs contains all private/internal CIDR ranges.
var BlockedCIDRs = mustParseCIDRs([]string{
	"127.0.0.0/8",    // IPv4 loopback
	"10.0.0.0/8",     // RFC1918
	"172.16.0.0/12",  // RFC1918
	"192.168.0.0/16", // RFC1918
	"169.254.0.0/16", // link-local (includes cloud metadata 169.254.169.254)
	"100.64.0.0/10",  // CGNAT
	"0.0.0.0/8",      // "this network"
	"::1/128",        // IPv6 loopback
	"fc00::/7",       // IPv6 ULA
	"fe80::/10",      // IPv6 link-local
	"::/128",         // IPv6 unspecified
})

var safeDialer = &net.Dialer{
	Timeout:   10 * time.Second,
	KeepAlive: 30 * time.Second,
}

func mustParseCIDRs(cidrs []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			panic("ssrf: invalid CIDR " + c + ": " + err.Error())
		}
		out = append(out, n)
	}
	return out
}

// IsBlockedHostname checks if a hostname is in the block list.
func IsBlockedHostname(hostname string) bool {
	if hostname == "" {
		return true
	}
	_, blocked := BlockedHostnames[strings.ToLower(hostname)]
	return blocked
}

// IsPrivateIP checks if an IP falls within blocked CIDR ranges.
func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsUnspecified() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() {
		return true
	}
	for _, n := range BlockedCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateURL checks if a URL is safe to request (not targeting internal networks).
// It performs DNS resolution to detect DNS rebinding attempts.
func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http/https schemes
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("blocked scheme: %s (only http/https allowed)", u.Scheme)
	}

	hostname := u.Hostname()
	if IsBlockedHostname(hostname) {
		return fmt.Errorf("blocked hostname: %s", hostname)
	}

	// Check if it's a direct IP
	if ip := net.ParseIP(hostname); ip != nil {
		if IsPrivateIP(ip) {
			return fmt.Errorf("blocked private IP: %s", hostname)
		}
		return nil
	}

	// Resolve hostname and check all IPs
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return fmt.Errorf("DNS resolution failed for %s: %w", hostname, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("no DNS records for %s", hostname)
	}
	for _, a := range addrs {
		if IsPrivateIP(a.IP) {
			return fmt.Errorf("hostname %s resolves to private IP %s", hostname, a.IP)
		}
	}
	return nil
}

// SafeDialContext dials the target address after verifying the resolved IP is not private.
// This prevents DNS rebinding attacks where DNS changes between validation and connection.
func SafeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	// Direct IP literal - fast path
	if ip := net.ParseIP(host); ip != nil {
		if IsPrivateIP(ip) {
			return nil, &net.AddrError{Err: "blocked by SSRF policy", Addr: address}
		}
		return safeDialer.DialContext(ctx, network, address)
	}
	if IsBlockedHostname(host) {
		return nil, &net.AddrError{Err: "blocked by SSRF policy", Addr: address}
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, &net.AddrError{Err: "no addresses for host", Addr: host}
	}
	var lastErr error
	for _, a := range addrs {
		if IsPrivateIP(a.IP) {
			lastErr = &net.AddrError{Err: "blocked by SSRF policy", Addr: a.IP.String()}
			continue
		}
		conn, dialErr := safeDialer.DialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr == nil {
		lastErr = &net.AddrError{Err: "no usable addresses", Addr: host}
	}
	return nil, lastErr
}

// NewSafeHTTPClient creates an http.Client with SSRF-safe transport.
// The transport uses SafeDialContext to prevent connections to internal networks.
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		DialContext:           SafeDialContext,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
