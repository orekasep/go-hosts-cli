package migrate

import (
	"bufio"
	"bytes"
	"net"
	"regexp"
	"strings"

	"github.com/orekasep/go-hosts-cli/internal/core/etchosts"
	"github.com/orekasep/go-hosts-cli/internal/domain"
)

var (
	groupHeaderRegex = regexp.MustCompile(`^#+\s*\[?([a-zA-Z0-9_\-/\.\s]+)\]?\s*$`)
	whitespaceRegex  = regexp.MustCompile(`\s+`)
)

// ParseHostsFile parses the content of a hosts file into domain.HostsFile.
func ParseHostsFile(content []byte) (*domain.HostsFile, error) {
	// Strip any existing hostcli markers so we only parse underlying entries
	preamble, _, suffix, err := etchosts.ExtractManagedBlock(content)
	if err == nil && (len(preamble) > 0 || len(suffix) > 0) {
		var combined []byte
		combined = append(combined, preamble...)
		combined = append(combined, suffix...)
		content = combined
	}

	hf := &domain.HostsFile{
		Version: 1,
		Groups:  []domain.Group{},
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	currentGroup := "default"

	for scanner.Scan() {
		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)

		if trimmed == "" {
			continue
		}

		// Check if line is a group header comment (e.g. "# [work/dev]" or "# Docker Services")
		if strings.HasPrefix(trimmed, "#") {
			// Check if this commented line is actually a disabled host entry
			disabledContent := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if entry, ok := parseEntryLine(disabledContent, false); ok {
				grp := hf.GetOrCreateGroup(currentGroup)
				grp.Entries = append(grp.Entries, entry)
				continue
			}

			// Check for group name pattern
			matches := groupHeaderRegex.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				headerText := strings.TrimSpace(matches[1])
				// Exclude generic comments or copyright notices
				if !strings.Contains(strings.ToLower(headerText), "copyright") &&
					!strings.Contains(strings.ToLower(headerText), "hostcli") &&
					!strings.Contains(strings.ToLower(headerText), "generated") &&
					len(headerText) < 50 {
					currentGroup = strings.ToLower(headerText)
				}
			}
			continue
		}

		// Active line
		if entry, ok := parseEntryLine(trimmed, true); ok {
			group := currentGroup
			// Auto-categorize standard loopback to "system" group
			if isSystemHost(entry.IP, entry.Hostname) && (currentGroup == "default" || strings.Contains(currentGroup, "system") || strings.Contains(currentGroup, "loopback")) {
				group = "system"
			}
			grp := hf.GetOrCreateGroup(group)
			grp.Entries = append(grp.Entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hf, nil
}

func parseEntryLine(line string, enabled bool) (domain.HostEntry, bool) {
	// Separate inline comment
	comment := ""
	if idx := strings.Index(line, "#"); idx != -1 {
		comment = strings.TrimSpace(line[idx+1:])
		line = strings.TrimSpace(line[:idx])
	}

	fields := whitespaceRegex.Split(line, -1)
	if len(fields) < 2 {
		return domain.HostEntry{}, false
	}

	ipStr := fields[0]
	if net.ParseIP(ipStr) == nil {
		return domain.HostEntry{}, false
	}

	hostname := fields[1]
	if domain.ValidateHostname(hostname) != nil {
		return domain.HostEntry{}, false
	}

	var aliases []string
	for _, f := range fields[2:] {
		if strings.TrimSpace(f) != "" && domain.ValidateHostname(f) == nil {
			aliases = append(aliases, f)
		}
	}

	return domain.HostEntry{
		ID:       domain.NewULID(),
		IP:       ipStr,
		Hostname: hostname,
		Aliases:  aliases,
		Enabled:  enabled,
		Comment:  comment,
	}, true
}

func isSystemHost(ip string, hostname string) bool {
	h := strings.ToLower(hostname)
	if h == "localhost" || h == "broadcasthost" || h == "local" || h == "ip6-localhost" || h == "ip6-loopback" {
		return true
	}
	if ip == "127.0.0.1" || ip == "::1" || ip == "fe00::0" || ip == "ff00::0" || ip == "ff02::1" || ip == "ff02::2" {
		return true
	}
	return false
}
