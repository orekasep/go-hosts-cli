package domain

import (
	"testing"
)

func TestValidateIP(t *testing.T) {
	validIPs := []string{
		"127.0.0.1",
		"192.168.1.1",
		"10.0.0.1",
		"::1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	}
	for _, ip := range validIPs {
		if err := ValidateIP(ip); err != nil {
			t.Errorf("expected valid IP %q, got error: %v", ip, err)
		}
	}

	invalidIPs := []string{
		"",
		"999.999.999.999",
		"not-an-ip",
		"192.168.1",
	}
	for _, ip := range invalidIPs {
		if err := ValidateIP(ip); err == nil {
			t.Errorf("expected invalid IP %q, got nil error", ip)
		}
	}
}

func TestValidateHostname(t *testing.T) {
	validHosts := []string{
		"localhost",
		"api.local",
		"staging.api.service.io",
		"my-host-1",
		"host_with_underscore",
	}
	for _, h := range validHosts {
		if err := ValidateHostname(h); err != nil {
			t.Errorf("expected valid hostname %q, got error: %v", h, err)
		}
	}

	invalidHosts := []string{
		"",
		"-startwithdash",
		"endwithdash-",
		"has space",
		"has!special@char",
	}
	for _, h := range invalidHosts {
		if err := ValidateHostname(h); err == nil {
			t.Errorf("expected invalid hostname %q, got nil error", h)
		}
	}
}
