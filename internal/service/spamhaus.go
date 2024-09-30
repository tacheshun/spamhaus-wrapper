package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type SpamhausService struct {
	resolver *net.Resolver
}

func NewSpamhausService() *SpamhausService {
	return &SpamhausService{
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", "1.1.1.1:53")
			},
		},
	}
}

func (s *SpamhausService) LookupIP(ctx context.Context, ip string) (string, error) {
	reversedIP := reverseIP(ip)
	domain := fmt.Sprintf("%s.zen.spamhaus.org", reversedIP)

	var responseCode string
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		responseCode, err = s.lookup(ctx, domain)
		if err == nil && responseCode != "127.255.255.254" {
			return responseCode, nil
		}
		if err != nil && !strings.Contains(err.Error(), "NXDOMAIN") {
			return "", err
		}
		// Exponential backoff
		time.Sleep(time.Duration(1<<attempt) * time.Second)
	}

	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("received non-determinative response after retries")
}

func (s *SpamhausService) lookup(ctx context.Context, domain string) (string, error) {
	ips, err := s.resolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return "", nil // NXDOMAIN, not registered
		}
		return "", err
	}
	if len(ips) > 0 {
		return ips[0].String(), nil
	}
	return "", fmt.Errorf("no IP addresses returned")
}

func reverseIP(ip string) string {
	parts := strings.Split(ip, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}
