package domain

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

var hostnameRegex = regexp.MustCompile(`^([a-zA-Z0-9_]|[a-zA-Z0-9_][a-zA-Z0-9\-_]*[a-zA-Z0-9_])(\.([a-zA-Z0-9_]|[a-zA-Z0-9_][a-zA-Z0-9\-_]*[a-zA-Z0-9_]))*$`)

// ValidateIP checks whether ipStr is a valid IPv4 or IPv6 address.
func ValidateIP(ipStr string) error {
	trimmed := strings.TrimSpace(ipStr)
	if trimmed == "" {
		return fmt.Errorf("IP address cannot be empty")
	}
	ip := net.ParseIP(trimmed)
	if ip == nil {
		return fmt.Errorf("invalid IP address format: %q", ipStr)
	}
	return nil
}

// ValidateHostname checks whether hostname is a valid hostname according to standard RFC rules.
func ValidateHostname(hostname string) error {
	trimmed := strings.TrimSpace(hostname)
	if trimmed == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if len(trimmed) > 253 {
		return fmt.Errorf("hostname exceeds maximum length of 253 characters")
	}
	if !hostnameRegex.MatchString(trimmed) {
		return fmt.Errorf("invalid hostname format: %q", hostname)
	}
	return nil
}

// ValidateAliases checks a slice of alias strings.
func ValidateAliases(aliases []string) error {
	for _, alias := range aliases {
		trimmed := strings.TrimSpace(alias)
		if trimmed == "" {
			continue
		}
		if err := ValidateHostname(trimmed); err != nil {
			return fmt.Errorf("invalid alias %q: %w", alias, err)
		}
	}
	return nil
}

// ValidateEntry validates IP, hostname, and aliases of an entry.
func ValidateEntry(entry *HostEntry) error {
	if err := ValidateIP(entry.IP); err != nil {
		return err
	}
	if err := ValidateHostname(entry.Hostname); err != nil {
		return err
	}
	if err := ValidateAliases(entry.Aliases); err != nil {
		return err
	}
	return nil
}
